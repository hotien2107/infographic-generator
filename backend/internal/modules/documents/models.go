package documents

import "time"

type Status string

const (
	StatusUploaded   Status = "uploaded"
	StatusQueued     Status = "queued"
	StatusProcessing Status = "processing"
	StatusProcessed  Status = "processed"
	StatusFailed     Status = "failed"
)

type Document struct {
	ID                   string     `json:"id"`
	ProjectID            string     `json:"project_id"`
	Filename             string     `json:"filename"`
	MimeType             string     `json:"mime_type"`
	SizeBytes            int64      `json:"size_bytes"`
	StorageKey           string     `json:"storage_key"`
	Status               Status     `json:"status"`
	ProcessingStartedAt  *time.Time `json:"processing_started_at"`
	ProcessingFinishedAt *time.Time `json:"processing_finished_at"`
	ErrorMessage         *string    `json:"error_message"`
	ExtractedTextPreview *string    `json:"extracted_text_preview"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}
