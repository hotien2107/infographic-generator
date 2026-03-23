package utils

import "testing"

func TestNewUUIDFormat(t *testing.T) {
	value := NewUUID()
	if len(value) != 36 {
		t.Fatalf("expected UUID length 36, got %d", len(value))
	}
	for _, idx := range []int{8, 13, 18, 23} {
		if value[idx] != '-' {
			t.Fatalf("expected hyphen at position %d, got %q", idx, value[idx])
		}
	}
}
