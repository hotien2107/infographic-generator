package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"infographic-generator/backend/internal/config"
	"infographic-generator/backend/internal/modules/documents"
	"infographic-generator/backend/internal/modules/projects"
	"infographic-generator/backend/internal/storage"
	"infographic-generator/backend/internal/utils"
)

type projectService interface {
	CreateProject(ctx context.Context, title string, inputMode projects.InputMode) (projects.Project, error)
	GetProject(ctx context.Context, projectID string) (projects.Project, []documents.Document, error)
	UploadDocument(ctx context.Context, projectID, originalFilename string, fileHeader *multipart.FileHeader) (projects.Project, documents.Document, error)
	TriggerProcessing(ctx context.Context, projectID string) (projects.Project, documents.Document, error)
}

type App struct {
	config       config.Config
	projectStore projects.Store
	storage      storage.BlobStorage
	service      projectService
}

type Meta struct {
	RequestID string    `json:"request_id"`
	Timestamp time.Time `json:"timestamp"`
}

type ErrorDetail struct {
	Code    string  `json:"code"`
	Message string  `json:"message"`
	Field   *string `json:"field"`
}

type createProjectRequest struct {
	Title     string `json:"title"`
	InputMode string `json:"input_mode"`
}

func New(cfg config.Config, store projects.Store, blobStorage storage.BlobStorage, service projectService) *App {
	return &App{
		config:       cfg,
		projectStore: store,
		storage:      blobStorage,
		service:      service,
	}
}

func (a *App) Close() {
	if a.projectStore != nil {
		a.projectStore.Close()
	}
	if a.storage != nil {
		_ = a.storage.Close()
	}
}

func (a *App) Handler() http.Handler {
	return http.HandlerFunc(a.serveHTTP)
}

func (a *App) serveHTTP(w http.ResponseWriter, r *http.Request) {
	a.applyCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if r.Method == http.MethodGet && r.URL.Path == "/healthz" {
		a.writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
		return
	}

	if r.Method == http.MethodPost && r.URL.Path == "/api/v1/projects" {
		a.createProject(w, r)
		return
	}

	if strings.HasPrefix(r.URL.Path, "/api/v1/projects/") {
		tail := strings.TrimPrefix(r.URL.Path, "/api/v1/projects/")
		parts := strings.Split(strings.Trim(tail, "/"), "/")
		if len(parts) == 1 && r.Method == http.MethodGet {
			a.getProject(w, r, parts[0])
			return
		}
		if len(parts) == 2 && parts[1] == "documents" && r.Method == http.MethodPost {
			a.uploadDocument(w, r, parts[0])
			return
		}
		if len(parts) == 2 && parts[1] == "processing" && r.Method == http.MethodPost {
			a.triggerProcessing(w, r, parts[0])
			return
		}
	}

	a.writeJSON(w, http.StatusNotFound, map[string]any{
		"data":  nil,
		"error": ErrorDetail{Code: "ROUTE_NOT_FOUND", Message: "route not found", Field: nil},
		"meta":  meta(utils.NewUUID()),
	})
}

func (a *App) createProject(w http.ResponseWriter, r *http.Request) {
	requestID := utils.NewUUID()
	defer r.Body.Close()

	var payload createProjectRequest
	if err := decodeStrictJSON(r, &payload); err != nil {
		field := "body"
		a.writeError(w, http.StatusBadRequest, requestID, "VALIDATION_ERROR", err.Error(), &field)
		return
	}

	title := strings.TrimSpace(payload.Title)
	if len(title) < 3 || len(title) > 120 {
		field := "title"
		a.writeError(w, http.StatusBadRequest, requestID, "VALIDATION_ERROR", "title must be between 3 and 120 characters", &field)
		return
	}

	inputMode := projects.InputMode(payload.InputMode)
	if inputMode != projects.InputModeFile && inputMode != projects.InputModeText {
		field := "input_mode"
		a.writeError(w, http.StatusBadRequest, requestID, "VALIDATION_ERROR", "input_mode must be one of: file, text", &field)
		return
	}

	project, err := a.service.CreateProject(r.Context(), title, inputMode)
	if err != nil {
		a.writeError(w, http.StatusInternalServerError, requestID, "PROJECT_CREATE_FAILED", "failed to persist project", nil)
		return
	}

	a.writeJSON(w, http.StatusCreated, envelope(project, nil, meta(requestID)))
}

