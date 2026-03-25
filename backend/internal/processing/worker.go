package processing

import (
	"context"
	"fmt"
	"log"
	"time"

	"infographic-generator/backend/internal/extraction"
	"infographic-generator/backend/internal/modules/documents"
)

type StateStore interface {
	MarkDocumentExtractionStarted(ctx context.Context, projectID, documentID string, startedAt time.Time) error
	MarkDocumentExtracted(ctx context.Context, projectID, documentID string, startedAt, endedAt time.Time, rawText string, metadata documents.RawContentMetadata) error
	MarkDocumentExtractionFailed(ctx context.Context, projectID, documentID string, startedAt, endedAt time.Time, message string) error
	LoadDocumentPayload(ctx context.Context, storageKey string) ([]byte, error)
}

type Task struct {
	ProjectID string
	Document  documents.Document
}

type Worker struct {
	store     StateStore
	queue     chan Task
	extractor *extraction.Service
}

func NewWorker(store StateStore, queueBuffer int) *Worker {
	if queueBuffer <= 0 {
		queueBuffer = 8
	}
	return &Worker{store: store, queue: make(chan Task, queueBuffer), extractor: extraction.NewService()}
}

func (w *Worker) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case task := <-w.queue:
				w.processTask(ctx, task)
			}
		}
	}()
}

func (w *Worker) Enqueue(task Task) error {
	select {
	case w.queue <- task:
		return nil
	default:
		return fmt.Errorf("processing queue is full")
	}
}

func (w *Worker) processTask(ctx context.Context, task Task) {
	log.Printf("[extract][start] project=%s document=%s", task.ProjectID, task.Document.ID)
	startedAt := time.Now().UTC()
	_ = w.store.MarkDocumentExtractionStarted(ctx, task.ProjectID, task.Document.ID, startedAt)

	result, err := w.extract(task, ctx)
	if err != nil {
		endedAt := time.Now().UTC()
		log.Printf("[extract][failed] project=%s document=%s err=%v", task.ProjectID, task.Document.ID, err)
		_ = w.store.MarkDocumentExtractionFailed(ctx, task.ProjectID, task.Document.ID, startedAt, endedAt, err.Error())
		return
	}
	result.Metadata.ExtractedAt = time.Now().UTC()
	endedAt := time.Now().UTC()
	if err := w.store.MarkDocumentExtracted(ctx, task.ProjectID, task.Document.ID, startedAt, endedAt, result.RawText, result.Metadata); err != nil {
		log.Printf("[extract][failed-to-save] project=%s document=%s err=%v", task.ProjectID, task.Document.ID, err)
		return
	}
	log.Printf("[extract][success] project=%s document=%s chars=%d", task.ProjectID, task.Document.ID, len(result.RawText))
}

func (w *Worker) extract(task Task, ctx context.Context) (extraction.Result, error) {
	if task.Document.SourceType == documents.SourceTypeText {
		if task.Document.RawText == nil {
			return extraction.Result{}, fmt.Errorf("missing raw text for text document")
		}
		return w.extractor.ExtractFromText(*task.Document.RawText)
	}
	payload, err := w.store.LoadDocumentPayload(ctx, task.Document.StorageKey)
	if err != nil {
		return extraction.Result{}, err
	}
	return w.extractor.ExtractFromFile(task.Document.FileType, payload)
}
