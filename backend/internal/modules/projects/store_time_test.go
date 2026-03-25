package projects

import (
	"testing"
	"time"
)

func TestParseTimestamp(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{name: "rfc3339", input: "2026-03-25T13:38:47.177049+07:00"},
		{name: "postgres timezone short", input: "2026-03-25 13:38:47.177049+07"},
		{name: "postgres timezone full", input: "2026-03-25 13:38:47.177049+07:00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := parseTimestamp(tt.input)
			if err != nil {
				t.Fatalf("parseTimestamp() error = %v", err)
			}
			if parsed.IsZero() {
				t.Fatal("parseTimestamp() returned zero time")
			}
		})
	}
}

func TestParseTimestamp_Invalid(t *testing.T) {
	_, err := parseTimestamp("2026/03/25 13:38:47")
	if err == nil {
		t.Fatal("expected error for invalid timestamp format")
	}
}

func TestProjectFromRow_ParsesPostgresTimestampFormat(t *testing.T) {
	row := []string{
		"project-1",
		"title",
		"description",
		string(InputModeUpload),
		string(StatusDraft),
		string(StepWaitingUpload),
		"2026-03-25 13:38:47.177049+07",
		"2026-03-25 13:38:47.177049+07",
	}

	project, err := projectFromRow(row)
	if err != nil {
		t.Fatalf("projectFromRow() error = %v", err)
	}

	if project.CreatedAt.IsZero() || project.UpdatedAt.IsZero() {
		t.Fatal("expected parsed created_at and updated_at")
	}

	if project.CreatedAt.Location() == time.UTC {
		t.Fatal("expected timezone information to be preserved")
	}
}
