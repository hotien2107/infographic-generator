package projects

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"infographic-generator/backend/internal/modules/documents"
	"infographic-generator/backend/internal/processing"
	"infographic-generator/backend/internal/storage"
	"infographic-generator/backend/internal/utils"
)

type Processor interface {
	Enqueue(task processing.Task) error
}

type Service struct {
	store            Store
	blobStorage      storage.BlobStorage
	processor        Processor
	autoProcessAfter bool
}

func NewService(store Store, blobStorage storage.BlobStorage, processor Processor, autoProcessAfter bool) *Service {
	return &Service{store: store, blobStorage: blobStorage, processor: processor, autoProcessAfter: autoProcessAfter}
}

func (s *Service) SetProcessor(processor Processor) { s.processor = processor }

func (s *Service) DashboardSummary(ctx context.Context) (DashboardSummary, error) {
	return s.store.GetDashboardSummary(ctx)
}
func (s *Service) ListProjects(ctx context.Context) ([]ProjectListItem, error) {
	return s.store.ListProjects(ctx)
}
func (s *Service) CreateProject(ctx context.Context, title, description string, inputMode InputMode) (Project, error) {
	return s.store.CreateProject(ctx, title, description, inputMode)
}
func (s *Service) GetProject(ctx context.Context, projectID string) (Project, []documents.Document, error) {
	return s.store.GetProject(ctx, projectID)
}
func (s *Service) UpdateProject(ctx context.Context, projectID string, update ProjectUpdate) (Project, error) {
	return s.store.UpdateProject(ctx, projectID, update)
}
func (s *Service) DeleteProject(ctx context.Context, projectID string) error {
	return s.store.DeleteProject(ctx, projectID)
}
func (s *Service) ListDocuments(ctx context.Context, projectID string) ([]documents.Document, error) {
	return s.store.ListDocuments(ctx, projectID)
}
func (s *Service) UpdateDocument(ctx context.Context, projectID, documentID, filename string) (documents.Document, error) {
	return s.store.UpdateDocument(ctx, projectID, documentID, filename)
}
func (s *Service) DeleteDocument(ctx context.Context, projectID, documentID string) error {
	return s.store.DeleteDocument(ctx, projectID, documentID)
}

func (s *Service) UploadDocument(ctx context.Context, projectID, originalFilename string, fileHeader *multipart.FileHeader) (Project, documents.Document, error) {
	storageKey, err := s.blobStorage.Save(ctx, fileHeader)
	if err != nil {
		return Project{}, documents.Document{}, fmt.Errorf("persist upload into object storage: %w", err)
	}
	doc, err := newFileDocument(projectID, originalFilename, fileHeader, storageKey)
	if err != nil {
		return Project{}, documents.Document{}, err
	}

	project, _, err := s.store.AddDocument(ctx, projectID, doc)
	if err != nil {
		return Project{}, documents.Document{}, err
	}
	if s.autoProcessAfter {
		go func() { _, _, _ = s.queueDocumentExtraction(context.Background(), projectID, doc) }()
	}
	return project, doc, nil
}

