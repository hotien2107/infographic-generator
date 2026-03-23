package config

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Config struct {
	AppEnv           string
	Port             string
	StorageDir       string
	MaxUploadSizeMB  int
	AllowedFileTypes []string
	FrontendOrigin   string
}

func Load() Config {
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}

	storageDir := getEnv("OBJECT_STORAGE_DIR", filepath.Join(cwd, "tmp", "uploads"))

	return Config{
		AppEnv:           getEnv("APP_ENV", "development"),
		Port:             getEnv("API_PORT", "8080"),
		StorageDir:       storageDir,
		MaxUploadSizeMB:  getEnvAsInt("MAX_UPLOAD_SIZE_MB", 10),
		AllowedFileTypes: getEnvAsSlice("ALLOWED_FILE_TYPES", []string{"pdf", "docx", "txt"}),
		FrontendOrigin:   getEnv("FRONTEND_ORIGIN", "http://localhost:5173"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && strings.TrimSpace(value) != "" {
		return value
	}

	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	value, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(value) == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func getEnvAsSlice(key string, fallback []string) []string {
	value, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(value) == "" {
		return fallback
	}

	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.ToLower(strings.TrimSpace(part))
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	if len(result) == 0 {
		return fallback
	}

	return result
}
