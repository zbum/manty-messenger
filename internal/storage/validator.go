package storage

import (
	"log"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
)

var allowedMimeTypes = map[string]bool{
	// Images
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,

	// Documents
	"application/pdf":    true,
	"application/msword": true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
	"application/vnd.ms-excel": true,
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":       true,
	"application/vnd.ms-powerpoint":                                           true,
	"application/vnd.openxmlformats-officedocument.presentationml.presentation": true,

	// Archives (xlsx, docx, pptx are actually zip files)
	"application/zip":              true,
	"application/x-zip-compressed": true,
	"application/x-rar-compressed": true,
	"application/octet-stream":     true, // Generic binary - rely on extension check

	// Text
	"text/plain": true,
}

var allowedExtensions = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
	".pdf": true, ".doc": true, ".docx": true, ".xls": true, ".xlsx": true,
	".ppt": true, ".pptx": true,
	".zip": true, ".rar": true, ".txt": true,
}

func ValidateFile(file multipart.File, header *multipart.FileHeader) error {
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !allowedExtensions[ext] {
		return ErrInvalidFileType
	}

	buffer := make([]byte, 512)
	_, err := file.Read(buffer)
	if err != nil {
		return err
	}

	if _, err := file.Seek(0, 0); err != nil {
		return err
	}

	mimeType := http.DetectContentType(buffer)
	log.Printf("[ValidateFile] filename=%s, ext=%s, detected_mime=%s", header.Filename, ext, mimeType)
	if !allowedMimeTypes[mimeType] {
		log.Printf("[ValidateFile] REJECTED: mime type %s not in allowed list", mimeType)
		return ErrInvalidFileType
	}

	return nil
}

func IsImageType(mimeType string) bool {
	return strings.HasPrefix(mimeType, "image/")
}
