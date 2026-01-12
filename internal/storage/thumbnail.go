package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/davidbyttow/govips/v2/vips"
)

const (
	ThumbnailWidth  = 300
	ThumbnailHeight = 300
	ThumbnailSuffix = "_thumb"
)

var thumbnailInitialized = false

// InitThumbnail initializes the vips library
func InitThumbnail() {
	if !thumbnailInitialized {
		vips.Startup(nil)
		thumbnailInitialized = true
	}
}

// ShutdownThumbnail shuts down the vips library
func ShutdownThumbnail() {
	if thumbnailInitialized {
		vips.Shutdown()
		thumbnailInitialized = false
	}
}

// IsImageFile checks if the mime type is an image
func IsImageFile(mimeType string) bool {
	return strings.HasPrefix(mimeType, "image/")
}

// GenerateThumbnail creates a thumbnail for an image file
func GenerateThumbnail(srcPath, dstPath string) error {
	// Load image
	image, err := vips.NewImageFromFile(srcPath)
	if err != nil {
		return fmt.Errorf("failed to load image: %w", err)
	}
	defer image.Close()

	// Calculate thumbnail dimensions maintaining aspect ratio
	width := image.Width()
	height := image.Height()

	var scale float64
	if width > height {
		scale = float64(ThumbnailWidth) / float64(width)
	} else {
		scale = float64(ThumbnailHeight) / float64(height)
	}

	// Only resize if image is larger than thumbnail size
	if scale < 1 {
		err = image.Resize(scale, vips.KernelLanczos3)
		if err != nil {
			return fmt.Errorf("failed to resize image: %w", err)
		}
	}

	// Export based on original format or default to JPEG
	ext := strings.ToLower(filepath.Ext(srcPath))
	var imageBytes []byte

	switch ext {
	case ".png":
		ep := vips.NewPngExportParams()
		ep.Compression = 6
		imageBytes, _, err = image.ExportPng(ep)
	case ".gif":
		ep := vips.NewGifExportParams()
		imageBytes, _, err = image.ExportGIF(ep)
	case ".webp":
		ep := vips.NewWebpExportParams()
		ep.Quality = 80
		imageBytes, _, err = image.ExportWebp(ep)
	default:
		ep := vips.NewJpegExportParams()
		ep.Quality = 80
		ep.StripMetadata = true
		imageBytes, _, err = image.ExportJpeg(ep)
	}

	if err != nil {
		return fmt.Errorf("failed to export thumbnail: %w", err)
	}

	// Write thumbnail file
	err = os.WriteFile(dstPath, imageBytes, 0644)
	if err != nil {
		return fmt.Errorf("failed to write thumbnail: %w", err)
	}

	return nil
}

// GetThumbnailPath returns the thumbnail path for a given file path
func GetThumbnailPath(filePath string) string {
	ext := filepath.Ext(filePath)
	base := strings.TrimSuffix(filePath, ext)
	return base + ThumbnailSuffix + ext
}

// GetThumbnailURL returns the thumbnail URL for a given file URL
func GetThumbnailURL(fileURL string) string {
	ext := filepath.Ext(fileURL)
	base := strings.TrimSuffix(fileURL, ext)
	return base + ThumbnailSuffix + ext
}
