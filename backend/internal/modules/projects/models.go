package projects

import "time"

type InputMode string

type Status string

type Step string

const (
	InputModeFile InputMode = "file"
	InputModeText InputMode = "text"

	StatusDraft    Status = "draft"
	StatusUploaded Status = "uploaded"

	StepProjectCreated Step = "project_created"
	StepWaitingUpload  Step = "waiting_for_upload"
	StepUploaded       Step = "uploaded"
)

type Project struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	InputMode   InputMode `json:"input_mode"`
	Status      Status    `json:"status"`
	CurrentStep Step      `json:"current_step"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Detail struct {
	Project
	Documents any `json:"documents"`
}
