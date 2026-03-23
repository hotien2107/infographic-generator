package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
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

func (s *memoryStore) CreateProject(_ context.Context, title string, inputMode projects.InputMode) (projects.Project, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	project := projects.Project{
		ID:          utils.NewUUID(),
		Title:       title,
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

func TestCreateProjectHappyPath(t *testing.T) {
	handler := newTestHandler(t, false)

	payload := []byte(`{"title":"Sprint 1 Project","input_mode":"file"}`)
	res := performJSONRequest(t, handler, http.MethodPost, "/api/v1/projects", payload)
	if res.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", res.Code)
	}

	var response struct {
		Data  projects.Project `json:"data"`
		Error any              `json:"error"`
	}
	decodeJSON(t, res, &response)
	if response.Data.Status != projects.StatusDraft || response.Data.CurrentStep != projects.StepWaitingUpload {
		t.Fatalf("unexpected project state: %+v", response.Data)
	}
	if response.Error != nil {
		t.Fatalf("expected nil error, got %+v", response.Error)
	}
}

func TestCreateProjectValidationErrors(t *testing.T) {
	handler := newTestHandler(t, false)

	t.Run("reject unknown field", func(t *testing.T) {
		res := performJSONRequest(t, handler, http.MethodPost, "/api/v1/projects", []byte(`{"title":"Valid title","input_mode":"file","extra":"nope"}`))
		assertErrorCode(t, res, http.StatusBadRequest, "VALIDATION_ERROR")
	})

	t.Run("reject short title", func(t *testing.T) {
		res := performJSONRequest(t, handler, http.MethodPost, "/api/v1/projects", []byte(`{"title":"ab","input_mode":"file"}`))
		assertErrorCode(t, res, http.StatusBadRequest, "VALIDATION_ERROR")
	})

	t.Run("reject invalid input mode", func(t *testing.T) {
		res := performJSONRequest(t, handler, http.MethodPost, "/api/v1/projects", []byte(`{"title":"Hello","input_mode":"voice"}`))
		assertErrorCode(t, res, http.StatusBadRequest, "VALIDATION_ERROR")
	})
}

func TestGetProjectHappyPath(t *testing.T) {
	handler := newTestHandler(t, false)
	projectID := createProject(t, handler, "Get Detail", "file")

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/"+projectID, nil)
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.Code)
	}

	var response struct {
		Data struct {
			ID        string               `json:"id"`
			Documents []documents.Document `json:"documents"`
		} `json:"data"`
	}
	decodeJSON(t, res, &response)
	if response.Data.ID != projectID {
		t.Fatalf("expected project id %s, got %s", projectID, response.Data.ID)
	}
	if len(response.Data.Documents) != 0 {
		t.Fatalf("expected no documents, got %d", len(response.Data.Documents))
	}
}

func TestGetProjectNotFound(t *testing.T) {
	handler := newTestHandler(t, false)
	missingID := utils.NewUUID()
	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/"+missingID, nil)
	handler.ServeHTTP(res, req)
	assertErrorCode(t, res, http.StatusNotFound, "PROJECT_NOT_FOUND")
}

func TestGetProjectInvalidUUID(t *testing.T) {
	handler := newTestHandler(t, false)
	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/not-a-uuid", nil)
	handler.ServeHTTP(res, req)
	assertErrorCode(t, res, http.StatusBadRequest, "VALIDATION_ERROR")
}

func TestUploadFileHappyPath(t *testing.T) {
	handler := newTestHandler(t, false)
	projectID := createProject(t, handler, "Upload Target", "file")

	res := uploadDocument(t, handler, projectID, "sample.txt", "brief.txt", []byte("hello sprint 1"), nil)
	if res.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", res.Code)
	}

	var response struct {
		Data struct {
			Project  projects.Project   `json:"project"`
			Document documents.Document `json:"document"`
		} `json:"data"`
	}
	decodeJSON(t, res, &response)
	if response.Data.Project.Status != projects.StatusUploaded || response.Data.Project.CurrentStep != projects.StepUploaded {
		t.Fatalf("unexpected project state: %+v", response.Data.Project)
	}
	if response.Data.Document.Filename != "brief.txt" {
		t.Fatalf("unexpected document filename: %s", response.Data.Document.Filename)
	}
}

func TestUploadDocumentRejectsUnsupportedType(t *testing.T) {
	handler := newTestHandler(t, false)
	projectID := createProject(t, handler, "Upload Target", "file")
	res := uploadDocument(t, handler, projectID, "script.exe", "script.exe", []byte("bad file"), nil)
	assertErrorCode(t, res, http.StatusBadRequest, "INVALID_FILE_TYPE")
}

func TestUploadDocumentRejectsLargeFile(t *testing.T) {
	handler := newTestHandler(t, false)
	projectID := createProject(t, handler, "Upload Target", "file")
	res := uploadDocument(t, handler, projectID, "big.txt", "big.txt", bytes.Repeat([]byte("a"), 2*1024*1024), nil)
	assertErrorCode(t, res, http.StatusBadRequest, "FILE_TOO_LARGE")
}