func (a *App) getProject(w http.ResponseWriter, r *http.Request, projectID string) {
	requestID := utils.NewUUID()
	if !utils.IsUUID(projectID) {
		field := "projectId"
		a.writeError(w, http.StatusBadRequest, requestID, "VALIDATION_ERROR", "projectId must be a valid UUID", &field)
		return
	}

	project, docs, err := a.service.GetProject(r.Context(), projectID)
	if err != nil {
		status, code, message := mapDomainError(err)
		a.writeError(w, status, requestID, code, message, nil)
		return
	}

	a.writeJSON(w, http.StatusOK, envelope(map[string]any{
		"id":                 project.ID,
		"title":              project.Title,
		"input_mode":         project.InputMode,
		"status":             project.Status,
		"current_step":       project.CurrentStep,
		"processing_summary": project.ProcessingSummary,
		"documents":          docs,
		"created_at":         project.CreatedAt,
		"updated_at":         project.UpdatedAt,
	}, nil, meta(requestID)))
}

func (a *App) uploadDocument(w http.ResponseWriter, r *http.Request, projectID string) {
	requestID := utils.NewUUID()
	if !utils.IsUUID(projectID) {
		field := "projectId"
		a.writeError(w, http.StatusBadRequest, requestID, "VALIDATION_ERROR", "projectId must be a valid UUID", &field)
		return
	}

	fileHeader, originalFilename, err := a.validateUploadRequest(r)
	if err != nil {
		status, code, field, message := mapUploadError(err, a.config.MaxUploadSizeMB)
		a.writeError(w, status, requestID, code, message, &field)
		return
	}

	project, document, err := a.service.UploadDocument(r.Context(), projectID, originalFilename, fileHeader)
	if err != nil {
		status, code, message := mapDomainError(err)
		a.writeError(w, status, requestID, code, message, nil)
		return
	}

	a.writeJSON(w, http.StatusAccepted, envelope(map[string]any{
		"project":  project,
		"document": document,
	}, nil, meta(requestID)))
}

func (a *App) triggerProcessing(w http.ResponseWriter, r *http.Request, projectID string) {
	requestID := utils.NewUUID()
	if !utils.IsUUID(projectID) {
		field := "projectId"
		a.writeError(w, http.StatusBadRequest, requestID, "VALIDATION_ERROR", "projectId must be a valid UUID", &field)
		return
	}

	project, document, err := a.service.TriggerProcessing(r.Context(), projectID)
	if err != nil {
		status, code, message := mapDomainError(err)
		a.writeError(w, status, requestID, code, message, nil)
		return
	}

	a.writeJSON(w, http.StatusAccepted, envelope(map[string]any{
		"project":  project,
		"document": document,
	}, nil, meta(requestID)))
}

func (a *App) validateUploadRequest(r *http.Request) (*multipart.FileHeader, string, error) {
	if err := r.ParseMultipartForm(int64(a.config.MaxUploadSizeMB+1) * 1024 * 1024); err != nil {
		return nil, "", errInvalidMultipart
	}
	if r.MultipartForm == nil {
		return nil, "", errInvalidMultipart
	}
	if err := validateMultipartFields(r.MultipartForm); err != nil {
		return nil, "", err
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		return nil, "", errMissingFile
	}
	_ = file.Close()

	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(fileHeader.Filename)), ".")
	if ext == "" || !contains(a.config.AllowedFileTypes, ext) {
		return nil, "", fmt.Errorf("%w:%s", errInvalidFileType, ext)
	}
	if fileHeader.Size <= 0 {
		return nil, "", errEmptyFile
	}
	maxBytes := int64(a.config.MaxUploadSizeMB) * 1024 * 1024
	if fileHeader.Size > maxBytes {
		return nil, "", errFileTooLarge
	}

	return fileHeader, strings.TrimSpace(r.FormValue("original_filename")), nil
}

