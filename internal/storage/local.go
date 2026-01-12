package storage

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

type LocalStorage struct {
	basePath    string
	baseURL     string
	maxFileSize int64
}

func NewLocalStorage(basePath, baseURL string, maxFileSize int64) (*LocalStorage, error) {
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}
	return &LocalStorage{
		basePath:    basePath,
		baseURL:     baseURL,
		maxFileSize: maxFileSize,
	}, nil
}

func (s *LocalStorage) Save(ctx context.Context, file multipart.File, header *multipart.FileHeader) (*FileInfo, error) {
	if header.Size > s.maxFileSize {
		return nil, ErrFileTooLarge
	}

	ext := filepath.Ext(header.Filename)
	fileID := uuid.New().String()
	dateDir := time.Now().Format("2006/01/02")
	storedName := fileID + ext

	fullDir := filepath.Join(s.basePath, dateDir)
	if err := os.MkdirAll(fullDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	destPath := filepath.Join(fullDir, storedName)
	dst, err := os.Create(destPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	written, err := io.Copy(dst, file)
	if err != nil {
		os.Remove(destPath)
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	mimeType := header.Header.Get("Content-Type")
	fileURL := fmt.Sprintf("%s/%s/%s", s.baseURL, dateDir, storedName)

	fileInfo := &FileInfo{
		ID:           fileID,
		OriginalName: header.Filename,
		StoredName:   storedName,
		Size:         written,
		MimeType:     mimeType,
		URL:          fileURL,
	}

	// Generate thumbnail for images
	if IsImageFile(mimeType) {
		thumbPath := GetThumbnailPath(destPath)
		if err := GenerateThumbnail(destPath, thumbPath); err == nil {
			thumbURL := GetThumbnailURL(fileURL)
			fileInfo.ThumbnailURL = &thumbURL
		}
	}

	return fileInfo, nil
}

func (s *LocalStorage) Delete(ctx context.Context, fileID string) error {
	pattern := filepath.Join(s.basePath, "*", "*", "*", fileID+".*")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}

	if len(matches) == 0 {
		return ErrFileNotFound
	}

	for _, match := range matches {
		if err := os.Remove(match); err != nil {
			return err
		}
	}

	return nil
}

func (s *LocalStorage) Get(ctx context.Context, fileID string) (io.ReadCloser, *FileInfo, error) {
	pattern := filepath.Join(s.basePath, "*", "*", "*", fileID+".*")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, nil, err
	}

	if len(matches) == 0 {
		return nil, nil, ErrFileNotFound
	}

	filePath := matches[0]
	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, err
	}

	stat, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, nil, err
	}

	relPath, _ := filepath.Rel(s.basePath, filePath)
	storedName := filepath.Base(filePath)

	return file, &FileInfo{
		ID:         fileID,
		StoredName: storedName,
		Size:       stat.Size(),
		URL:        fmt.Sprintf("%s/%s", s.baseURL, relPath),
	}, nil
}
