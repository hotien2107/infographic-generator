package processing

import (
	"context"
	"fmt"
	"strings"
	"time"

	"infographic-generator/backend/internal/modules/documents"
)

type StateStore interface {
	MarkDocumentProcessingStarted(ctx context.Context, projectID, documentID string, startedAt time.Time) error
	MarkDocumentProcessed(ctx context.Context, projectID, documentID string, startedAt, finishedAt time.Time, preview string) error
	MarkDocumentFailed(ctx context.Context, projectID, documentID string, startedAt, finishedAt time.Time, message string) error
}

type Task struct {
	ProjectID string
	Document  documents.Document
}

type Worker struct {
	store       StateStore
	queue       chan Task
	stepDelay   time.Duration
	failPattern string
}

func NewWorker(store StateStore, queueBuffer int, stepDelay time.Duration, failPattern string) *Worker {
	if queueBuffer <= 0 {
		queueBuffer = 8
	}
	if stepDelay <= 0 {
		stepDelay = 150 * time.Millisecond
	}
	return &Worker{
		store:       store,
		queue:       make(chan Task, queueBuffer),
		stepDelay:   stepDelay,
		failPattern: strings.ToLower(strings.TrimSpace(failPattern)),
	}
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
	startedAt := time.Now().UTC()
	_ = w.store.MarkDocumentProcessingStarted(ctx, task.ProjectID, task.Document.ID, startedAt)

	timer := time.NewTimer(w.stepDelay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return
	case <-timer.C:
	}

	filename := strings.ToLower(task.Document.Filename)
	if w.failPattern != "" && strings.Contains(filename, w.failPattern) {
		finishedAt := time.Now().UTC()
		message := fmt.Sprintf("simulated processing failure for %s", task.Document.Filename)
		_ = w.store.MarkDocumentFailed(ctx, task.ProjectID, task.Document.ID, startedAt, finishedAt, message)
		return
	}

	finishedAt := time.Now().UTC()
	preview := buildPreview(task.Document)
	_ = w.store.MarkDocumentProcessed(ctx, task.ProjectID, task.Document.ID, startedAt, finishedAt, preview)
}

func buildPreview(document documents.Document) string {
	return fmt.Sprintf("Simulated extraction preview for %s (%s, %d bytes).", document.Filename, document.MimeType, document.SizeBytes)
}
