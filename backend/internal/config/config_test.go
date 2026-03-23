package config_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"infographic-generator/backend/internal/config"
)

func TestLoadFromPathsPrefersDotEnvValues(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("API_PORT", "9999")
	t.Setenv("MAX_UPLOAD_SIZE_MB", "99")

	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	content := []byte("APP_ENV=staging\nAPI_PORT=8090\nMAX_UPLOAD_SIZE_MB=25\nALLOWED_FILE_TYPES=pdf,txt\nPOSTGRES_URL=postgres://demo:demo@localhost:5432/demo?sslmode=disable\nPROCESSING_STEP_DELAY_MS=123\n")
	if err := os.WriteFile(envPath, content, 0o600); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	cfg := config.LoadFromPaths(envPath)
	if cfg.AppEnv != "staging" {
		t.Fatalf("expected APP_ENV from .env, got %q", cfg.AppEnv)
	}
	if cfg.Port != "8090" {
		t.Fatalf("expected API_PORT from .env, got %q", cfg.Port)
	}
	if cfg.MaxUploadSizeMB != 25 {
		t.Fatalf("expected MAX_UPLOAD_SIZE_MB from .env, got %d", cfg.MaxUploadSizeMB)
	}
	if len(cfg.AllowedFileTypes) != 2 || cfg.AllowedFileTypes[0] != "pdf" || cfg.AllowedFileTypes[1] != "txt" {
		t.Fatalf("unexpected ALLOWED_FILE_TYPES: %+v", cfg.AllowedFileTypes)
	}
	if cfg.PostgresURL != "postgres://demo:demo@localhost:5432/demo?sslmode=disable" {
		t.Fatalf("unexpected POSTGRES_URL: %q", cfg.PostgresURL)
	}
	if cfg.ProcessingStepDelay != 123*time.Millisecond {
		t.Fatalf("unexpected PROCESSING_STEP_DELAY_MS: %s", cfg.ProcessingStepDelay)
	}
}

func TestLoadFromPathsFallsBackWhenDotEnvMissing(t *testing.T) {
	cfg := config.LoadFromPaths(filepath.Join(t.TempDir(), ".env"))
	if cfg.AppEnv != "development" {
		t.Fatalf("expected default APP_ENV, got %q", cfg.AppEnv)
	}
	if cfg.Port != "8080" {
		t.Fatalf("expected default API_PORT, got %q", cfg.Port)
	}
}
