package storage

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"mime/multipart"
	"path"
	"path/filepath"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"infographic-generator/backend/internal/config"
	"infographic-generator/backend/internal/utils"
)

type MinIOStorage struct {
	client              *minio.Client
	bucket              string
	multipartPartSizeMB uint64
}

func NewMinIOStorage(ctx context.Context, cfg config.Config) (*MinIOStorage, error) {
	client, err := minio.New(cfg.MinIOEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinIOAccessKey, cfg.MinIOSecretKey, ""),
		Secure: cfg.MinIOUseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("create minio client: %w", err)
	}

	storage := &MinIOStorage{
		client:              client,
		bucket:              cfg.MinIOBucket,
		multipartPartSizeMB: cfg.MultipartPartSizeMB,
	}

	if cfg.MinIOAutoCreateBucket {
		if err := storage.ensureBucket(ctx); err != nil {
			return nil, err
		}
	}

	return storage, nil
}

func (s *MinIOStorage) Save(ctx context.Context, fileHeader *multipart.FileHeader) (string, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("open upload stream: %w", err)
	}
	defer file.Close()

	objectKey := path.Join(utils.NewUUID(), sanitizeFilename(fileHeader.Filename))
	contentType := fileHeader.Header.Get("Content-Type")
	if strings.TrimSpace(contentType) == "" {
		contentType = "application/octet-stream"
	}

	putOptions := minio.PutObjectOptions{ContentType: contentType}
	if s.multipartPartSizeMB > 0 {
		putOptions.PartSize = s.multipartPartSizeMB * 1024 * 1024
	}

	_, err = s.client.PutObject(ctx, s.bucket, objectKey, file, fileHeader.Size, putOptions)
	if err != nil {
		return "", fmt.Errorf("upload object to minio: %w", err)
	}

	return objectKey, nil
}

func (s *MinIOStorage) Close() error { return nil }

func (s *MinIOStorage) ensureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("check minio bucket: %w", err)
	}
	if exists {
		return nil
	}

	if err := s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{}); err != nil {
		existsAfterCreate, existsErr := s.client.BucketExists(ctx, s.bucket)
		if existsErr != nil {
			return fmt.Errorf("create minio bucket: %w", err)
		}
		if !existsAfterCreate {
			return fmt.Errorf("create minio bucket: %w", err)
		}
	}

	return nil
}

func sanitizeFilename(name string) string {
	base := filepath.Base(strings.TrimSpace(name))
	if base == "." || base == string(filepath.Separator) || base == "" {
		buf := make([]byte, 8)
		if _, err := rand.Read(buf); err == nil {
			return hex.EncodeToString(buf) + ".bin"
		}
		return "upload.bin"
	}
	return base
}
