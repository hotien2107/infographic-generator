package documents

import "time"

type Status string

const StatusUploaded Status = "uploaded"

type Document struct {
	ID         string    `json:"id"`
	ProjectID  string    `json:"project_id"`
	Filename   string    `json:"filename"`
	MimeType   string    `json:"mime_type"`
	SizeBytes  int64     `json:"size_bytes"`
	StorageKey string    `json:"storage_key"`
	Status     Status    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}
