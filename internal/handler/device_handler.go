package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"Mmessenger/internal/middleware"
	"Mmessenger/internal/models"
	"Mmessenger/internal/service"
)

type DeviceHandler struct {
	fcmService *service.FCMService
}

func NewDeviceHandler(fcmService *service.FCMService) *DeviceHandler {
	return &DeviceHandler{fcmService: fcmService}
}

// Register handles device token registration
func (h *DeviceHandler) Register(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req models.RegisterDeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.DeviceID == "" || req.Token == "" {
		respondError(w, http.StatusBadRequest, "device_id and token are required")
		return
	}

	if req.Platform == "" {
		req.Platform = "fcm"
	}

	if err := h.fcmService.RegisterDevice(r.Context(), claims.UserID, &req); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to register device")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Device registered successfully",
	})
}

// Unregister handles device token removal
func (h *DeviceHandler) Unregister(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	deviceID := vars["deviceId"]
	if deviceID == "" {
		respondError(w, http.StatusBadRequest, "Device ID is required")
		return
	}

	if err := h.fcmService.UnregisterDevice(r.Context(), claims.UserID, deviceID); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to unregister device")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Device unregistered successfully",
	})
}
