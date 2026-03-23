package api

import (
	"encoding/json"
	"fmt"
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

type App struct {
	config  config.Config
	store   *projects.Store
	storage *storage.LocalStorage
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

func New(cfg config.Config) *App {
	return &App{
		config:  cfg,
		store:   projects.NewStore(),
		storage: storage.NewLocalStorage(cfg.StorageDir),
	}
}

func (a *App) Handler() http.Handler {
	return http.HandlerFunc(a.serveHTTP)
}

func (a *App) serveHTTP(w http.ResponseWriter, r *http.Request) {
	a.applyCORS(w, r)
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
	}

	a.writeJSON(w, http.StatusNotFound, map[string]any{
		"data":  nil,
		"error": ErrorDetail{Code: "PROJECT_NOT_FOUND", Message: "route not found", Field: nil},
		"meta":  meta(utils.NewUUID()),
	})
}

func (a *App) applyCORS(w http.ResponseWriter, r *http.Request) {
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	allowedOrigin := strings.TrimSpace(a.config.FrontendOrigin)
	if origin != "" && origin == allowedOrigin {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Vary", "Origin")
	}

	w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func (a *App) createProject(w http.ResponseWriter, r *http.Request) {
	requestID := utils.NewUUID()
	defer r.Body.Close()

	var payload createProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		a.writeError(w, http.StatusBadRequest, requestID, "VALIDATION_ERROR", "invalid JSON payload", nil)
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

	project := a.store.CreateProject(title, inputMode)
	a.writeJSON(w, http.StatusCreated, map[string]any{
		"data":  project,
		"error": nil,
		"meta":  meta(requestID),
	})
}

func (a *App) getProject(w http.ResponseWriter, r *http.Request, projectID string) {
	requestID := utils.NewUUID()
	if !utils.IsUUID(projectID) {
		field := "projectId"
		a.writeError(w, http.StatusBadRequest, requestID, "VALIDATION_ERROR", "projectId must be a valid UUID", &field)
		return
	}

	project, docs, err := a.store.GetProject(projectID)
	if err != nil {
		a.writeError(w, http.StatusNotFound, requestID, "PROJECT_NOT_FOUND", "project not found", nil)
		return
	}

	a.writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"id":           project.ID,
			"title":        project.Title,
			"input_mode":   project.InputMode,
			"status":       project.Status,
			"current_step": project.CurrentStep,
			"documents":    docs,
			"created_at":   project.CreatedAt,
			"updated_at":   project.UpdatedAt,
		},
		"error": nil,
		"meta":  meta(requestID),
	})
}

func (a *App) uploadDocument(w http.ResponseWriter, r *http.Request, projectID string) {
	requestID := utils.NewUUID()
	if !utils.IsUUID(projectID) {
		field := "projectId"
		a.writeError(w, http.StatusBadRequest, requestID, "VALIDATION_ERROR", "projectId must be a valid UUID", &field)
		return
	}

	if err := r.ParseMultipartForm(int64(a.config.MaxUploadSizeMB+1) * 1024 * 1024); err != nil {
		field := "file"
		a.writeError(w, http.StatusBadRequest, requestID, "VALIDATION_ERROR", "invalid multipart form payload", &field)
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		field := "file"
		a.writeError(w, http.StatusBadRequest, requestID, "VALIDATION_ERROR", "file is required", &field)
		return
	}
	_ = file.Close()

	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(fileHeader.Filename)), ".")
	if !contains(a.config.AllowedFileTypes, ext) {
		field := "file"
		a.writeError(w, http.StatusBadRequest, requestID, "INVALID_FILE_TYPE", fmt.Sprintf("file type .%s is not allowed", ext), &field)
		return
	}

	if fileHeader.Size <= 0 {
		field := "file"
		a.writeError(w, http.StatusBadRequest, requestID, "VALIDATION_ERROR", "file must not be empty", &field)
		return
	}

	maxBytes := int64(a.config.MaxUploadSizeMB) * 1024 * 1024
	if fileHeader.Size > maxBytes {
		field := "file"
		a.writeError(w, http.StatusBadRequest, requestID, "FILE_TOO_LARGE", fmt.Sprintf("file exceeds %d MB", a.config.MaxUploadSizeMB), &field)
		return
	}

	storageKey, err := a.storage.Save(fileHeader)
	if err != nil {
		a.writeError(w, http.StatusInternalServerError, requestID, "VALIDATION_ERROR", "failed to persist uploaded file", nil)
		return
	}

	document := documents.Document{
		ID:         utils.NewUUID(),
		ProjectID:  projectID,
		Filename:   firstNonEmpty(strings.TrimSpace(r.FormValue("original_filename")), fileHeader.Filename),
		MimeType:   mimeTypeForExtension(ext),
		SizeBytes:  fileHeader.Size,
		StorageKey: storageKey,
		Status:     documents.StatusUploaded,
		CreatedAt:  time.Now().UTC(),
	}

	project, _, err := a.store.AddDocument(projectID, document)
	if err != nil {
		a.writeError(w, http.StatusNotFound, requestID, "PROJECT_NOT_FOUND", "project not found", nil)
		return
	}

	a.writeJSON(w, http.StatusAccepted, map[string]any{
		"data": map[string]any{
			"project":  project,
			"document": document,
		},
		"error": nil,
		"meta":  meta(requestID),
	})
}

func (a *App) writeError(w http.ResponseWriter, status int, requestID, code, message string, field *string) {
	a.writeJSON(w, status, map[string]any{
		"data":  nil,
		"error": ErrorDetail{Code: code, Message: message, Field: field},
		"meta":  meta(requestID),
	})
}

func (a *App) writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func meta(requestID string) Meta {
	return Meta{RequestID: requestID, Timestamp: time.Now().UTC()}
}

func contains(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func mimeTypeForExtension(extension string) string {
	switch extension {
	case "pdf":
		return "application/pdf"
	case "docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case "txt":
		return "text/plain"
	default:
		return "application/octet-stream"
	}
}
