package projects

import "time"

type InputMode string

type Status string

type Step string

const (
	InputModeFile InputMode = "file"
	InputModeText InputMode = "text"

	StatusDraft      Status = "draft"
	StatusUploaded   Status = "uploaded"
	StatusExtracting Status = "extracting"
	StatusExtracted  Status = "extracted"
	StatusFailed     Status = "failed"

	StepWaitingUpload      Step = "waiting_for_upload"
	StepUploaded           Step = "uploaded"
	StepQueuedForExtract   Step = "queued_for_extraction"
	StepExtracting         Step = "extracting"
	StepReadyForGeneration Step = "ready_for_generation"
	StepFailed             Step = "failed"
)

type Project struct {
	ID                string            `json:"id"`
	Title             string            `json:"title"`
	Description       string            `json:"description"`
	InputMode         InputMode         `json:"input_mode"`
	Status            Status            `json:"status"`
	CurrentStep       Step              `json:"current_step"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
	ProcessingSummary ProcessingSummary `json:"processing_summary"`
}

type ProjectListItem struct {
	Project
	DocumentCount int `json:"document_count"`
}

type DashboardSummary struct {
	TotalProjects      int `json:"total_projects"`
	TotalDocuments     int `json:"total_documents"`
	ProcessingProjects int `json:"processing_projects"`
	CompletedProjects  int `json:"completed_projects"`
	AttentionProjects  int `json:"attention_projects"`
	DraftProjects      int `json:"draft_projects"`
}

type ProjectUpdate struct {
	Title       *string
	Description *string
	InputMode   *InputMode
}

type ProcessingSummary struct {
	TotalDocuments      int        `json:"total_documents"`
	UploadedDocuments   int        `json:"uploaded_documents"`
	ExtractingDocuments int        `json:"extracting_documents"`
	ExtractedDocuments  int        `json:"extracted_documents"`
	FailedDocuments     int        `json:"failed_documents"`
	LastExtractionAt    *time.Time `json:"last_extraction_at"`
	LastExtractionError *string    `json:"last_extraction_error"`
}

type Detail struct {
	Project
	Documents any `json:"documents"`
}
