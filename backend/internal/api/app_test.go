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
	return &memoryStore{projects: map[string]projects.Project{}, documents: map[string][]documents.Document{}}
}
func (s *memoryStore) GetDashboardSummary(_ context.Context) (projects.DashboardSummary, error) {
	return projects.DashboardSummary{}, nil
}
func (s *memoryStore) ListProjects(_ context.Context) ([]projects.ProjectListItem, error) {
	return []projects.ProjectListItem{}, nil
}
func (s *memoryStore) CreateProject(_ context.Context, title, description string, inputMode projects.InputMode) (projects.Project, error) {
	now := time.Now().UTC()
	p := projects.Project{ID: utils.NewUUID(), Title: title, Description: description, InputMode: inputMode, Status: projects.StatusDraft, CurrentStep: projects.StepWaitingUpload, CreatedAt: now, UpdatedAt: now}
	s.projects[p.ID] = p
	return p, nil
}
func (s *memoryStore) GetProject(_ context.Context, id string) (projects.Project, []documents.Document, error) {
	p, ok := s.projects[id]
	if !ok {
		return projects.Project{}, nil, projects.ErrProjectNotFound
	}
	docs := s.documents[id]
	p.ProcessingSummary = buildSummary(docs)
	return p, docs, nil
}
func (s *memoryStore) UpdateProject(_ context.Context, id string, u projects.ProjectUpdate) (projects.Project, error) {
	return s.projects[id], nil
}
func (s *memoryStore) DeleteProject(_ context.Context, id string) error {
	delete(s.projects, id)
	delete(s.documents, id)
	return nil
}
func (s *memoryStore) AddDocument(_ context.Context, projectID string, doc documents.Document) (projects.Project, []documents.Document, error) {
	p := s.projects[projectID]
	p.Status = projects.StatusUploaded
	p.CurrentStep = projects.StepUploaded
	s.projects[projectID] = p
	s.documents[projectID] = append(s.documents[projectID], doc)
	p.ProcessingSummary = buildSummary(s.documents[projectID])
	return p, s.documents[projectID], nil
}
func (s *memoryStore) ListDocuments(_ context.Context, projectID string) ([]documents.Document, error) {
	return s.documents[projectID], nil
}
func (s *memoryStore) UpdateDocument(_ context.Context, projectID, documentID, filename string) (documents.Document, error) {
	return documents.Document{}, projects.ErrDocumentNotFound
}
func (s *memoryStore) DeleteDocument(_ context.Context, projectID, documentID string) error {
	return nil
}
func (s *memoryStore) GetLatestDocument(_ context.Context, projectID string) (documents.Document, error) {
	docs := s.documents[projectID]
	if len(docs) == 0 {
		return documents.Document{}, projects.ErrNoDocumentsForProject
	}
	return docs[len(docs)-1], nil
}
func (s *memoryStore) UpdateDocumentProcessing(_ context.Context, projectID, documentID string, params projects.DocumentProcessingUpdate) (projects.Project, documents.Document, error) {
	p := s.projects[projectID]
	p.Status = params.ProjectStatus
	p.CurrentStep = params.ProjectStep
	s.projects[projectID] = p
	docs := s.documents[projectID]
	for i, d := range docs {
		if d.ID == documentID {
			d.Status = params.DocumentStatus
			d.RawText = params.RawText
			d.Metadata = params.Metadata
			d.ExtractionStartedAt = params.ExtractionStartedAt
			d.ExtractionEndedAt = params.ExtractionEndedAt
			d.ErrorMessage = params.ErrorMessage
			docs[i] = d
			s.documents[projectID] = docs
			p.ProcessingSummary = buildSummary(docs)
			return p, d, nil
		}
	}
	return projects.Project{}, documents.Document{}, projects.ErrDocumentNotFound
}
func (s *memoryStore) Close() {}

type fakeBlobStorage struct{ payload map[string][]byte }