func (a *App) writeError(w http.ResponseWriter, status int, requestID, code, message string, field *string) {
	a.writeJSON(w, status, envelope(nil, &ErrorDetail{Code: code, Message: message, Field: field}, meta(requestID)))
}

func (a *App) writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func meta(requestID string) Meta {
	return Meta{RequestID: requestID, Timestamp: time.Now().UTC()}
}

func envelope(data any, err *ErrorDetail, meta Meta) map[string]any {
	return map[string]any{
		"data":  data,
		"error": err,
		"meta":  meta,
	}
}

func (a *App) applyCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

func contains(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func decodeStrictJSON(r *http.Request, dst any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return normalizeJSONError(err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return errors.New("request body must contain a single JSON object")
	}
	return nil
}

func normalizeJSONError(err error) error {
	if strings.Contains(err.Error(), "unknown field") {
		return fmt.Errorf("request body contains unknown field")
	}
	return errors.New("invalid JSON payload")
}

var (
	errInvalidMultipart = errors.New("invalid multipart form payload")
	errMissingFile      = errors.New("file is required")
	errInvalidFileType  = errors.New("invalid file type")
	errFileTooLarge     = errors.New("file exceeds maximum upload size")
	errEmptyFile        = errors.New("file must not be empty")
	errUnknownFormField = errors.New("multipart form contains unsupported field")
)

func validateMultipartFields(form *multipart.Form) error {
	allowedValues := map[string]bool{"original_filename": true}
	allowedFiles := map[string]bool{"file": true}
	for key := range form.Value {
		if !allowedValues[key] {
			return fmt.Errorf("%w: %s", errUnknownFormField, key)
		}
	}
	for key, entries := range form.File {
		if !allowedFiles[key] {
			return fmt.Errorf("%w: %s", errUnknownFormField, key)
		}
		if key == "file" && len(entries) != 1 {
			return errInvalidMultipart
		}
	}
	return nil
}

func mapUploadError(err error, maxUploadSizeMB int) (status int, code string, field string, message string) {
	field = "file"
	switch {
	case errors.Is(err, errInvalidMultipart):
		return http.StatusBadRequest, "VALIDATION_ERROR", field, "invalid multipart form payload"
	case errors.Is(err, errMissingFile):
		return http.StatusBadRequest, "VALIDATION_ERROR", field, "file is required"
	case errors.Is(err, errEmptyFile):
		return http.StatusBadRequest, "VALIDATION_ERROR", field, "file must not be empty"
	case errors.Is(err, errFileTooLarge):
		return http.StatusBadRequest, "FILE_TOO_LARGE", field, fmt.Sprintf("file exceeds %d MB", maxUploadSizeMB)
	case errors.Is(err, errUnknownFormField):
		return http.StatusBadRequest, "VALIDATION_ERROR", field, "multipart form contains unsupported field"
	case strings.Contains(err.Error(), errInvalidFileType.Error()):
		ext := strings.TrimPrefix(strings.TrimPrefix(err.Error(), errInvalidFileType.Error()+":"), ".")
		if ext == "" {
			return http.StatusBadRequest, "INVALID_FILE_TYPE", field, "file type is not allowed"
		}
		return http.StatusBadRequest, "INVALID_FILE_TYPE", field, fmt.Sprintf("file type .%s is not allowed", ext)
	default:
		return http.StatusBadRequest, "VALIDATION_ERROR", field, err.Error()
	}
}

func mapDomainError(err error) (status int, code, message string) {
	switch {
	case err == nil:
		return http.StatusOK, "", ""
	case errors.Is(err, projects.ErrProjectNotFound):
		return http.StatusNotFound, "PROJECT_NOT_FOUND", "project not found"
	case errors.Is(err, projects.ErrDocumentNotFound), errors.Is(err, projects.ErrNoDocumentsForProject):
		return http.StatusNotFound, "DOCUMENT_NOT_FOUND", "document not found"
	default:
		return http.StatusInternalServerError, "PERSISTENCE_ERROR", "failed to query persistent storage"
	}
}
