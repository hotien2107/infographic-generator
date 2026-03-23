package postgres

import (
	"strings"
	"testing"
)

func TestFormatQueryReplacesDoubleDigitPlaceholdersWithoutCorruption(t *testing.T) {
	query := "INSERT INTO documents VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)"

	formatted, err := FormatQuery(query,
		"one",
		"two",
		"three",
		"four",
		"five",
		"six",
		"seven",
		"eight",
		"nine",
		"ten",
		"eleven",
		"twelve",
		"thirteen",
	)
	if err != nil {
		t.Fatalf("FormatQuery() error = %v", err)
	}

	for _, want := range []string{"'one'", "'ten'", "'eleven'", "'thirteen'"} {
		if !strings.Contains(formatted, want) {
			t.Fatalf("formatted query %q missing %s", formatted, want)
		}
	}
	for _, unexpected := range []string{"'one'0", "'one'1", "'one'2", "'one'3"} {
		if strings.Contains(formatted, unexpected) {
			t.Fatalf("formatted query %q unexpectedly contains %s", formatted, unexpected)
		}
	}
}

func TestFormatQueryLeavesUnknownPlaceholdersUntouched(t *testing.T) {
	formatted, err := FormatQuery("SELECT $1, $2", "one")
	if err != nil {
		t.Fatalf("FormatQuery() error = %v", err)
	}
	if formatted != "SELECT 'one', $2" {
		t.Fatalf("FormatQuery() = %q, want %q", formatted, "SELECT 'one', $2")
	}
}
