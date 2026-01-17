package handler

import (
	"encoding/json"
	"net/http"

	"Mmessenger/internal/middleware"
	"Mmessenger/internal/models"
	"Mmessenger/internal/service"
)

type PushHandler struct {
	pushService *service.PushService
}

func NewPushHandler(pushService *service.PushService) *PushHandler {
	return &PushHandler{pushService: pushService}
}

// GetVAPIDPublicKey returns the VAPID public key for the client
func (h *PushHandler) GetVAPIDPublicKey(w http.ResponseWriter, r *http.Request) {
	if !h.pushService.IsConfigured() {
		respondError(w, http.StatusServiceUnavailable, "Web Push is not configured")
		return
	}

	publicKey := h.pushService.GetVAPIDPublicKey()
	respondJSON(w, http.StatusOK, map[string]string{
		"public_key": publicKey,
	})
}

// Subscribe handles push subscription registration
func (h *PushHandler) Subscribe(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req models.SubscribePushRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Endpoint == "" || req.Keys.P256dh == "" || req.Keys.Auth == "" {
		respondError(w, http.StatusBadRequest, "Missing required fields")
		return
	}

	if err := h.pushService.Subscribe(r.Context(), claims.UserID, &req); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to subscribe")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Subscribed successfully",
	})
}

// Unsubscribe handles push subscription removal
func (h *PushHandler) Unsubscribe(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if err := h.pushService.Unsubscribe(r.Context(), claims.UserID); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to unsubscribe")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Unsubscribed successfully",
	})
}
