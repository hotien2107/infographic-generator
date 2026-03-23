package storage

import (
	"context"
	"mime/multipart"
)

type BlobStorage interface {
	Save(ctx context.Context, fileHeader *multipart.FileHeader) (string, error)
	Close() error
}
