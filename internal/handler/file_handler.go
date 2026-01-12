package handler

import (
	"net/http"

	"Mmessenger/internal/middleware"
	"Mmessenger/internal/storage"
)

type FileHandler struct {
	storage storage.Storage
	maxSize int64
}

func NewFileHandler(s storage.Storage, maxSize int64) *FileHandler {
	return &FileHandler{
		storage: s,
		maxSize: maxSize,
	}
}

func (h *FileHandler) Upload(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, h.maxSize)
	if err := r.ParseMultipartForm(h.maxSize); err != nil {
		respondError(w, http.StatusBadRequest, "File too large or invalid form data")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Failed to read file")
		return
	}
	defer file.Close()

	if err := storage.ValidateFile(file, header); err != nil {
		if err == storage.ErrInvalidFileType {
			respondError(w, http.StatusBadRequest, "File type not allowed")
			return
		}
		respondError(w, http.StatusBadRequest, "Invalid file")
		return
	}

	fileInfo, err := h.storage.Save(r.Context(), file, header)
	if err != nil {
		if err == storage.ErrFileTooLarge {
			respondError(w, http.StatusBadRequest, "File too large")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to save file")
		return
	}

	respondJSON(w, http.StatusOK, fileInfo)
}
