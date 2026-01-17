package handler

import (
	"log"
	"net/http"

	"Mmessenger/internal/middleware"
	"Mmessenger/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		log.Printf("[AuthHandler.Me] No claims in context")
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	log.Printf("[AuthHandler.Me] UserID: %d, KeycloakID: %s", claims.UserID, claims.KeycloakID)

	user, err := h.authService.GetUserByID(r.Context(), claims.UserID)
	if err != nil {
		log.Printf("[AuthHandler.Me] Error getting user: %v, UserID: %d", err, claims.UserID)
		respondError(w, http.StatusInternalServerError, "Failed to get user")
		return
	}

	respondJSON(w, http.StatusOK, user.ToResponse())
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if err := h.authService.Logout(r.Context(), claims.UserID); err != nil {
		log.Printf("[AuthHandler.Logout] Error: %v, UserID: %d", err, claims.UserID)
		respondError(w, http.StatusInternalServerError, "Failed to logout")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Logged out successfully"})
}
