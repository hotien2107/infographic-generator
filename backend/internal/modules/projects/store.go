package projects

import (
	"context"
	"encoding/json"
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
	ProjectStatus       Status
	ProjectStep         Step
	DocumentStatus      documents.Status
	ExtractionStartedAt *time.Time
	ExtractionEndedAt   *time.Time
	RawText             *string
	Metadata            *documents.RawContentMetadata
	ErrorMessage        *string
}

type PostgresStore struct{ client *postgres.Client }

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

func (s *PostgresStore) Close() {}

func (s *PostgresStore) GetDashboardSummary(ctx context.Context) (DashboardSummary, error) {
	projectsList, err := s.ListProjects(ctx)
	if err != nil {
		return DashboardSummary{}, err
	}
	summary := DashboardSummary{TotalProjects: len(projectsList)}
	for _, p := range projectsList {
		summary.TotalDocuments += p.DocumentCount
		switch p.Status {
		case StatusExtracting:
			summary.ProcessingProjects++
		case StatusExtracted:
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
	query := `SELECT p.id,p.title,COALESCE(p.description,''),p.input_mode,p.status,p.current_step,p.created_at,p.updated_at,COALESCE(COUNT(d.id),0)
	FROM projects p LEFT JOIN documents d ON d.project_id=p.id
	GROUP BY p.id,p.title,p.description,p.input_mode,p.status,p.current_step,p.created_at,p.updated_at
	ORDER BY p.updated_at DESC,p.created_at DESC`
	rows, err := s.client.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	items := make([]ProjectListItem, 0, len(rows))
	for _, row := range rows {
		item, err := projectListItemFromRow(row)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (s *PostgresStore) CreateProject(ctx context.Context, title, description string, inputMode InputMode) (Project, error) {
	now := time.Now().UTC()
	project := Project{ID: utils.NewUUID(), Title: title, Description: description, InputMode: inputMode, Status: StatusDraft, CurrentStep: StepWaitingUpload, CreatedAt: now, UpdatedAt: now}
	q, err := postgres.FormatQuery(`INSERT INTO projects (id,title,description,input_mode,status,current_step,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`, project.ID, project.Title, project.Description, string(project.InputMode), string(project.Status), string(project.CurrentStep), project.CreatedAt, project.UpdatedAt)
	if err != nil {
		return Project{}, err
	}
	if err := s.client.Exec(ctx, q); err != nil {
		return Project{}, err
	}
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
	q, err := postgres.FormatQuery(`UPDATE projects SET title=COALESCE($1,title),description=COALESCE($2,description),input_mode=COALESCE($3,input_mode),updated_at=$4 WHERE id=$5 RETURNING id,title,COALESCE(description,''),input_mode,status,current_step,created_at,updated_at`, nullableValue(update.Title), nullableValue(update.Description), nullableInputMode(update.InputMode), time.Now().UTC(), projectID)
	if err != nil {
		return Project{}, err
	}
	row, err := s.client.QueryRow(ctx, q)
	if err != nil {
		if errors.Is(err, postgres.ErrNoRows) {
			return Project{}, ErrProjectNotFound
		}
		return Project{}, err
	}
	project, err := projectFromRow(row)
	if err != nil {
		return Project{}, err
	}
	docs, _ := s.documentsByProject(ctx, s.client, projectID)
	project.ProcessingSummary = buildProcessingSummary(docs)
	return project, nil
}

func (s *PostgresStore) DeleteProject(ctx context.Context, projectID string) error {
	q, _ := postgres.FormatQuery(`DELETE FROM projects WHERE id=$1 RETURNING id`, projectID)
	_, err := s.client.QueryRow(ctx, q)
	if errors.Is(err, postgres.ErrNoRows) {
		return ErrProjectNotFound
	}
	return err
}

func (s *PostgresStore) AddDocument(ctx context.Context, projectID string, doc documents.Document) (Project, []documents.Document, error) {
	tx, err := s.client.Begin(ctx)
	if err != nil {
		return Project{}, nil, err
	}
	defer tx.Rollback(ctx)
	now := time.Now().UTC()
	pq, _ := postgres.FormatQuery(`UPDATE projects SET status=$1,current_step=$2,updated_at=$3 WHERE id=$4 RETURNING id,title,COALESCE(description,''),input_mode,status,current_step,created_at,updated_at`, string(StatusUploaded), string(StepUploaded), now, projectID)
	prow, err := tx.QueryRow(ctx, pq)
	if err != nil {
		if errors.Is(err, postgres.ErrNoRows) {
			return Project{}, nil, ErrProjectNotFound
		}
		return Project{}, nil, err
	}
	project, err := projectFromRow(prow)
	if err != nil {
		return Project{}, nil, err
	}
	metaJSON := marshalMetadata(doc.Metadata)
	dq, _ := postgres.FormatQuery(`INSERT INTO documents (id,project_id,filename,mime_type,size_bytes,storage_key,source_type,file_type,status,raw_text,metadata,extraction_started_at,extraction_ended_at,error_message,created_at,updated_at)
	VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`, doc.ID, doc.ProjectID, doc.Filename, doc.MimeType, doc.SizeBytes, doc.StorageKey, string(doc.SourceType), string(doc.FileType), string(doc.Status), doc.RawText, metaJSON, doc.ExtractionStartedAt, doc.ExtractionEndedAt, doc.ErrorMessage, doc.CreatedAt, doc.UpdatedAt)
	if err := tx.Exec(ctx, dq); err != nil {
		return Project{}, nil, err
	}
	docs, err := s.documentsByProject(ctx, tx, projectID)
	if err != nil {
		return Project{}, nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Project{}, nil, err
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
	q, _ := postgres.FormatQuery(`UPDATE documents SET filename=$1,updated_at=$2 WHERE id=$3 AND project_id=$4 RETURNING id,project_id,filename,mime_type,size_bytes,storage_key,source_type,file_type,status,raw_text,metadata,extraction_started_at,extraction_ended_at,error_message,created_at,updated_at`, filename, time.Now().UTC(), documentID, projectID)
	row, err := s.client.QueryRow(ctx, q)
	if err != nil {
		if errors.Is(err, postgres.ErrNoRows) {
			return documents.Document{}, ErrDocumentNotFound
		}
		return documents.Document{}, err
	}
	return documentFromRow(row)
}

func (s *PostgresStore) DeleteDocument(ctx context.Context, projectID, documentID string) error {
	tx, err := s.client.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := s.projectByIDWithExecutor(ctx, tx, projectID); err != nil {
		return err
	}
	dq, _ := postgres.FormatQuery(`DELETE FROM documents WHERE id=$1 AND project_id=$2 RETURNING id`, documentID, projectID)
	if _, err := tx.QueryRow(ctx, dq); err != nil {
		if errors.Is(err, postgres.ErrNoRows) {
			return ErrDocumentNotFound
		}
		return err
	}
	docs, err := s.documentsByProject(ctx, tx, projectID)
	if err != nil {
		return err
	}
	status, step := deriveProjectState(docs)
	pq, _ := postgres.FormatQuery(`UPDATE projects SET status=$1,current_step=$2,updated_at=$3 WHERE id=$4`, string(status), string(step), time.Now().UTC(), projectID)
	if err := tx.Exec(ctx, pq); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *PostgresStore) GetLatestDocument(ctx context.Context, projectID string) (documents.Document, error) {
	q, _ := postgres.FormatQuery(`SELECT id,project_id,filename,mime_type,size_bytes,storage_key,source_type,file_type,status,raw_text,metadata,extraction_started_at,extraction_ended_at,error_message,created_at,updated_at FROM documents WHERE project_id=$1 ORDER BY created_at DESC LIMIT 1`, projectID)
	row, err := s.client.QueryRow(ctx, q)
	if err != nil {
		if errors.Is(err, postgres.ErrNoRows) {
			return documents.Document{}, ErrNoDocumentsForProject
		}
		return documents.Document{}, err
	}
	return documentFromRow(row)
}

func (s *PostgresStore) UpdateDocumentProcessing(ctx context.Context, projectID, documentID string, p DocumentProcessingUpdate) (Project, documents.Document, error) {
	tx, err := s.client.Begin(ctx)
	if err != nil {
		return Project{}, documents.Document{}, err
	}
	defer tx.Rollback(ctx)
	now := time.Now().UTC()
	pq, _ := postgres.FormatQuery(`UPDATE projects SET status=$1,current_step=$2,updated_at=$3 WHERE id=$4 RETURNING id,title,COALESCE(description,''),input_mode,status,current_step,created_at,updated_at`, string(p.ProjectStatus), string(p.ProjectStep), now, projectID)
	prow, err := tx.QueryRow(ctx, pq)
	if err != nil {
		if errors.Is(err, postgres.ErrNoRows) {
			return Project{}, documents.Document{}, ErrProjectNotFound
		}
		return Project{}, documents.Document{}, err
	}
	project, err := projectFromRow(prow)
	if err != nil {
		return Project{}, documents.Document{}, err
	}
	metaJSON := marshalMetadata(p.Metadata)
	dq, _ := postgres.FormatQuery(`UPDATE documents SET status=$1,raw_text=COALESCE($2,raw_text),metadata=COALESCE($3,metadata),extraction_started_at=$4,extraction_ended_at=$5,error_message=$6,updated_at=$7 WHERE id=$8 AND project_id=$9 RETURNING id,project_id,filename,mime_type,size_bytes,storage_key,source_type,file_type,status,raw_text,metadata,extraction_started_at,extraction_ended_at,error_message,created_at,updated_at`, string(p.DocumentStatus), p.RawText, metaJSON, p.ExtractionStartedAt, p.ExtractionEndedAt, p.ErrorMessage, now, documentID, projectID)
	drow, err := tx.QueryRow(ctx, dq)
	if err != nil {
		if errors.Is(err, postgres.ErrNoRows) {
			return Project{}, documents.Document{}, ErrDocumentNotFound
		}
		return Project{}, documents.Document{}, err
	}
	doc, err := documentFromRow(drow)
	if err != nil {
		return Project{}, documents.Document{}, err
	}
	docs, err := s.documentsByProject(ctx, tx, projectID)
	if err != nil {
		return Project{}, documents.Document{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Project{}, documents.Document{}, err
	}
	project.ProcessingSummary = buildProcessingSummary(docs)
	return project, doc, nil
}

func (s *PostgresStore) initSchema(ctx context.Context) error {
	q := `CREATE TABLE IF NOT EXISTS projects (
	id UUID PRIMARY KEY,title VARCHAR(120) NOT NULL,description TEXT NOT NULL DEFAULT '',input_mode TEXT NOT NULL,status TEXT NOT NULL,current_step TEXT NOT NULL,created_at TIMESTAMPTZ NOT NULL,updated_at TIMESTAMPTZ NOT NULL);
	CREATE TABLE IF NOT EXISTS documents (
	id UUID PRIMARY KEY,project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,filename TEXT NOT NULL,mime_type TEXT NOT NULL,size_bytes BIGINT NOT NULL,storage_key TEXT NOT NULL DEFAULT '',source_type TEXT NOT NULL DEFAULT 'file',file_type TEXT NOT NULL,status TEXT NOT NULL,raw_text TEXT NULL,metadata JSONB NULL,extraction_started_at TIMESTAMPTZ NULL,extraction_ended_at TIMESTAMPTZ NULL,error_message TEXT NULL,created_at TIMESTAMPTZ NOT NULL,updated_at TIMESTAMPTZ NOT NULL);
	ALTER TABLE documents ADD COLUMN IF NOT EXISTS source_type TEXT NOT NULL DEFAULT 'file';
	ALTER TABLE documents ADD COLUMN IF NOT EXISTS file_type TEXT NOT NULL DEFAULT 'txt';
	ALTER TABLE documents ADD COLUMN IF NOT EXISTS raw_text TEXT NULL;
	ALTER TABLE documents ADD COLUMN IF NOT EXISTS metadata JSONB NULL;
	ALTER TABLE documents ADD COLUMN IF NOT EXISTS extraction_started_at TIMESTAMPTZ NULL;
	ALTER TABLE documents ADD COLUMN IF NOT EXISTS extraction_ended_at TIMESTAMPTZ NULL;
	CREATE INDEX IF NOT EXISTS idx_documents_project_id_created_at ON documents(project_id,created_at);`
	return s.client.Exec(ctx, q)
}

type queryExecutor interface {
	Query(ctx context.Context, query string) ([]postgres.Row, error)
}
type queryRowExecutor interface {
	QueryRow(ctx context.Context, query string) (postgres.Row, error)
}

func (s *PostgresStore) documentsByProject(ctx context.Context, ex queryExecutor, projectID string) ([]documents.Document, error) {
	q, _ := postgres.FormatQuery(`SELECT id,project_id,filename,mime_type,size_bytes,storage_key,source_type,file_type,status,raw_text,metadata,extraction_started_at,extraction_ended_at,error_message,created_at,updated_at FROM documents WHERE project_id=$1 ORDER BY created_at ASC`, projectID)
	rows, err := ex.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	res := make([]documents.Document, 0, len(rows))
	for _, r := range rows {
		d, err := documentFromRow(r)
		if err != nil {
			return nil, err
		}
		res = append(res, d)
	}
	return res, nil
}
func (s *PostgresStore) projectByID(ctx context.Context, projectID string) (Project, error) {
	return s.projectByIDWithExecutor(ctx, s.client, projectID)
}
func (s *PostgresStore) projectByIDWithExecutor(ctx context.Context, ex queryRowExecutor, projectID string) (Project, error) {
	q, _ := postgres.FormatQuery(`SELECT id,title,COALESCE(description,''),input_mode,status,current_step,created_at,updated_at FROM projects WHERE id=$1`, projectID)
	row, err := ex.QueryRow(ctx, q)
	if err != nil {
		if errors.Is(err, postgres.ErrNoRows) {
			return Project{}, ErrProjectNotFound
		}
		return Project{}, err
	}
	return projectFromRow(row)
}

func projectFromRow(row postgres.Row) (Project, error) {
	createdAt, _ := parseTimestamp(row[6])
	updatedAt, _ := parseTimestamp(row[7])
	return Project{ID: row[0], Title: row[1], Description: row[2], InputMode: InputMode(row[3]), Status: Status(row[4]), CurrentStep: Step(row[5]), CreatedAt: createdAt, UpdatedAt: updatedAt}, nil
}
func projectListItemFromRow(row postgres.Row) (ProjectListItem, error) {
	project, err := projectFromRow(row[:8])
	if err != nil {
		return ProjectListItem{}, err
	}
	count, _ := strconv.Atoi(row[8])
	return ProjectListItem{Project: project, DocumentCount: count}, nil
}

func documentFromRow(row postgres.Row) (documents.Document, error) {
	size, _ := strconv.ParseInt(row[4], 10, 64)
	rawText := nullableString(row[9])
	meta := nullableMetadata(row[10])
	started, _ := nullableTime(row[11])
	ended, _ := nullableTime(row[12])
	created, _ := parseTimestamp(row[14])
	updated, _ := parseTimestamp(row[15])
	return documents.Document{ID: row[0], ProjectID: row[1], Filename: row[2], MimeType: row[3], SizeBytes: size, StorageKey: row[5], SourceType: documents.SourceType(row[6]), FileType: documents.FileType(row[7]), Status: documents.Status(row[8]), RawText: rawText, Metadata: meta, ExtractionStartedAt: started, ExtractionEndedAt: ended, ErrorMessage: nullableString(row[13]), CreatedAt: created, UpdatedAt: updated}, nil
}

func buildProcessingSummary(docs []documents.Document) ProcessingSummary {
	s := ProcessingSummary{TotalDocuments: len(docs)}
	for _, d := range docs {
		switch d.Status {
		case documents.StatusUploaded:
			s.UploadedDocuments++
		case documents.StatusExtracting:
			s.ExtractingDocuments++
		case documents.StatusExtracted:
			s.ExtractedDocuments++
			if d.ExtractionEndedAt != nil && (s.LastExtractionAt == nil || d.ExtractionEndedAt.After(*s.LastExtractionAt)) {
				t := *d.ExtractionEndedAt
				s.LastExtractionAt = &t
			}
		case documents.StatusFailed:
			s.FailedDocuments++
			if d.ErrorMessage != nil {
				m := *d.ErrorMessage
				s.LastExtractionError = &m
			}
		}
	}
	return s
}

func deriveProjectState(docs []documents.Document) (Status, Step) {
	if len(docs) == 0 {
		return StatusDraft, StepWaitingUpload
	}
	s := buildProcessingSummary(docs)
	switch {
	case s.ExtractingDocuments > 0:
		return StatusExtracting, StepExtracting
	case s.FailedDocuments > 0:
		return StatusFailed, StepFailed
	case s.ExtractedDocuments == s.TotalDocuments:
		return StatusExtracted, StepReadyForGeneration
	default:
		return StatusUploaded, StepUploaded
	}
}

func parseTimestamp(value string) (time.Time, error) {
	layouts := []string{time.RFC3339Nano, "2006-01-02 15:04:05.999999999Z07:00", "2006-01-02 15:04:05.999999999Z07", "2006-01-02 15:04:05Z07:00", "2006-01-02 15:04:05Z07"}
	for _, l := range layouts {
		if t, err := time.Parse(l, value); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported timestamp format")
}
func nullableTime(value string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}
	t, err := parseTimestamp(value)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
func nullableString(value string) *string {
	if value == "" {
		return nil
	}
	v := value
	return &v
}
func nullableValue(v *string) *string { return v }
func nullableInputMode(v *InputMode) *string {
	if v == nil {
		return nil
	}
	r := string(*v)
	return &r
}
func marshalMetadata(meta *documents.RawContentMetadata) *string {
	if meta == nil {
		return nil
	}
	b, err := json.Marshal(meta)
	if err != nil {
		return nil
	}
	s := string(b)
	return &s
}
func nullableMetadata(value string) *documents.RawContentMetadata {
	if value == "" {
		return nil
	}
	var meta documents.RawContentMetadata
	if err := json.Unmarshal([]byte(value), &meta); err != nil {
		return nil
	}
	return &meta
}
