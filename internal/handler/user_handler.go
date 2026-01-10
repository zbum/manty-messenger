package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"Mmessenger/internal/middleware"
	"Mmessenger/internal/repository"
)

type UserHandler struct {
	userRepo *repository.UserRepository
}

func NewUserHandler(userRepo *repository.UserRepository) *UserHandler {
	return &UserHandler{userRepo: userRepo}
}

func (h *UserHandler) Search(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	keyword := r.URL.Query().Get("q")
	if keyword == "" {
		respondError(w, http.StatusBadRequest, "Search query is required")
		return
	}

	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 50 {
			limit = parsed
		}
	}

	users, err := h.userRepo.Search(r.Context(), keyword, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to search users")
		return
	}

	var responses []interface{}
	for _, user := range users {
		responses = append(responses, user.ToResponse())
	}

	respondJSON(w, http.StatusOK, responses)
}

func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	userID, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	user, err := h.userRepo.GetByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondError(w, http.StatusNotFound, "User not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to get user")
		return
	}

	respondJSON(w, http.StatusOK, user.ToResponse())
}
