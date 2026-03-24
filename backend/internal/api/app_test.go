package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"infographic-generator/backend/internal/api"
	"infographic-generator/backend/internal/config"
	"infographic-generator/backend/internal/modules/documents"
	"infographic-generator/backend/internal/modules/projects"
	"infographic-generator/backend/internal/processing"
	"infographic-generator/backend/internal/utils"
)

type memoryStore struct {
	mu        sync.RWMutex
	projects  map[string]projects.Project
	documents map[string][]documents.Document
}

func newMemoryStore() *memoryStore {
	return &memoryStore{
		projects:  make(map[string]projects.Project),
		documents: make(map[string][]documents.Document),
	}
}

func (s *memoryStore) GetDashboardSummary(_ context.Context) (projects.DashboardSummary, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	summary := projects.DashboardSummary{TotalProjects: len(s.projects)}
	for projectID, project := range s.projects {
		summary.TotalDocuments += len(s.documents[projectID])
		switch project.Status {
		case projects.StatusProcessing:
			summary.ProcessingProjects++
		case projects.StatusProcessed:
			summary.CompletedProjects++
		case projects.StatusFailed:
			summary.AttentionProjects++
		case projects.StatusDraft:
			summary.DraftProjects++
		}
	}
	return summary, nil
}

func (s *memoryStore) ListProjects(_ context.Context) ([]projects.ProjectListItem, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]projects.ProjectListItem, 0, len(s.projects))
	for projectID, project := range s.projects {
		project.ProcessingSummary = buildSummary(s.documents[projectID])
		items = append(items, projects.ProjectListItem{Project: project, DocumentCount: len(s.documents[projectID])})
	}
	return items, nil
}

