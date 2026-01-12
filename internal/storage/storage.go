package storage

import (
	"context"
	"io"
	"mime/multipart"
)

type FileInfo struct {
	ID           string `json:"id"`
	OriginalName string `json:"original_name"`
	StoredName   string `json:"stored_name"`
	Size         int64  `json:"size"`
	MimeType     string `json:"mime_type"`
	URL          string `json:"url"`
}

type Storage interface {
	Save(ctx context.Context, file multipart.File, header *multipart.FileHeader) (*FileInfo, error)
	Delete(ctx context.Context, fileID string) error
	Get(ctx context.Context, fileID string) (io.ReadCloser, *FileInfo, error)
}
