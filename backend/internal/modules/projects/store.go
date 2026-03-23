package projects

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"infographic-generator/backend/internal/modules/documents"
	"infographic-generator/backend/internal/platform/postgres"
	"infographic-generator/backend/internal/utils"
)

var ErrProjectNotFound = errors.New("project not found")

type Store interface {
	CreateProject(ctx context.Context, title string, inputMode InputMode) (Project, error)
	GetProject(ctx context.Context, projectID string) (Project, []documents.Document, error)
	AddDocument(ctx context.Context, projectID string, document documents.Document) (Project, []documents.Document, error)
	Close()
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

	query := fmt.Sprintf(`
		INSERT INTO projects (id, title, input_mode, status, current_step, created_at, updated_at)
		VALUES (%s, %s, %s, %s, %s, %s, %s)
	`, sqlString(project.ID), sqlString(project.Title), sqlString(string(project.InputMode)), sqlString(string(project.Status)), sqlString(string(project.CurrentStep)), sqlTime(project.CreatedAt), sqlTime(project.UpdatedAt))
	if err := s.client.Exec(ctx, query); err != nil {
		return Project{}, fmt.Errorf("insert project: %w", err)
	}

	return project, nil
}

func (s *PostgresStore) GetProject(ctx context.Context, projectID string) (Project, []documents.Document, error) {
	projectRow, err := s.client.QueryRow(ctx, fmt.Sprintf(`
		SELECT id, title, input_mode, status, current_step, created_at, updated_at
		FROM projects
		WHERE id = %s
	`, sqlString(projectID)))
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

	docRows, err := s.client.Query(ctx, fmt.Sprintf(`
		SELECT id, project_id, filename, mime_type, size_bytes, storage_key, status, created_at
		FROM documents
		WHERE project_id = %s
		ORDER BY created_at ASC
	`, sqlString(projectID)))
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

	return project, docs, nil
}

func (s *PostgresStore) AddDocument(ctx context.Context, projectID string, document documents.Document) (Project, []documents.Document, error) {
	tx, err := s.client.Begin(ctx)
	if err != nil {
		return Project{}, nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	now := time.Now().UTC()
	projectRow, err := tx.QueryRow(ctx, fmt.Sprintf(`
		UPDATE projects
		SET status = %s, current_step = %s, updated_at = %s
		WHERE id = %s
		RETURNING id, title, input_mode, status, current_step, created_at, updated_at
	`, sqlString(string(StatusUploaded)), sqlString(string(StepUploaded)), sqlTime(now), sqlString(projectID)))
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

	insertDocumentQuery := fmt.Sprintf(`
		INSERT INTO documents (id, project_id, filename, mime_type, size_bytes, storage_key, status, created_at)
		VALUES (%s, %s, %s, %s, %d, %s, %s, %s)
	`, sqlString(document.ID), sqlString(document.ProjectID), sqlString(document.Filename), sqlString(document.MimeType), document.SizeBytes, sqlString(document.StorageKey), sqlString(string(document.Status)), sqlTime(document.CreatedAt))
	if err := tx.Exec(ctx, insertDocumentQuery); err != nil {
		return Project{}, nil, fmt.Errorf("insert document: %w", err)
	}

	docRows, err := tx.Query(ctx, fmt.Sprintf(`
		SELECT id, project_id, filename, mime_type, size_bytes, storage_key, status, created_at
		FROM documents
		WHERE project_id = %s
		ORDER BY created_at ASC
	`, sqlString(projectID)))
	if err != nil {
		return Project{}, nil, fmt.Errorf("select documents in transaction: %w", err)
	}

	docs := make([]documents.Document, 0, len(docRows))
	for _, row := range docRows {
		doc, err := documentFromRow(row)
		if err != nil {
			return Project{}, nil, err
		}
		docs = append(docs, doc)
	}

	if err := tx.Commit(ctx); err != nil {
		return Project{}, nil, fmt.Errorf("commit transaction: %w", err)
	}

	return project, docs, nil
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
			created_at TIMESTAMPTZ NOT NULL
		);

		CREATE INDEX IF NOT EXISTS idx_documents_project_id_created_at
		ON documents(project_id, created_at);
	`
	if err := s.client.Exec(ctx, query); err != nil {
		return fmt.Errorf("initialize postgres schema: %w", err)
	}
	return nil
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
	if len(row) != 8 {
		return documents.Document{}, fmt.Errorf("unexpected document row column count: %d", len(row))
	}
	sizeBytes, err := strconv.ParseInt(row[4], 10, 64)
	if err != nil {
		return documents.Document{}, fmt.Errorf("parse document size_bytes: %w", err)
	}
	createdAt, err := time.Parse(time.RFC3339Nano, row[7])
	if err != nil {
		return documents.Document{}, fmt.Errorf("parse document created_at: %w", err)
	}
	return documents.Document{
		ID:         row[0],
		ProjectID:  row[1],
		Filename:   row[2],
		MimeType:   row[3],
		SizeBytes:  sizeBytes,
		StorageKey: row[5],
		Status:     documents.Status(row[6]),
		CreatedAt:  createdAt,
	}, nil
}

func sqlString(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}

func sqlTime(value time.Time) string {
	return sqlString(value.UTC().Format(time.RFC3339Nano)) + "::timestamptz"
}
