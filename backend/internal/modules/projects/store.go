package projects

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"infographic-generator/backend/internal/modules/documents"
	"infographic-generator/backend/internal/platform/postgres"
	"infographic-generator/backend/internal/utils"
)

var (
	ErrProjectNotFound       = errors.New("project not found")
	ErrDocumentNotFound      = errors.New("document not found")
	ErrNoDocumentsForProject = errors.New("project has no documents")
)

type Store interface {
	CreateProject(ctx context.Context, title string, inputMode InputMode) (Project, error)
	GetProject(ctx context.Context, projectID string) (Project, []documents.Document, error)
	AddDocument(ctx context.Context, projectID string, document documents.Document) (Project, []documents.Document, error)
	GetLatestDocument(ctx context.Context, projectID string) (documents.Document, error)
	UpdateDocumentProcessing(ctx context.Context, projectID, documentID string, params DocumentProcessingUpdate) (Project, documents.Document, error)
	Close()
}

type DocumentProcessingUpdate struct {
	ProjectStatus         Status
	ProjectStep           Step
	DocumentStatus        documents.Status
	ProcessingStartedAt   *time.Time
	ProcessingFinishedAt  *time.Time
	ErrorMessage          *string
	ExtractedTextPreview  *string
	TouchProjectUpdatedAt bool
}

type PostgresStore struct {
	client *postgres.Client
}

func NewPostgresStore(ctx context.Context, databaseURL string) (*PostgresStore, error) {
	client, err := postgres.NewClient(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("create postgres client: %w", err)
	}

	store := &PostgresStore{client: client}
	if err := store.initSchema(ctx); err != nil {
		return nil, err
	}

	return store, nil
}

