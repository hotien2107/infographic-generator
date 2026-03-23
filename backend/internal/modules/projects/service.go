package projects

import (
	"context"
	"fmt"
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
	return &Service{
		store:            store,
		blobStorage:      blobStorage,
		processor:        processor,
		autoProcessAfter: autoProcessAfter,
	}
}

func (s *Service) SetProcessor(processor Processor) {
	s.processor = processor
}

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

	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(fileHeader.Filename)), ".")
	now := time.Now().UTC()
	document := documents.Document{
		ID:         utils.NewUUID(),
		ProjectID:  projectID,
		Filename:   firstNonEmpty(strings.TrimSpace(originalFilename), fileHeader.Filename),
		MimeType:   mimeTypeForExtension(ext),
		SizeBytes:  fileHeader.Size,
		StorageKey: storageKey,
		Status:     documents.StatusUploaded,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	project, _, err := s.store.AddDocument(ctx, projectID, document)
	if err != nil {
		return Project{}, documents.Document{}, err
	}

	if s.autoProcessAfter && s.processor != nil {
		go func(doc documents.Document) {
			_, _, _ = s.queueDocumentProcessing(context.Background(), projectID, doc)
		}(document)
	}

	return project, document, nil
}

func (s *Service) TriggerProcessing(ctx context.Context, projectID string) (Project, documents.Document, error) {
	document, err := s.store.GetLatestDocument(ctx, projectID)
	if err != nil {
		return Project{}, documents.Document{}, err
	}
	return s.queueDocumentProcessing(ctx, projectID, document)
}

func (s *Service) queueDocumentProcessing(ctx context.Context, projectID string, document documents.Document) (Project, documents.Document, error) {
	if s.processor == nil {
		return Project{}, documents.Document{}, fmt.Errorf("processor is not configured")
	}
	project, queuedDocument, err := s.store.UpdateDocumentProcessing(ctx, projectID, document.ID, DocumentProcessingUpdate{
		ProjectStatus:        StatusProcessing,
		ProjectStep:          StepQueuedProcessing,
		DocumentStatus:       documents.StatusQueued,
		ErrorMessage:         nil,
		ExtractedTextPreview: nil,
	})
	if err != nil {
		return Project{}, documents.Document{}, err
	}

	queuedDocument.Filename = document.Filename
	queuedDocument.MimeType = document.MimeType
	queuedDocument.SizeBytes = document.SizeBytes

	if err := s.processor.Enqueue(processing.Task{ProjectID: projectID, Document: queuedDocument}); err != nil {
		message := "failed to enqueue document for processing"
		failedProject, failedDocument, updateErr := s.store.UpdateDocumentProcessing(ctx, projectID, document.ID, DocumentProcessingUpdate{
			ProjectStatus:        StatusFailed,
			ProjectStep:          StepFailed,
			DocumentStatus:       documents.StatusFailed,
			ErrorMessage:         &message,
			ProcessingFinishedAt: timePointer(time.Now().UTC()),
		})
		if updateErr == nil {
			return failedProject, failedDocument, err
		}
		return project, queuedDocument, err
	}

	return project, queuedDocument, nil
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
	default:
		return "text/plain"
	}
}

func timePointer(value time.Time) *time.Time {
	return &value
}

func (s *Service) MarkDocumentProcessingStarted(ctx context.Context, projectID, documentID string, startedAt time.Time) error {
	_, _, err := s.store.UpdateDocumentProcessing(ctx, projectID, documentID, DocumentProcessingUpdate{
		ProjectStatus:        StatusProcessing,
		ProjectStep:          StepExtracting,
		DocumentStatus:       documents.StatusProcessing,
		ProcessingStartedAt:  &startedAt,
		ErrorMessage:         nil,
		ExtractedTextPreview: nil,
	})
	return err
}

func (s *Service) MarkDocumentProcessed(ctx context.Context, projectID, documentID string, startedAt, finishedAt time.Time, preview string) error {
	_, _, err := s.store.UpdateDocumentProcessing(ctx, projectID, documentID, DocumentProcessingUpdate{
		ProjectStatus:        StatusProcessed,
		ProjectStep:          StepReadyForGeneration,
		DocumentStatus:       documents.StatusProcessed,
		ProcessingStartedAt:  &startedAt,
		ProcessingFinishedAt: &finishedAt,
		ErrorMessage:         nil,
		ExtractedTextPreview: &preview,
	})
	return err
}

func (s *Service) MarkDocumentFailed(ctx context.Context, projectID, documentID string, startedAt, finishedAt time.Time, message string) error {
	_, _, err := s.store.UpdateDocumentProcessing(ctx, projectID, documentID, DocumentProcessingUpdate{
		ProjectStatus:        StatusFailed,
		ProjectStep:          StepFailed,
		DocumentStatus:       documents.StatusFailed,
		ProcessingStartedAt:  &startedAt,
		ProcessingFinishedAt: &finishedAt,
		ErrorMessage:         &message,
		ExtractedTextPreview: nil,
	})
	return err
}