func TestUploadDocumentRequiresFile(t *testing.T) {
	handler := newTestHandler(t, false)
	projectID := createProject(t, handler, "Upload Target", "file")

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("original_filename", "missing.txt")
	_ = writer.Close()

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/"+projectID+"/documents", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	handler.ServeHTTP(res, req)
	assertErrorCode(t, res, http.StatusBadRequest, "VALIDATION_ERROR")
}

func TestUploadDocumentQueuesAndProcesses(t *testing.T) {
	handler := newTestHandler(t, true)
	projectID := createProject(t, handler, "Async Flow", "file")

	res := uploadDocument(t, handler, projectID, "story.txt", "story.txt", []byte("hello"), nil)
	if res.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", res.Code)
	}

	waitForProjectState(t, handler, projectID, projects.StatusProcessed, documents.StatusProcessed)
}

func TestTriggerProcessingFailureFlow(t *testing.T) {
	handler := newTestHandler(t, false)
	projectID := createProject(t, handler, "Manual Flow", "file")
	uploadDocument(t, handler, projectID, "will-fail.txt", "will-fail.txt", []byte("boom"), nil)

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/"+projectID+"/processing", nil)
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", res.Code)
	}

	waitForProjectState(t, handler, projectID, projects.StatusFailed, documents.StatusFailed)
}

func TestProjectDetailReflectsProcessingSummary(t *testing.T) {
	handler := newTestHandler(t, true)
	projectID := createProject(t, handler, "Summary", "file")
	uploadDocument(t, handler, projectID, "report.txt", "report.txt", []byte("summary"), nil)
	waitForProjectState(t, handler, projectID, projects.StatusProcessed, documents.StatusProcessed)

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/"+projectID, nil)
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.Code)
	}

	var response struct {
		Data struct {
			Status            string `json:"status"`
			ProcessingSummary struct {
				ProcessedDocuments int `json:"processed_documents"`
				TotalDocuments     int `json:"total_documents"`
			} `json:"processing_summary"`
		} `json:"data"`
	}
	decodeJSON(t, res, &response)
	if response.Data.Status != string(projects.StatusProcessed) {
		t.Fatalf("expected processed project, got %s", response.Data.Status)
	}
	if response.Data.ProcessingSummary.TotalDocuments != 1 || response.Data.ProcessingSummary.ProcessedDocuments != 1 {
		t.Fatalf("unexpected summary: %+v", response.Data.ProcessingSummary)
	}
}

func performJSONRequest(t *testing.T, handler http.Handler, method, path string, payload []byte) *httptest.ResponseRecorder {
	t.Helper()
	res := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(res, req)
	return res
}

func createProject(t *testing.T, handler http.Handler, title, inputMode string) string {
	t.Helper()
	res := performJSONRequest(t, handler, http.MethodPost, "/api/v1/projects", []byte(`{"title":"`+title+`","input_mode":"`+inputMode+`"}`))
	if res.Code != http.StatusCreated {
		t.Fatalf("expected create project to return 201, got %d", res.Code)
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
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fileWriter, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := io.Copy(fileWriter, bytes.NewReader(content)); err != nil {
		t.Fatalf("write file content: %v", err)
	}
	if originalFilename != "" {
		if err := writer.WriteField("original_filename", originalFilename); err != nil {
			t.Fatalf("write original filename field: %v", err)
		}
	}
	for key, value := range extraFields {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatalf("write extra field: %v", err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/"+projectID+"/documents", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	handler.ServeHTTP(res, req)
	return res
}

func assertErrorCode(t *testing.T, res *httptest.ResponseRecorder, wantStatus int, wantCode string) {
	t.Helper()
	if res.Code != wantStatus {
		t.Fatalf("expected status %d, got %d", wantStatus, res.Code)
	}
	var payload struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	decodeJSON(t, res, &payload)
	if payload.Error.Code != wantCode {
		t.Fatalf("expected error code %s, got %s", wantCode, payload.Error.Code)
	}
}

func decodeJSON(t *testing.T, res *httptest.ResponseRecorder, dst any) {
	t.Helper()
	if err := json.NewDecoder(res.Body).Decode(dst); err != nil {
		t.Fatalf("decode JSON response: %v", err)
	}
}

func waitForProjectState(t *testing.T, handler http.Handler, projectID string, expectedProject projects.Status, expectedDocument documents.Status) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/"+projectID, nil)
		handler.ServeHTTP(res, req)
		if res.Code != http.StatusOK {
			t.Fatalf("expected 200 while polling, got %d", res.Code)
		}
		var payload struct {
			Data struct {
				Status    string               `json:"status"`
				Documents []documents.Document `json:"documents"`
			} `json:"data"`
		}
		decodeJSON(t, res, &payload)
		if payload.Data.Status == string(expectedProject) && len(payload.Data.Documents) == 1 && payload.Data.Documents[0].Status == expectedDocument {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for project %s and document %s", expectedProject, expectedDocument)
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
			if doc.ErrorMessage != nil {
				message := strings.Clone(*doc.ErrorMessage)
				summary.LastError = &message
			}
		}
	}
	return summary
}
