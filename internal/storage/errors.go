package storage

import "errors"

var (
	ErrFileTooLarge    = errors.New("file size exceeds maximum allowed")
	ErrInvalidFileType = errors.New("file type not allowed")
	ErrFileNotFound    = errors.New("file not found")
)
