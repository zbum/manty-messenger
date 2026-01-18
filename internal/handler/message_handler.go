package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"Mmessenger/internal/middleware"
	"Mmessenger/internal/models"
	"Mmessenger/internal/service"
)

type MessageHandler struct {
	messageService *service.MessageService
}

func NewMessageHandler(messageService *service.MessageService) *MessageHandler {
	return &MessageHandler{messageService: messageService}
}

func (h *MessageHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	roomID, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid room ID")
		return
	}

	limit := 50
	offset := 0
	var afterID uint64

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	// after_id: fetch messages after this ID (for reconnection scenarios)
	if a := r.URL.Query().Get("after_id"); a != "" {
		if parsed, err := strconv.ParseUint(a, 10, 64); err == nil {
			afterID = parsed
		}
	}

	var messages []*models.MessageResponse

	if afterID > 0 {
		// Use after_id query (ignores offset)
		messages, err = h.messageService.GetByRoomIDAfter(r.Context(), roomID, claims.UserID, afterID, limit)
	} else {
		// Use standard pagination
		messages, err = h.messageService.GetByRoomID(r.Context(), roomID, claims.UserID, limit, offset)
	}

	if err != nil {
		if errors.Is(err, service.ErrNotMember) {
			respondError(w, http.StatusForbidden, "You are not a member of this room")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to get messages")
		return
	}

	respondJSON(w, http.StatusOK, messages)
}

func (h *MessageHandler) GetMessage(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	roomID, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid room ID")
		return
	}

	msgID, err := strconv.ParseUint(vars["msgId"], 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid message ID")
		return
	}

	message, err := h.messageService.GetByID(r.Context(), roomID, msgID, claims.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondError(w, http.StatusNotFound, "Message not found")
			return
		}
		if errors.Is(err, service.ErrNotMember) {
			respondError(w, http.StatusForbidden, "You are not a member of this room")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to get message")
		return
	}

	respondJSON(w, http.StatusOK, message)
}

func (h *MessageHandler) Update(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	msgID, err := strconv.ParseUint(vars["msgId"], 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid message ID")
		return
	}

	var req models.UpdateMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Content == "" {
		respondError(w, http.StatusBadRequest, "Content is required")
		return
	}

	message, err := h.messageService.Update(r.Context(), msgID, claims.UserID, &req)
	if err != nil {
		if errors.Is(err, service.ErrNotOwner) {
			respondError(w, http.StatusForbidden, "You can only edit your own messages")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to update message")
		return
	}

	respondJSON(w, http.StatusOK, message)
}

func (h *MessageHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	msgID, err := strconv.ParseUint(vars["msgId"], 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid message ID")
		return
	}

	if err := h.messageService.Delete(r.Context(), msgID, claims.UserID); err != nil {
		if errors.Is(err, service.ErrNotOwner) {
			respondError(w, http.StatusForbidden, "You can only delete your own messages")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to delete message")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