func (s *memoryStore) CreateProject(_ context.Context, title, description string, inputMode projects.InputMode) (projects.Project, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	project := projects.Project{
		ID:          utils.NewUUID(),
		Title:       title,
		Description: description,
		InputMode:   inputMode,
		Status:      projects.StatusDraft,
		CurrentStep: projects.StepWaitingUpload,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	s.projects[project.ID] = project
	return project, nil
}

func (s *memoryStore) GetProject(_ context.Context, projectID string) (projects.Project, []documents.Document, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	project, ok := s.projects[projectID]
	if !ok {
		return projects.Project{}, nil, projects.ErrProjectNotFound
	}
	docs := append([]documents.Document(nil), s.documents[projectID]...)
	project.ProcessingSummary = buildSummary(docs)
	return project, docs, nil
}

func (s *memoryStore) UpdateProject(_ context.Context, projectID string, update projects.ProjectUpdate) (projects.Project, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	project, ok := s.projects[projectID]
	if !ok {
		return projects.Project{}, projects.ErrProjectNotFound
	}
	if update.Title != nil {
		project.Title = *update.Title
	}
	if update.Description != nil {
		project.Description = *update.Description
	}
	if update.InputMode != nil {
		project.InputMode = *update.InputMode
	}
	project.UpdatedAt = time.Now().UTC()
	project.ProcessingSummary = buildSummary(s.documents[projectID])
	s.projects[projectID] = project
	return project, nil
}

func (s *memoryStore) DeleteProject(_ context.Context, projectID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.projects[projectID]; !ok {
		return projects.ErrProjectNotFound
	}
	delete(s.projects, projectID)
	delete(s.documents, projectID)
	return nil
}

func (s *memoryStore) AddDocument(_ context.Context, projectID string, document documents.Document) (projects.Project, []documents.Document, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	project, ok := s.projects[projectID]
	if !ok {
		return projects.Project{}, nil, projects.ErrProjectNotFound
	}
	project.Status = projects.StatusUploaded
	project.CurrentStep = projects.StepUploaded
	project.UpdatedAt = time.Now().UTC()
	s.projects[projectID] = project
	s.documents[projectID] = append(s.documents[projectID], document)
	docs := append([]documents.Document(nil), s.documents[projectID]...)
	project.ProcessingSummary = buildSummary(docs)
	return project, docs, nil
}

func (s *memoryStore) ListDocuments(_ context.Context, projectID string) ([]documents.Document, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if _, ok := s.projects[projectID]; !ok {
		return nil, projects.ErrProjectNotFound
	}
	return append([]documents.Document(nil), s.documents[projectID]...), nil
}

func (s *memoryStore) UpdateDocument(_ context.Context, projectID, documentID, filename string) (documents.Document, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.projects[projectID]; !ok {
		return documents.Document{}, projects.ErrProjectNotFound
	}
	for index, doc := range s.documents[projectID] {
		if doc.ID != documentID {
			continue
		}
		doc.Filename = filename
		doc.UpdatedAt = time.Now().UTC()
		s.documents[projectID][index] = doc
		return doc, nil
	}
	return documents.Document{}, projects.ErrDocumentNotFound
}

func (s *memoryStore) DeleteDocument(_ context.Context, projectID, documentID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	project, ok := s.projects[projectID]
	if !ok {
		return projects.ErrProjectNotFound
	}
	docs := s.documents[projectID]
	for index, doc := range docs {
		if doc.ID != documentID {
			continue
		}
		docs = append(docs[:index], docs[index+1:]...)
		s.documents[projectID] = docs
		project.Status, project.CurrentStep = deriveProjectState(docs)
		project.UpdatedAt = time.Now().UTC()
		s.projects[projectID] = project
		return nil
	}
	return projects.ErrDocumentNotFound
}

func (s *memoryStore) GetLatestDocument(_ context.Context, projectID string) (documents.Document, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if _, ok := s.projects[projectID]; !ok {
		return documents.Document{}, projects.ErrProjectNotFound
	}
	docs := s.documents[projectID]
	if len(docs) == 0 {
		return documents.Document{}, projects.ErrNoDocumentsForProject
	}
	return docs[len(docs)-1], nil
}

func (s *memoryStore) UpdateDocumentProcessing(_ context.Context, projectID, documentID string, params projects.DocumentProcessingUpdate) (projects.Project, documents.Document, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	project, ok := s.projects[projectID]
	if !ok {
		return projects.Project{}, documents.Document{}, projects.ErrProjectNotFound
	}
	project.Status = params.ProjectStatus
	project.CurrentStep = params.ProjectStep
	project.UpdatedAt = time.Now().UTC()
	s.projects[projectID] = project

	docs := s.documents[projectID]
	for index, doc := range docs {
		if doc.ID != documentID {
			continue
		}
		doc.Status = params.DocumentStatus
		doc.ProcessingStartedAt = params.ProcessingStartedAt
		doc.ProcessingFinishedAt = params.ProcessingFinishedAt
		doc.ErrorMessage = params.ErrorMessage
		doc.ExtractedTextPreview = params.ExtractedTextPreview
		doc.UpdatedAt = time.Now().UTC()
		docs[index] = doc
		s.documents[projectID] = docs
		project.ProcessingSummary = buildSummary(docs)
		return project, doc, nil
	}
	return projects.Project{}, documents.Document{}, projects.ErrDocumentNotFound
}

func (s *memoryStore) Close() {}

type fakeBlobStorage struct{}

func (s *fakeBlobStorage) Save(_ context.Context, fileHeader *multipart.FileHeader) (string, error) {
	return filepath.Join("documents", fileHeader.Filename), nil
}

func (s *fakeBlobStorage) Close() error { return nil }

func newTestHandler(t *testing.T, autoProcess bool) http.Handler {
	t.Helper()
	cfg := config.Config{
		AppEnv:                "test",
		Port:                  "0",
		MaxUploadSizeMB:       1,
		AllowedFileTypes:      []string{"pdf", "docx", "txt"},
		MultipartThresholdMB:  16,
		MultipartPartSizeMB:   8,
		AutoProcessDocuments:  autoProcess,
		ProcessingQueueBuffer: 8,
		ProcessingStepDelay:   20 * time.Millisecond,
		ProcessingFailPattern: "fail",
	}

	store := newMemoryStore()
	service := projects.NewService(store, &fakeBlobStorage{}, nil, autoProcess)
	worker := processing.NewWorker(service, cfg.ProcessingQueueBuffer, cfg.ProcessingStepDelay, cfg.ProcessingFailPattern)
	service.SetProcessor(worker)
	worker.Start(context.Background())
	app := api.New(cfg, store, &fakeBlobStorage{}, service)
	t.Cleanup(app.Close)
	return app.Handler()
}

func TestProjectManagementFlow(t *testing.T) {
	handler := newTestHandler(t, false)

	createRes := performJSONRequest(t, handler, http.MethodPost, "/api/v1/projects", []byte(`{"title":"Báo cáo quý 2","description":"Tổng hợp số liệu chiến dịch","input_mode":"file"}`))
	if createRes.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", createRes.Code)
	}

	var created struct {
		Data map[string]any `json:"data"`
	}
	decodeJSON(t, createRes, &created)
	projectID := created.Data["id"].(string)

	listRes := performJSONRequest(t, handler, http.MethodGet, "/api/v1/projects", nil)
	if listRes.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", listRes.Code)
	}

	updateRes := performJSONRequest(t, handler, http.MethodPatch, "/api/v1/projects/"+projectID, []byte(`{"title":"Báo cáo quý 2 đã cập nhật"}`))
	if updateRes.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", updateRes.Code)
	}

	getRes := performJSONRequest(t, handler, http.MethodGet, "/api/v1/projects/"+projectID, nil)
	if getRes.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", getRes.Code)
	}

	deleteRes := performJSONRequest(t, handler, http.MethodDelete, "/api/v1/projects/"+projectID, nil)
	if deleteRes.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", deleteRes.Code)
	}
}

