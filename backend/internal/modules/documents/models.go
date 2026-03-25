package documents

import "time"

type SourceType string

type FileType string

type Status string

const (
	SourceTypeFile SourceType = "file"
	SourceTypeText SourceType = "text"

	FileTypePDF  FileType = "pdf"
	FileTypeTXT  FileType = "txt"
	FileTypeText FileType = "text"

	StatusUploaded   Status = "uploaded"
	StatusExtracting Status = "extracting"
	StatusExtracted  Status = "extracted"
	StatusFailed     Status = "failed"
)

type RawContentMetadata struct {
	FileType        FileType   `json:"file_type"`
	SourceType      SourceType `json:"source_type"`
	PageCount       int        `json:"page_count"`
	SectionHeadings []string   `json:"section_headings"`
	ExtractedAt     time.Time  `json:"extracted_at"`
	CharacterCount  int        `json:"character_count"`
}

type Document struct {
	ID                  string              `json:"id"`
	ProjectID           string              `json:"project_id"`
	Filename            string              `json:"filename"`
	MimeType            string              `json:"mime_type"`
	SizeBytes           int64               `json:"size_bytes"`
	StorageKey          string              `json:"storage_key"`
	SourceType          SourceType          `json:"source_type"`
	FileType            FileType            `json:"file_type"`
	Status              Status              `json:"status"`
	RawText             *string             `json:"raw_text"`
	Metadata            *RawContentMetadata `json:"metadata"`
	ExtractionStartedAt *time.Time          `json:"extraction_started_at"`
	ExtractionEndedAt   *time.Time          `json:"extraction_ended_at"`
	ErrorMessage        *string             `json:"error_message"`
	CreatedAt           time.Time           `json:"created_at"`
	UpdatedAt           time.Time           `json:"updated_at"`
}
