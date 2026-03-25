package storage

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"infographic-generator/backend/internal/config"
	"infographic-generator/backend/internal/utils"
)

// MinIOStorage giữ nguyên tên để tương thích code cũ, nhưng dùng local filesystem cho môi trường dev/test.
type MinIOStorage struct {
	basePath string
}

func NewMinIOStorage(_ context.Context, cfg config.Config) (*MinIOStorage, error) {
	basePath := filepath.Join(os.TempDir(), "infographic-generator", cfg.MinIOBucket)
	if err := os.MkdirAll(basePath, 0o755); err != nil {
		return nil, fmt.Errorf("create storage directory: %w", err)
	}
	return &MinIOStorage{basePath: basePath}, nil
}

func (s *MinIOStorage) Save(_ context.Context, fileHeader *multipart.FileHeader) (string, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("open upload stream: %w", err)
	}
	defer file.Close()
	key := filepath.Join(utils.NewUUID(), sanitizeFilename(fileHeader.Filename))
	path := filepath.Join(s.basePath, key)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	out, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer out.Close()
	if _, err := io.Copy(out, file); err != nil {
		return "", err
	}
	return key, nil
}

func (s *MinIOStorage) Read(_ context.Context, storageKey string) ([]byte, error) {
	return os.ReadFile(filepath.Join(s.basePath, storageKey))
}

func (s *MinIOStorage) Close() error { return nil }

func sanitizeFilename(name string) string {
	base := filepath.Base(strings.TrimSpace(name))
	if base == "." || base == string(filepath.Separator) || base == "" {
		return "upload.bin"
	}
	return base
}