func TestDocumentManagementFlow(t *testing.T) {
	handler := newTestHandler(t, false)
	projectID := createProject(t, handler, "Tài liệu khách hàng", "Tổng hợp bản thảo", "file")

	uploadRes := uploadDocument(t, handler, projectID, "sample.txt", "brief.txt", []byte("hello product"), nil)
	if uploadRes.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", uploadRes.Code)
	}

	var uploaded struct {
		Data struct {
			Document map[string]any `json:"document"`
		} `json:"data"`
	}
	decodeJSON(t, uploadRes, &uploaded)
	documentID := uploaded.Data.Document["id"].(string)
	if _, exists := uploaded.Data.Document["storage_key"]; exists {
		t.Fatalf("storage_key must not be exposed in API response")
	}

	listRes := performJSONRequest(t, handler, http.MethodGet, "/api/v1/projects/"+projectID+"/documents", nil)
	if listRes.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", listRes.Code)
	}

	updateRes := performJSONRequest(t, handler, http.MethodPatch, "/api/v1/projects/"+projectID+"/documents/"+documentID, []byte(`{"filename":"customer-brief.txt"}`))
	if updateRes.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", updateRes.Code)
	}

	deleteRes := performJSONRequest(t, handler, http.MethodDelete, "/api/v1/projects/"+projectID+"/documents/"+documentID, nil)
	if deleteRes.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", deleteRes.Code)
	}
}

func TestDashboardSummary(t *testing.T) {
	handler := newTestHandler(t, false)
	projectID := createProject(t, handler, "Tổng quan", "Mô tả", "file")
	uploadDocument(t, handler, projectID, "sample.txt", "summary.txt", []byte("summary"), nil)

	res := performJSONRequest(t, handler, http.MethodGet, "/api/v1/dashboard/summary", nil)
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.Code)
	}

	var payload struct {
		Data projects.DashboardSummary `json:"data"`
	}
	decodeJSON(t, res, &payload)
	if payload.Data.TotalProjects != 1 || payload.Data.TotalDocuments != 1 {
		t.Fatalf("unexpected dashboard summary: %+v", payload.Data)
	}
}

