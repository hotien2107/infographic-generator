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
	StatusProcessing Status = "processing"
	StatusProcessed  Status = "processed"
	StatusFailed     Status = "failed"

	StepWaitingUpload      Step = "waiting_for_upload"
	StepUploaded           Step = "uploaded"
	StepQueuedProcessing   Step = "queued_for_processing"
	StepExtracting         Step = "extracting"
	StepReadyForGeneration Step = "ready_for_generation"
	StepFailed             Step = "failed"
)

type Project struct {
	ID                string            `json:"id"`
	Title             string            `json:"title"`
	InputMode         InputMode         `json:"input_mode"`
	Status            Status            `json:"status"`
	CurrentStep       Step              `json:"current_step"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
	ProcessingSummary ProcessingSummary `json:"processing_summary"`
}

type ProcessingSummary struct {
	TotalDocuments      int        `json:"total_documents"`
	UploadedDocuments   int        `json:"uploaded_documents"`
	QueuedDocuments     int        `json:"queued_documents"`
	ProcessingDocuments int        `json:"processing_documents"`
	ProcessedDocuments  int        `json:"processed_documents"`
	FailedDocuments     int        `json:"failed_documents"`
	LastProcessedAt     *time.Time `json:"last_processed_at"`
	LastError           *string    `json:"last_error"`
}

type Detail struct {
	Project
	Documents any `json:"documents"`
}