func (s *fakeBlobStorage) Save(_ context.Context, f *multipart.FileHeader) (string, error) {
	key := filepath.Join("documents", f.Filename)
	file, _ := f.Open()
	defer file.Close()
	b := new(bytes.Buffer)
	_, _ = b.ReadFrom(file)
	if s.payload == nil {
		s.payload = map[string][]byte{}
	}
	s.payload[key] = b.Bytes()
	return key, nil
}
func (s *fakeBlobStorage) Read(_ context.Context, key string) ([]byte, error) {
	return s.payload[key], nil
}
func (s *fakeBlobStorage) Close() error { return nil }

func newTestHandler(t *testing.T) http.Handler {
	store := newMemoryStore()
	blob := &fakeBlobStorage{payload: map[string][]byte{}}
	service := projects.NewService(store, blob, nil, true)
	worker := processing.NewWorker(service, 8)
	service.SetProcessor(worker)
	worker.Start(context.Background())
	app := api.New(config.Config{MaxUploadSizeMB: 1, AllowedFileTypes: []string{"pdf", "txt"}}, store, blob, service)
	return app.Handler()
}

func TestCreateProjectWithTextMode(t *testing.T) {
	h := newTestHandler(t)
	res := performJSONRequest(t, h, http.MethodPost, "/api/v1/projects", []byte(`{"title":"text project","input_mode":"text"}`))
	if res.Code != http.StatusCreated {
		t.Fatalf("%d", res.Code)
	}
}
func TestSubmitTextAndExtractSuccess(t *testing.T) {
	h := newTestHandler(t)
	pid := createProject(t, h, "text", "text")
	res := performJSONRequest(t, h, http.MethodPost, "/api/v1/projects/"+pid+"/text", []byte(`{"raw_text":"SECTION A\nXin chao sprint 2 extraction"}`))
	if res.Code != http.StatusAccepted {
		t.Fatalf("%d", res.Code)
	}
	time.Sleep(80 * time.Millisecond)
	detail := performJSONRequest(t, h, http.MethodGet, "/api/v1/projects/"+pid, nil)
	if detail.Code != http.StatusOK {
		t.Fatalf("%d", detail.Code)
	}
}
func TestUploadTxtAndExtractSuccess(t *testing.T) {
	h := newTestHandler(t)
	pid := createProject(t, h, "file", "file")
	res := uploadDocument(t, h, pid, "sample.txt", []byte("TITLE\nhello world"))
	if res.Code != http.StatusAccepted {
		t.Fatalf("%d", res.Code)
	}
	time.Sleep(80 * time.Millisecond)
}
func TestValidationForInvalidTextPayload(t *testing.T) {
	h := newTestHandler(t)
	pid := createProject(t, h, "file", "text")
	res := performJSONRequest(t, h, http.MethodPost, "/api/v1/projects/"+pid+"/text", []byte(`{"raw_text":"short"}`))
	if res.Code != http.StatusBadRequest {
		t.Fatalf("%d", res.Code)
	}
}

func createProject(t *testing.T, h http.Handler, title, mode string) string {
	res := performJSONRequest(t, h, http.MethodPost, "/api/v1/projects", []byte(`{"title":"`+title+`","input_mode":"`+mode+`"}`))
	var v struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	_ = json.NewDecoder(strings.NewReader(res.Body.String())).Decode(&v)
	return v.Data.ID
}
func uploadDocument(t *testing.T, h http.Handler, projectID, filename string, content []byte) *httptest.ResponseRecorder {
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	fw, _ := w.CreateFormFile("file", filename)
	_, _ = fw.Write(content)
	_ = w.Close()
	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/"+projectID+"/documents", &body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	h.ServeHTTP(res, req)
	return res
}
func performJSONRequest(t *testing.T, h http.Handler, method, path string, body []byte) *httptest.ResponseRecorder {
	res := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	h.ServeHTTP(res, req)
	return res
}
func buildSummary(docs []documents.Document) projects.ProcessingSummary {
	s := projects.ProcessingSummary{TotalDocuments: len(docs)}
	for _, d := range docs {
		switch d.Status {
		case documents.StatusUploaded:
			s.UploadedDocuments++
		case documents.StatusExtracting:
			s.ExtractingDocuments++
		case documents.StatusExtracted:
			s.ExtractedDocuments++
		case documents.StatusFailed:
			s.FailedDocuments++
		}
	}
	return s
}