func (s *PostgresStore) CreateProject(ctx context.Context, title string, inputMode InputMode) (Project, error) {
	now := time.Now().UTC()
	project := Project{
		ID:          utils.NewUUID(),
		Title:       title,
		InputMode:   inputMode,
		Status:      StatusDraft,
		CurrentStep: StepWaitingUpload,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	query, err := postgres.FormatQuery(`
		INSERT INTO projects (id, title, input_mode, status, current_step, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, project.ID, project.Title, string(project.InputMode), string(project.Status), string(project.CurrentStep), project.CreatedAt, project.UpdatedAt)
	if err != nil {
		return Project{}, fmt.Errorf("format project insert query: %w", err)
	}
	if err := s.client.Exec(ctx, query); err != nil {
		return Project{}, fmt.Errorf("insert project: %w", err)
	}

	project.ProcessingSummary = ProcessingSummary{}
	return project, nil
}

func (s *PostgresStore) GetProject(ctx context.Context, projectID string) (Project, []documents.Document, error) {
	projectQuery, err := postgres.FormatQuery(`
		SELECT id, title, input_mode, status, current_step, created_at, updated_at
		FROM projects
		WHERE id = $1
	`, projectID)
	if err != nil {
		return Project{}, nil, fmt.Errorf("format project select query: %w", err)
	}
	projectRow, err := s.client.QueryRow(ctx, projectQuery)
	if err != nil {
		if errors.Is(err, postgres.ErrNoRows) {
			return Project{}, nil, ErrProjectNotFound
		}
		return Project{}, nil, fmt.Errorf("select project: %w", err)
	}

	project, err := projectFromRow(projectRow)
	if err != nil {
		return Project{}, nil, err
	}

	docQuery, err := postgres.FormatQuery(`
		SELECT id, project_id, filename, mime_type, size_bytes, storage_key, status,
		       processing_started_at, processing_finished_at, error_message, extracted_text_preview,
		       created_at, updated_at
		FROM documents
		WHERE project_id = $1
		ORDER BY created_at ASC
	`, projectID)
	if err != nil {
		return Project{}, nil, fmt.Errorf("format documents select query: %w", err)
	}
	docRows, err := s.client.Query(ctx, docQuery)
	if err != nil {
		return Project{}, nil, fmt.Errorf("select documents: %w", err)
	}

	docs := make([]documents.Document, 0, len(docRows))
	for _, row := range docRows {
		doc, err := documentFromRow(row)
		if err != nil {
			return Project{}, nil, err
		}
		docs = append(docs, doc)
	}

	project.ProcessingSummary = buildProcessingSummary(docs)
	return project, docs, nil
}

func (s *PostgresStore) AddDocument(ctx context.Context, projectID string, document documents.Document) (Project, []documents.Document, error) {
	tx, err := s.client.Begin(ctx)
	if err != nil {
		return Project{}, nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	now := time.Now().UTC()
	updateProjectQuery, err := postgres.FormatQuery(`
		UPDATE projects
		SET status = $1, current_step = $2, updated_at = $3
		WHERE id = $4
		RETURNING id, title, input_mode, status, current_step, created_at, updated_at
	`, string(StatusUploaded), string(StepUploaded), now, projectID)
	if err != nil {
		return Project{}, nil, fmt.Errorf("format project update query: %w", err)
	}
	projectRow, err := tx.QueryRow(ctx, updateProjectQuery)
	if err != nil {
		if errors.Is(err, postgres.ErrNoRows) {
			return Project{}, nil, ErrProjectNotFound
		}
		return Project{}, nil, fmt.Errorf("update project: %w", err)
	}

	project, err := projectFromRow(projectRow)
	if err != nil {
		return Project{}, nil, err
	}

	insertDocumentQuery, err := postgres.FormatQuery(`
		INSERT INTO documents (
			id, project_id, filename, mime_type, size_bytes, storage_key, status,
			processing_started_at, processing_finished_at, error_message, extracted_text_preview,
			created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`,
		document.ID,
		document.ProjectID,
		document.Filename,
		document.MimeType,
		document.SizeBytes,
		document.StorageKey,
		string(document.Status),
		document.ProcessingStartedAt,
		document.ProcessingFinishedAt,
		document.ErrorMessage,
		document.ExtractedTextPreview,
		document.CreatedAt,
		document.UpdatedAt,
	)
	if err != nil {
		return Project{}, nil, fmt.Errorf("format document insert query: %w", err)
	}
	if err := tx.Exec(ctx, insertDocumentQuery); err != nil {
		return Project{}, nil, fmt.Errorf("insert document: %w", err)
	}

	docs, err := s.documentsByProject(ctx, tx, projectID)
	if err != nil {
		return Project{}, nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return Project{}, nil, fmt.Errorf("commit transaction: %w", err)
	}

	project.ProcessingSummary = buildProcessingSummary(docs)
	return project, docs, nil
}

func (s *PostgresStore) GetLatestDocument(ctx context.Context, projectID string) (documents.Document, error) {
	query, err := postgres.FormatQuery(`
		SELECT id, project_id, filename, mime_type, size_bytes, storage_key, status,
		       processing_started_at, processing_finished_at, error_message, extracted_text_preview,
		       created_at, updated_at
		FROM documents
		WHERE project_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`, projectID)
	if err != nil {
		return documents.Document{}, fmt.Errorf("format latest document query: %w", err)
	}
	row, err := s.client.QueryRow(ctx, query)
	if err != nil {
		if errors.Is(err, postgres.ErrNoRows) {
			return documents.Document{}, ErrNoDocumentsForProject
		}
		return documents.Document{}, fmt.Errorf("select latest document: %w", err)
	}
	return documentFromRow(row)
}

func (s *PostgresStore) UpdateDocumentProcessing(ctx context.Context, projectID, documentID string, params DocumentProcessingUpdate) (Project, documents.Document, error) {
	tx, err := s.client.Begin(ctx)
	if err != nil {
		return Project{}, documents.Document{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	now := time.Now().UTC()
	updateProjectQuery, err := postgres.FormatQuery(`
		UPDATE projects
		SET status = $1, current_step = $2, updated_at = $3
		WHERE id = $4
		RETURNING id, title, input_mode, status, current_step, created_at, updated_at
	`, string(params.ProjectStatus), string(params.ProjectStep), now, projectID)
	if err != nil {
		return Project{}, documents.Document{}, fmt.Errorf("format project processing update query: %w", err)
	}
	projectRow, err := tx.QueryRow(ctx, updateProjectQuery)
	if err != nil {
		if errors.Is(err, postgres.ErrNoRows) {
			return Project{}, documents.Document{}, ErrProjectNotFound
		}
		return Project{}, documents.Document{}, fmt.Errorf("update project processing state: %w", err)
	}

	project, err := projectFromRow(projectRow)
	if err != nil {
		return Project{}, documents.Document{}, err
	}

	updateDocumentQuery, err := postgres.FormatQuery(`
		UPDATE documents
		SET status = $1,
		    processing_started_at = $2,
		    processing_finished_at = $3,
		    error_message = $4,
		    extracted_text_preview = $5,
		    updated_at = $6
		WHERE id = $7 AND project_id = $8
		RETURNING id, project_id, filename, mime_type, size_bytes, storage_key, status,
		          processing_started_at, processing_finished_at, error_message, extracted_text_preview,
		          created_at, updated_at
	`, string(params.DocumentStatus), params.ProcessingStartedAt, params.ProcessingFinishedAt, params.ErrorMessage, params.ExtractedTextPreview, now, documentID, projectID)
	if err != nil {
		return Project{}, documents.Document{}, fmt.Errorf("format document processing update query: %w", err)
	}
	documentRow, err := tx.QueryRow(ctx, updateDocumentQuery)
	if err != nil {
		if errors.Is(err, postgres.ErrNoRows) {
			return Project{}, documents.Document{}, ErrDocumentNotFound
		}
		return Project{}, documents.Document{}, fmt.Errorf("update document processing state: %w", err)
	}

	document, err := documentFromRow(documentRow)
	if err != nil {
		return Project{}, documents.Document{}, err
	}

	docs, err := s.documentsByProject(ctx, tx, projectID)
	if err != nil {
		return Project{}, documents.Document{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return Project{}, documents.Document{}, fmt.Errorf("commit transaction: %w", err)
	}

	project.ProcessingSummary = buildProcessingSummary(docs)
	return project, document, nil
}

func (s *PostgresStore) Close() {}

func (s *PostgresStore) initSchema(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS projects (
			id UUID PRIMARY KEY,
			title VARCHAR(120) NOT NULL,
			input_mode TEXT NOT NULL,
			status TEXT NOT NULL,
			current_step TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL,
			updated_at TIMESTAMPTZ NOT NULL
		);

		CREATE TABLE IF NOT EXISTS documents (
			id UUID PRIMARY KEY,
			project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
			filename TEXT NOT NULL,
			mime_type TEXT NOT NULL,
			size_bytes BIGINT NOT NULL,
			storage_key TEXT NOT NULL,
			status TEXT NOT NULL,
			processing_started_at TIMESTAMPTZ NULL,
			processing_finished_at TIMESTAMPTZ NULL,
			error_message TEXT NULL,
			extracted_text_preview TEXT NULL,
			created_at TIMESTAMPTZ NOT NULL,
			updated_at TIMESTAMPTZ NOT NULL
		);

		ALTER TABLE documents ADD COLUMN IF NOT EXISTS processing_started_at TIMESTAMPTZ NULL;
		ALTER TABLE documents ADD COLUMN IF NOT EXISTS processing_finished_at TIMESTAMPTZ NULL;
		ALTER TABLE documents ADD COLUMN IF NOT EXISTS error_message TEXT NULL;
		ALTER TABLE documents ADD COLUMN IF NOT EXISTS extracted_text_preview TEXT NULL;
		ALTER TABLE documents ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

		CREATE INDEX IF NOT EXISTS idx_documents_project_id_created_at
		ON documents(project_id, created_at);
	`
	if err := s.client.Exec(ctx, query); err != nil {
		return fmt.Errorf("initialize postgres schema: %w", err)
	}
	return nil
}

func (s *PostgresStore) documentsByProject(ctx context.Context, executor queryExecutor, projectID string) ([]documents.Document, error) {
	query, err := postgres.FormatQuery(`
		SELECT id, project_id, filename, mime_type, size_bytes, storage_key, status,
		       processing_started_at, processing_finished_at, error_message, extracted_text_preview,
		       created_at, updated_at
		FROM documents
		WHERE project_id = $1
		ORDER BY created_at ASC
	`, projectID)
	if err != nil {
		return nil, fmt.Errorf("format document list query: %w", err)
	}
	rows, err := executor.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("select documents in transaction: %w", err)
	}
	result := make([]documents.Document, 0, len(rows))
	for _, row := range rows {
		document, err := documentFromRow(row)
		if err != nil {
			return nil, err
		}
		result = append(result, document)
	}
	return result, nil
}

type queryExecutor interface {
	Query(ctx context.Context, query string) ([]postgres.Row, error)
}

func projectFromRow(row postgres.Row) (Project, error) {
	if len(row) != 7 {
		return Project{}, fmt.Errorf("unexpected project row column count: %d", len(row))
	}
	createdAt, err := time.Parse(time.RFC3339Nano, row[5])
	if err != nil {
		return Project{}, fmt.Errorf("parse project created_at: %w", err)
	}
	updatedAt, err := time.Parse(time.RFC3339Nano, row[6])
	if err != nil {
		return Project{}, fmt.Errorf("parse project updated_at: %w", err)
	}
	return Project{
		ID:          row[0],
		Title:       row[1],
		InputMode:   InputMode(row[2]),
		Status:      Status(row[3]),
		CurrentStep: Step(row[4]),
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}, nil
}

func documentFromRow(row postgres.Row) (documents.Document, error) {
	if len(row) != 13 {
		return documents.Document{}, fmt.Errorf("unexpected document row column count: %d", len(row))
	}
	sizeBytes, err := strconv.ParseInt(row[4], 10, 64)
	if err != nil {
		return documents.Document{}, fmt.Errorf("parse document size_bytes: %w", err)
	}
	processingStartedAt, err := nullableTime(row[7])
	if err != nil {
		return documents.Document{}, fmt.Errorf("parse document processing_started_at: %w", err)
	}
	processingFinishedAt, err := nullableTime(row[8])
	if err != nil {
		return documents.Document{}, fmt.Errorf("parse document processing_finished_at: %w", err)
	}
	createdAt, err := time.Parse(time.RFC3339Nano, row[11])
	if err != nil {
		return documents.Document{}, fmt.Errorf("parse document created_at: %w", err)
	}
	updatedAt, err := time.Parse(time.RFC3339Nano, row[12])
	if err != nil {
		return documents.Document{}, fmt.Errorf("parse document updated_at: %w", err)
	}
	return documents.Document{
		ID:                   row[0],
		ProjectID:            row[1],
		Filename:             row[2],
		MimeType:             row[3],
		SizeBytes:            sizeBytes,
		StorageKey:           row[5],
		Status:               documents.Status(row[6]),
		ProcessingStartedAt:  processingStartedAt,
		ProcessingFinishedAt: processingFinishedAt,
		ErrorMessage:         nullableString(row[9]),
		ExtractedTextPreview: nullableString(row[10]),
		CreatedAt:            createdAt,
		UpdatedAt:            updatedAt,
	}, nil
}

func nullableTime(value string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func nullableString(value string) *string {
	if value == "" {
		return nil
	}
	result := value
	return &result
}

func buildProcessingSummary(docs []documents.Document) ProcessingSummary {
	summary := ProcessingSummary{TotalDocuments: len(docs)}
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
			if doc.ProcessingFinishedAt != nil && (summary.LastProcessedAt == nil || doc.ProcessingFinishedAt.After(*summary.LastProcessedAt)) {
				timeCopy := *doc.ProcessingFinishedAt
				summary.LastProcessedAt = &timeCopy
			}
		case documents.StatusFailed:
			summary.FailedDocuments++
			if doc.ErrorMessage != nil {
				messageCopy := *doc.ErrorMessage
				summary.LastError = &messageCopy
			}
		}
	}
	return summary
}
