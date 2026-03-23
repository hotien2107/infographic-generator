package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	AppEnv                string
	Port                  string
	MaxUploadSizeMB       int
	AllowedFileTypes      []string
	PostgresURL           string
	MinIOEndpoint         string
	MinIOAccessKey        string
	MinIOSecretKey        string
	MinIOBucket           string
	MinIOUseSSL           bool
	MultipartThresholdMB  int64
	MultipartPartSizeMB   uint64
	MinIOAutoCreateBucket bool
}

func Load() Config {
	return Config{
		AppEnv:                getEnv("APP_ENV", "development"),
		Port:                  getEnv("API_PORT", "8080"),
		MaxUploadSizeMB:       getEnvAsInt("MAX_UPLOAD_SIZE_MB", 10),
		AllowedFileTypes:      getEnvAsSlice("ALLOWED_FILE_TYPES", []string{"pdf", "docx", "txt"}),
		PostgresURL:           getEnv("POSTGRES_URL", "postgres://postgres:postgres@localhost:5432/infographic_generator?sslmode=disable"),
		MinIOEndpoint:         getEnv("MINIO_ENDPOINT", "localhost:9000"),
		MinIOAccessKey:        getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		MinIOSecretKey:        getEnv("MINIO_SECRET_KEY", "minioadmin"),
		MinIOBucket:           getEnv("MINIO_BUCKET", "documents"),
		MinIOUseSSL:           getEnvAsBool("MINIO_USE_SSL", false),
		MultipartThresholdMB:  getEnvAsInt64("MINIO_MULTIPART_THRESHOLD_MB", 16),
		MultipartPartSizeMB:   getEnvAsUint64("MINIO_MULTIPART_PART_SIZE_MB", 8),
		MinIOAutoCreateBucket: getEnvAsBool("MINIO_AUTO_CREATE_BUCKET", true),
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

func getEnvAsInt64(key string, fallback int64) int64 {
	value, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(value) == "" {
		return fallback
	}

	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fallback
	}

	return parsed
}

func getEnvAsUint64(key string, fallback uint64) uint64 {
	value, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(value) == "" {
		return fallback
	}

	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return fallback
	}

	return parsed
}

func getEnvAsBool(key string, fallback bool) bool {
	value, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(value) == "" {
		return fallback
	}

	parsed, err := strconv.ParseBool(value)
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
