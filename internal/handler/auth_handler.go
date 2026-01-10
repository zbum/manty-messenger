package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"Mmessenger/internal/middleware"
	"Mmessenger/internal/models"
	"Mmessenger/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Email == "" || req.Username == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "Email, username, and password are required")
		return
	}

	resp, err := h.authService.Register(r.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrUserAlreadyExists) {
			respondError(w, http.StatusConflict, "User already exists")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to register user")
		return
	}

	respondJSON(w, http.StatusCreated, resp)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	resp, err := h.authService.Login(r.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			respondError(w, http.StatusUnauthorized, "Invalid email or password")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to login")
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.authService.Logout(r.Context(), claims.UserID, req.RefreshToken); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to logout")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Logged out successfully"})
}

func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.RefreshToken == "" {
		respondError(w, http.StatusBadRequest, "Refresh token is required")
		return
	}

	resp, err := h.authService.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "Invalid or expired refresh token")
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	user, err := h.authService.GetUserByID(r.Context(), claims.UserID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get user")
		return
	}

	respondJSON(w, http.StatusOK, user.ToResponse())
}