func TestValidationErrors(t *testing.T) {
	handler := newTestHandler(t, false)

	res := performJSONRequest(t, handler, http.MethodPost, "/api/v1/projects", []byte(`{"title":"ab","input_mode":"file"}`))
	assertErrorCode(t, res, http.StatusBadRequest, "VALIDATION_ERROR")

	res = performJSONRequest(t, handler, http.MethodPatch, "/api/v1/projects/not-a-uuid", []byte(`{"title":"abc"}`))
	assertErrorCode(t, res, http.StatusBadRequest, "VALIDATION_ERROR")
}

func createProject(t *testing.T, handler http.Handler, title, description, inputMode string) string {
	t.Helper()
	payload := map[string]string{"title": title, "description": description, "input_mode": inputMode}
	body, _ := json.Marshal(payload)
	res := performJSONRequest(t, handler, http.MethodPost, "/api/v1/projects", body)
	if res.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", res.Code)
	}
	var response struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	decodeJSON(t, res, &response)
	return response.Data.ID
}

func uploadDocument(t *testing.T, handler http.Handler, projectID, filename, originalFilename string, content []byte, extraFields map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	fileWriter, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := fileWriter.Write(content); err != nil {
		t.Fatalf("write file content: %v", err)
	}
	if originalFilename != "" {
		if err := writer.WriteField("original_filename", originalFilename); err != nil {
			t.Fatalf("write original filename: %v", err)
		}
	}
	for key, value := range extraFields {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatalf("write extra field: %v", err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/"+projectID+"/documents", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	handler.ServeHTTP(res, req)
	return res
}

func performJSONRequest(t *testing.T, handler http.Handler, method, path string, body []byte) *httptest.ResponseRecorder {
	t.Helper()
	var reader *bytes.Reader
	if body == nil {
		reader = bytes.NewReader(nil)
	} else {
		reader = bytes.NewReader(body)
	}
	res := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, reader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	handler.ServeHTTP(res, req)
	return res
}

func decodeJSON(t *testing.T, res *httptest.ResponseRecorder, target any) {
	t.Helper()
	if err := json.NewDecoder(strings.NewReader(res.Body.String())).Decode(target); err != nil {
		t.Fatalf("decode json: %v", err)
	}
}

func assertErrorCode(t *testing.T, res *httptest.ResponseRecorder, expectedStatus int, expectedCode string) {
	t.Helper()
	if res.Code != expectedStatus {
		t.Fatalf("expected status %d, got %d", expectedStatus, res.Code)
	}
	var payload struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	decodeJSON(t, res, &payload)
	if payload.Error.Code != expectedCode {
		t.Fatalf("expected code %s, got %s", expectedCode, payload.Error.Code)
	}
}

func buildSummary(docs []documents.Document) projects.ProcessingSummary {
	summary := projects.ProcessingSummary{TotalDocuments: len(docs)}
	for _, doc := range docs {
		switch doc.Status {
		case documents.StatusUploaded:
			summary.UploadedDocuments++
		case documents.StatusQueued:
			summary.QueuedDocuments++
		case documents.StatusProcessing:
			summary.ProcessingDocuments++
		case documents.StatusProcessed:
			summary.ProcessedDocuments++
		case documents.StatusFailed:
			summary.FailedDocuments++
		}
	}
	return summary
}

func deriveProjectState(docs []documents.Document) (projects.Status, projects.Step) {
	if len(docs) == 0 {
		return projects.StatusDraft, projects.StepWaitingUpload
	}
	for _, doc := range docs {
		if doc.Status == documents.StatusProcessing {
			return projects.StatusProcessing, projects.StepExtracting
		}
		if doc.Status == documents.StatusQueued {
			return projects.StatusProcessing, projects.StepQueuedProcessing
		}
	}
	for _, doc := range docs {
		if doc.Status == documents.StatusFailed {
			return projects.StatusFailed, projects.StepFailed
		}
	}
	allProcessed := true
	for _, doc := range docs {
		if doc.Status != documents.StatusProcessed {
			allProcessed = false
			break
		}
	}
	if allProcessed {
		return projects.StatusProcessed, projects.StepReadyForGeneration
	}
	return projects.StatusUploaded, projects.StepUploaded
}
