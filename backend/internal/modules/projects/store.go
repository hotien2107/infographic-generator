package projects

import (
	"context"
	"errors"
	"fmt"
	"sort"
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
	GetDashboardSummary(ctx context.Context) (DashboardSummary, error)
	ListProjects(ctx context.Context) ([]ProjectListItem, error)
	CreateProject(ctx context.Context, title, description string, inputMode InputMode) (Project, error)
	GetProject(ctx context.Context, projectID string) (Project, []documents.Document, error)
	UpdateProject(ctx context.Context, projectID string, update ProjectUpdate) (Project, error)
	DeleteProject(ctx context.Context, projectID string) error
	AddDocument(ctx context.Context, projectID string, document documents.Document) (Project, []documents.Document, error)
	ListDocuments(ctx context.Context, projectID string) ([]documents.Document, error)
	UpdateDocument(ctx context.Context, projectID, documentID, filename string) (documents.Document, error)
	DeleteDocument(ctx context.Context, projectID, documentID string) error
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

func (s *PostgresStore) GetDashboardSummary(ctx context.Context) (DashboardSummary, error) {
	projectsList, err := s.ListProjects(ctx)
	if err != nil {
		return DashboardSummary{}, err
	}

	summary := DashboardSummary{TotalProjects: len(projectsList)}
	for _, project := range projectsList {
		summary.TotalDocuments += project.DocumentCount
		switch project.Status {
		case StatusProcessing:
			summary.ProcessingProjects++
		case StatusProcessed:
			summary.CompletedProjects++
		case StatusFailed:
			summary.AttentionProjects++
		case StatusDraft:
			summary.DraftProjects++
		}
	}

	return summary, nil
}

func (s *PostgresStore) ListProjects(ctx context.Context) ([]ProjectListItem, error) {
	query := `
		SELECT p.id, p.title, COALESCE(p.description, ''), p.input_mode, p.status, p.current_step, p.created_at, p.updated_at,
		       COALESCE(COUNT(d.id), 0)
		FROM projects p
		LEFT JOIN documents d ON d.project_id = p.id
		GROUP BY p.id, p.title, p.description, p.input_mode, p.status, p.current_step, p.created_at, p.updated_at
		ORDER BY p.updated_at DESC, p.created_at DESC
	`
	rows, err := s.client.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}

	result := make([]ProjectListItem, 0, len(rows))
	for _, row := range rows {
		item, err := projectListItemFromRow(row)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, nil
}

func (s *PostgresStore) CreateProject(ctx context.Context, title, description string, inputMode InputMode) (Project, error) {
	now := time.Now().UTC()
	project := Project{
		ID:          utils.NewUUID(),
		Title:       title,
		Description: description,
		InputMode:   inputMode,
		Status:      StatusDraft,
		CurrentStep: StepWaitingUpload,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	query, err := postgres.FormatQuery(`
		INSERT INTO projects (id, title, description, input_mode, status, current_step, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, project.ID, project.Title, project.Description, string(project.InputMode), string(project.Status), string(project.CurrentStep), project.CreatedAt, project.UpdatedAt)
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
	project, err := s.projectByID(ctx, projectID)
	if err != nil {
		return Project{}, nil, err
	}

	docs, err := s.documentsByProject(ctx, s.client, projectID)
	if err != nil {
		return Project{}, nil, err
	}

	project.ProcessingSummary = buildProcessingSummary(docs)
	return project, docs, nil
}

func (s *PostgresStore) UpdateProject(ctx context.Context, projectID string, update ProjectUpdate) (Project, error) {
	query, err := postgres.FormatQuery(`
		UPDATE projects
		SET title = COALESCE($1, title),
		    description = COALESCE($2, description),
		    input_mode = COALESCE($3, input_mode),
		    updated_at = $4
		WHERE id = $5
		RETURNING id, title, COALESCE(description, ''), input_mode, status, current_step, created_at, updated_at
	`, nullableValue(update.Title), nullableValue(update.Description), nullableInputMode(update.InputMode), time.Now().UTC(), projectID)
	if err != nil {
		return Project{}, fmt.Errorf("format project update query: %w", err)
	}

	row, err := s.client.QueryRow(ctx, query)
	if err != nil {
		if errors.Is(err, postgres.ErrNoRows) {
			return Project{}, ErrProjectNotFound
		}
		return Project{}, fmt.Errorf("update project: %w", err)
	}

	project, err := projectFromRow(row)
	if err != nil {
		return Project{}, err
	}
	project.ProcessingSummary = buildProcessingSummaryMust(s.documentsByProject(ctx, s.client, projectID))
	return project, nil
}

func (s *PostgresStore) DeleteProject(ctx context.Context, projectID string) error {
	query, err := postgres.FormatQuery(`DELETE FROM projects WHERE id = $1 RETURNING id`, projectID)
	if err != nil {
		return fmt.Errorf("format project delete query: %w", err)
	}
	if _, err := s.client.QueryRow(ctx, query); err != nil {
		if errors.Is(err, postgres.ErrNoRows) {
			return ErrProjectNotFound
		}
		return fmt.Errorf("delete project: %w", err)
	}
	return nil
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
		RETURNING id, title, COALESCE(description, ''), input_mode, status, current_step, created_at, updated_at
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

func (s *PostgresStore) ListDocuments(ctx context.Context, projectID string) ([]documents.Document, error) {
	if _, err := s.projectByID(ctx, projectID); err != nil {
		return nil, err
	}
	return s.documentsByProject(ctx, s.client, projectID)
}

func (s *PostgresStore) UpdateDocument(ctx context.Context, projectID, documentID, filename string) (documents.Document, error) {
	query, err := postgres.FormatQuery(`
		UPDATE documents
		SET filename = $1, updated_at = $2
		WHERE id = $3 AND project_id = $4
		RETURNING id, project_id, filename, mime_type, size_bytes, storage_key, status,
		          processing_started_at, processing_finished_at, error_message, extracted_text_preview,
		          created_at, updated_at
	`, filename, time.Now().UTC(), documentID, projectID)
	if err != nil {
		return documents.Document{}, fmt.Errorf("format document update query: %w", err)
	}

	row, err := s.client.QueryRow(ctx, query)
	if err != nil {
		if errors.Is(err, postgres.ErrNoRows) {
			if _, projectErr := s.projectByID(ctx, projectID); projectErr != nil {
				return documents.Document{}, projectErr
			}
			return documents.Document{}, ErrDocumentNotFound
		}
		return documents.Document{}, fmt.Errorf("update document: %w", err)
	}
	return documentFromRow(row)
}

func (s *PostgresStore) DeleteDocument(ctx context.Context, projectID, documentID string) error {
	tx, err := s.client.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := s.projectByIDWithExecutor(ctx, tx, projectID); err != nil {
		return err
	}

	deleteQuery, err := postgres.FormatQuery(`DELETE FROM documents WHERE id = $1 AND project_id = $2 RETURNING id`, documentID, projectID)
	if err != nil {
		return fmt.Errorf("format document delete query: %w", err)
	}
	if _, err := tx.QueryRow(ctx, deleteQuery); err != nil {
		if errors.Is(err, postgres.ErrNoRows) {
			return ErrDocumentNotFound
		}
		return fmt.Errorf("delete document: %w", err)
	}

	docs, err := s.documentsByProject(ctx, tx, projectID)
	if err != nil {
		return err
	}
	status, step := deriveProjectState(docs)
	updateProjectQuery, err := postgres.FormatQuery(`
		UPDATE projects
		SET status = $1, current_step = $2, updated_at = $3
		WHERE id = $4
	`, string(status), string(step), time.Now().UTC(), projectID)
	if err != nil {
		return fmt.Errorf("format project state sync query: %w", err)
	}
	if err := tx.Exec(ctx, updateProjectQuery); err != nil {
		return fmt.Errorf("sync project state after delete: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
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
			if _, projectErr := s.projectByID(ctx, projectID); projectErr != nil {
				return documents.Document{}, projectErr
			}
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
		RETURNING id, title, COALESCE(description, ''), input_mode, status, current_step, created_at, updated_at
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
			description TEXT NOT NULL DEFAULT '',
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

		ALTER TABLE projects ADD COLUMN IF NOT EXISTS description TEXT NOT NULL DEFAULT '';
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

func (s *PostgresStore) projectByID(ctx context.Context, projectID string) (Project, error) {
	return s.projectByIDWithExecutor(ctx, s.client, projectID)
}

func (s *PostgresStore) projectByIDWithExecutor(ctx context.Context, executor queryRowExecutor, projectID string) (Project, error) {
	projectQuery, err := postgres.FormatQuery(`
		SELECT id, title, COALESCE(description, ''), input_mode, status, current_step, created_at, updated_at
		FROM projects
		WHERE id = $1
	`, projectID)
	if err != nil {
		return Project{}, fmt.Errorf("format project select query: %w", err)
	}
	projectRow, err := executor.QueryRow(ctx, projectQuery)
	if err != nil {
		if errors.Is(err, postgres.ErrNoRows) {
			return Project{}, ErrProjectNotFound
		}
		return Project{}, fmt.Errorf("select project: %w", err)
	}

	project, err := projectFromRow(projectRow)
	if err != nil {
		return Project{}, err
	}
	return project, nil
}

type queryExecutor interface {
	Query(ctx context.Context, query string) ([]postgres.Row, error)
}

type queryRowExecutor interface {
	QueryRow(ctx context.Context, query string) (postgres.Row, error)
}

func projectFromRow(row postgres.Row) (Project, error) {
	if len(row) != 8 {
		return Project{}, fmt.Errorf("unexpected project row column count: %d", len(row))
	}
	createdAt, err := parseTimestamp(row[6])
	if err != nil {
		return Project{}, fmt.Errorf("parse project created_at: %w", err)
	}
	updatedAt, err := parseTimestamp(row[7])
	if err != nil {
		return Project{}, fmt.Errorf("parse project updated_at: %w", err)
	}
	return Project{
		ID:          row[0],
		Title:       row[1],
		Description: row[2],
		InputMode:   InputMode(row[3]),
		Status:      Status(row[4]),
		CurrentStep: Step(row[5]),
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}, nil
}

func projectListItemFromRow(row postgres.Row) (ProjectListItem, error) {
	if len(row) != 9 {
		return ProjectListItem{}, fmt.Errorf("unexpected project list row column count: %d", len(row))
	}
	project, err := projectFromRow(row[:8])
	if err != nil {
		return ProjectListItem{}, err
	}
	documentCount, err := strconv.Atoi(row[8])
	if err != nil {
		return ProjectListItem{}, fmt.Errorf("parse project document_count: %w", err)
	}
	return ProjectListItem{Project: project, DocumentCount: documentCount}, nil
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
	createdAt, err := parseTimestamp(row[11])
	if err != nil {
		return documents.Document{}, fmt.Errorf("parse document created_at: %w", err)
	}
	updatedAt, err := parseTimestamp(row[12])
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
	parsed, err := parseTimestamp(value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func parseTimestamp(value string) (time.Time, error) {
	layouts := []string{
		time.RFC3339Nano,
		"2006-01-02 15:04:05.999999999Z07:00",
		"2006-01-02 15:04:05.999999999Z07",
		"2006-01-02 15:04:05Z07:00",
		"2006-01-02 15:04:05Z07",
	}
	var lastErr error
	for _, layout := range layouts {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed, nil
		}
		lastErr = err
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("unsupported timestamp format")
	}
	return time.Time{}, lastErr
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

func buildProcessingSummaryMust(docs []documents.Document, err error) ProcessingSummary {
	if err != nil {
		return ProcessingSummary{}
	}
	return buildProcessingSummary(docs)
}

func deriveProjectState(docs []documents.Document) (Status, Step) {
	if len(docs) == 0 {
		return StatusDraft, StepWaitingUpload
	}

	summary := buildProcessingSummary(docs)
	switch {
	case summary.ProcessingDocuments > 0:
		return StatusProcessing, StepExtracting
	case summary.QueuedDocuments > 0:
		return StatusProcessing, StepQueuedProcessing
	case summary.FailedDocuments > 0:
		return StatusFailed, StepFailed
	case summary.ProcessedDocuments == summary.TotalDocuments:
		return StatusProcessed, StepReadyForGeneration
	default:
		return StatusUploaded, StepUploaded
	}
}

func nullableValue(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := *value
	return &trimmed
}

func nullableInputMode(value *InputMode) *string {
	if value == nil {
		return nil
	}
	result := string(*value)
	return &result
}

func sortDocumentsNewestFirst(items []documents.Document) []documents.Document {
	result := append([]documents.Document(nil), items...)
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.After(result[j].CreatedAt)
	})
	return result
}