func (s *Service) SubmitText(ctx context.Context, projectID, rawText string) (Project, documents.Document, error) {
	trimmed := strings.TrimSpace(rawText)
	if trimmed == "" {
		return Project{}, documents.Document{}, fmt.Errorf("raw text must not be empty")
	}
	now := time.Now().UTC()
	doc := documents.Document{
		ID:         utils.NewUUID(),
		ProjectID:  projectID,
		Filename:   fmt.Sprintf("manual-input-%d.txt", now.Unix()),
		MimeType:   "text/plain",
		SizeBytes:  int64(len(trimmed)),
		StorageKey: "",
		SourceType: documents.SourceTypeText,
		FileType:   documents.FileTypeText,
		Status:     documents.StatusUploaded,
		RawText:    &trimmed,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	project, _, err := s.store.AddDocument(ctx, projectID, doc)
	if err != nil {
		return Project{}, documents.Document{}, err
	}
	if s.autoProcessAfter {
		go func() { _, _, _ = s.queueDocumentExtraction(context.Background(), projectID, doc) }()
	}
	return project, doc, nil
}

func (s *Service) TriggerProcessing(ctx context.Context, projectID string) (Project, documents.Document, error) {
	document, err := s.store.GetLatestDocument(ctx, projectID)
	if err != nil {
		return Project{}, documents.Document{}, err
	}
	return s.queueDocumentExtraction(ctx, projectID, document)
}

func (s *Service) queueDocumentExtraction(ctx context.Context, projectID string, document documents.Document) (Project, documents.Document, error) {
	if s.processor == nil {
		return Project{}, documents.Document{}, fmt.Errorf("processor is not configured")
	}
	log.Printf("[extract][queue] project=%s document=%s", projectID, document.ID)
	project, queuedDocument, err := s.store.UpdateDocumentProcessing(ctx, projectID, document.ID, DocumentProcessingUpdate{
		ProjectStatus: StatusExtracting, ProjectStep: StepQueuedForExtract, DocumentStatus: documents.StatusUploaded,
		ErrorMessage: nil,
	})
	if err != nil {
		return Project{}, documents.Document{}, err
	}
	if err := s.processor.Enqueue(processing.Task{ProjectID: projectID, Document: queuedDocument}); err != nil {
		message := "failed to enqueue extraction job"
		failedProject, failedDocument, updateErr := s.store.UpdateDocumentProcessing(ctx, projectID, document.ID, DocumentProcessingUpdate{
			ProjectStatus: StatusFailed, ProjectStep: StepFailed, DocumentStatus: documents.StatusFailed,
			ErrorMessage: &message, ExtractionEndedAt: timePointer(time.Now().UTC()),
		})
		if updateErr == nil {
			return failedProject, failedDocument, err
		}
		return project, queuedDocument, err
	}
	return project, queuedDocument, nil
}

func newFileDocument(projectID, originalFilename string, fileHeader *multipart.FileHeader, storageKey string) (documents.Document, error) {
	fileType, mimeType, err := fileTypeFromFilename(fileHeader.Filename)
	if err != nil {
		return documents.Document{}, err
	}
	now := time.Now().UTC()
	return documents.Document{
		ID:         utils.NewUUID(),
		ProjectID:  projectID,
		Filename:   firstNonEmpty(strings.TrimSpace(originalFilename), fileHeader.Filename),
		MimeType:   mimeType,
		SizeBytes:  fileHeader.Size,
		StorageKey: storageKey,
		SourceType: documents.SourceTypeFile,
		FileType:   fileType,
		Status:     documents.StatusUploaded,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}

func fileTypeFromFilename(filename string) (documents.FileType, string, error) {
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(filename)), ".")
	switch ext {
	case "pdf":
		return documents.FileTypePDF, "application/pdf", nil
	case "txt":
		return documents.FileTypeTXT, "text/plain", nil
	default:
		return "", "", fmt.Errorf("unsupported file extension: %s", ext)
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func timePointer(value time.Time) *time.Time { return &value }

func (s *Service) MarkDocumentExtractionStarted(ctx context.Context, projectID, documentID string, startedAt time.Time) error {
	_, _, err := s.store.UpdateDocumentProcessing(ctx, projectID, documentID, DocumentProcessingUpdate{
		ProjectStatus: StatusExtracting, ProjectStep: StepExtracting, DocumentStatus: documents.StatusExtracting,
		ExtractionStartedAt: &startedAt, ErrorMessage: nil,
	})
	return err
}

func (s *Service) MarkDocumentExtracted(ctx context.Context, projectID, documentID string, startedAt, endedAt time.Time, rawText string, metadata documents.RawContentMetadata) error {
	_, _, err := s.store.UpdateDocumentProcessing(ctx, projectID, documentID, DocumentProcessingUpdate{
		ProjectStatus: StatusExtracted, ProjectStep: StepReadyForGeneration, DocumentStatus: documents.StatusExtracted,
		ExtractionStartedAt: &startedAt, ExtractionEndedAt: &endedAt, RawText: &rawText, Metadata: &metadata,
	})
	return err
}

func (s *Service) MarkDocumentExtractionFailed(ctx context.Context, projectID, documentID string, startedAt, endedAt time.Time, message string) error {
	_, _, err := s.store.UpdateDocumentProcessing(ctx, projectID, documentID, DocumentProcessingUpdate{
		ProjectStatus: StatusFailed, ProjectStep: StepFailed, DocumentStatus: documents.StatusFailed,
		ExtractionStartedAt: &startedAt, ExtractionEndedAt: &endedAt, ErrorMessage: &message,
	})
	return err
}

func (s *Service) LoadDocumentPayload(ctx context.Context, storageKey string) ([]byte, error) {
	return s.blobStorage.Read(ctx, storageKey)
}
