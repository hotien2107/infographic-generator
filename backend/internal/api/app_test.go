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
	"testing"
	"time"

	"infographic-generator/backend/internal/api"
	"infographic-generator/backend/internal/config"
	"infographic-generator/backend/internal/modules/documents"
	"infographic-generator/backend/internal/modules/projects"
	"infographic-generator/backend/internal/utils"
)

type memoryStore struct {
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
	project, ok := s.projects[projectID]
	if !ok {
		return projects.Project{}, nil, projects.ErrProjectNotFound
	}
	return project, append([]documents.Document(nil), s.documents[projectID]...), nil
}

func (s *memoryStore) AddDocument(_ context.Context, projectID string, document documents.Document) (projects.Project, []documents.Document, error) {
	project, ok := s.projects[projectID]
	if !ok {
		return projects.Project{}, nil, projects.ErrProjectNotFound
	}
	project.Status = projects.StatusUploaded
	project.CurrentStep = projects.StepUploaded
	project.UpdatedAt = time.Now().UTC()
	s.projects[projectID] = project
	s.documents[projectID] = append(s.documents[projectID], document)
	return project, append([]documents.Document(nil), s.documents[projectID]...), nil
}

func (s *memoryStore) Close() {}

type fakeBlobStorage struct{}

func (s *fakeBlobStorage) Save(_ context.Context, fileHeader *multipart.FileHeader) (string, error) {
	return filepath.Join("documents", fileHeader.Filename), nil
}

func (s *fakeBlobStorage) Close() error { return nil }

func newTestHandler(t *testing.T) http.Handler {
	t.Helper()
	cfg := config.Config{
		AppEnv:               "test",
		Port:                 "0",
		MaxUploadSizeMB:      1,
		AllowedFileTypes:     []string{"pdf", "docx", "txt"},
		MultipartThresholdMB: 16,
		MultipartPartSizeMB:  8,
	}

	app := api.New(cfg, newMemoryStore(), &fakeBlobStorage{})
	t.Cleanup(app.Close)
	return app.Handler()
}

func TestCreateProjectAndGetDetail(t *testing.T) {
	handler := newTestHandler(t)

	payload := map[string]string{"title": "Sprint 1 Project", "input_mode": "file"}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	if res.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", res.Code)
	}

	var created struct {
		Data struct {
			ID          string `json:"id"`
			Status      string `json:"status"`
			CurrentStep string `json:"current_step"`
		} `json:"data"`
	}
	if err := json.NewDecoder(res.Body).Decode(&created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}

	if created.Data.Status != "draft" || created.Data.CurrentStep != "waiting_for_upload" {
		t.Fatalf("unexpected project state: %+v", created.Data)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/projects/"+created.Data.ID, nil)
	getRes := httptest.NewRecorder()
	handler.ServeHTTP(getRes, getReq)

	if getRes.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", getRes.Code)
	}

	var detail struct {
		Data struct {
			Documents []any `json:"documents"`
		} `json:"data"`
	}
	if err := json.NewDecoder(getRes.Body).Decode(&detail); err != nil {
		t.Fatalf("decode detail response: %v", err)
	}

	if len(detail.Data.Documents) != 0 {
		t.Fatalf("expected no documents, got %d", len(detail.Data.Documents))
	}
}

func TestUploadDocumentUpdatesProjectState(t *testing.T) {
	handler := newTestHandler(t)
	projectID := createProject(t, handler)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fileWriter, err := writer.CreateFormFile("file", "sample.txt")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := io.Copy(fileWriter, bytes.NewBufferString("hello sprint 1")); err != nil {
		t.Fatalf("write form file: %v", err)
	}
	_ = writer.WriteField("original_filename", "brief.txt")
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/"+projectID+"/documents", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	if res.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", res.Code)
	}

	var uploaded struct {
		Data struct {
			Project struct {
				Status      string `json:"status"`
				CurrentStep string `json:"current_step"`
			} `json:"project"`
			Document struct {
				Filename   string `json:"filename"`
				MimeType   string `json:"mime_type"`
				StorageKey string `json:"storage_key"`
			} `json:"document"`
		} `json:"data"`
	}
	if err := json.NewDecoder(res.Body).Decode(&uploaded); err != nil {
		t.Fatalf("decode upload response: %v", err)
	}

	if uploaded.Data.Project.Status != "uploaded" || uploaded.Data.Project.CurrentStep != "uploaded" {
		t.Fatalf("unexpected project state: %+v", uploaded.Data.Project)
	}
	if uploaded.Data.Document.Filename != "brief.txt" || uploaded.Data.Document.MimeType != "text/plain" {
		t.Fatalf("unexpected document data: %+v", uploaded.Data.Document)
	}
	if uploaded.Data.Document.StorageKey != filepath.Join("documents", "sample.txt") {
		t.Fatalf("unexpected storage key: %s", uploaded.Data.Document.StorageKey)
	}
}

func TestUploadDocumentRejectsUnsupportedType(t *testing.T) {
	handler := newTestHandler(t)
	projectID := createProject(t, handler)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fileWriter, err := writer.CreateFormFile("file", "script.exe")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := io.Copy(fileWriter, bytes.NewBufferString("bad file")); err != nil {
		t.Fatalf("write form file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/"+projectID+"/documents", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.Code)
	}

	var response struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		t.Fatalf("decode error response: %v", err)
	}

	if response.Error.Code != "INVALID_FILE_TYPE" {
		t.Fatalf("expected INVALID_FILE_TYPE, got %s", response.Error.Code)
	}
}

func createProject(t *testing.T, handler http.Handler) string {
	t.Helper()
	payload := []byte(`{"title":"Upload Target","input_mode":"file"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	var created struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(res.Body).Decode(&created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}

	return created.Data.ID
}
