package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"Mmessenger/internal/middleware"
	"Mmessenger/internal/models"
	"Mmessenger/internal/service"
	"Mmessenger/internal/websocket"
)

type RoomHandler struct {
	roomService *service.RoomService
	hub         *websocket.Hub
}

func NewRoomHandler(roomService *service.RoomService, hub *websocket.Hub) *RoomHandler {
	return &RoomHandler{
		roomService: roomService,
		hub:         hub,
	}
}

func (h *RoomHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req models.CreateRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "Room name is required")
		return
	}

	room, err := h.roomService.Create(r.Context(), claims.UserID, &req)
	if err != nil {
		log.Printf("[RoomHandler.Create] Error: %v, UserID: %d", err, claims.UserID)
		respondError(w, http.StatusInternalServerError, "Failed to create room")
		return
	}

	respondJSON(w, http.StatusCreated, room)
}

func (h *RoomHandler) GetMyRooms(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	rooms, err := h.roomService.GetByUserID(r.Context(), claims.UserID)
	if err != nil {
		log.Printf("[RoomHandler.GetMyRooms] Error: %v, UserID: %d", err, claims.UserID)
		respondError(w, http.StatusInternalServerError, "Failed to get rooms")
		return
	}

	respondJSON(w, http.StatusOK, rooms)
}

func (h *RoomHandler) GetByID(w http.ResponseWriter, r *http.Request) {
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

	room, err := h.roomService.GetByID(r.Context(), roomID, claims.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondError(w, http.StatusNotFound, "Room not found")
			return
		}
		if errors.Is(err, service.ErrNotMember) {
			respondError(w, http.StatusForbidden, "You are not a member of this room")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to get room")
		return
	}

	respondJSON(w, http.StatusOK, room)
}

func (h *RoomHandler) Update(w http.ResponseWriter, r *http.Request) {
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

	var req models.UpdateRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	room, err := h.roomService.Update(r.Context(), roomID, claims.UserID, &req)
	if err != nil {
		if errors.Is(err, service.ErrNotOwner) {
			respondError(w, http.StatusForbidden, "Only room owner can update")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to update room")
		return
	}

	respondJSON(w, http.StatusOK, room)
}

func (h *RoomHandler) Delete(w http.ResponseWriter, r *http.Request) {
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

	if err := h.roomService.Delete(r.Context(), roomID, claims.UserID); err != nil {
		if errors.Is(err, service.ErrNotOwner) {
			respondError(w, http.StatusForbidden, "Only room owner can delete")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to delete room")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *RoomHandler) GetMembers(w http.ResponseWriter, r *http.Request) {
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

	members, err := h.roomService.GetMembers(r.Context(), roomID, claims.UserID)
	if err != nil {
		if errors.Is(err, service.ErrNotMember) {
			respondError(w, http.StatusForbidden, "You are not a member of this room")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to get members")
		return
	}

	respondJSON(w, http.StatusOK, members)
}

func (h *RoomHandler) AddMember(w http.ResponseWriter, r *http.Request) {
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

	var req models.AddMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.roomService.AddMember(r.Context(), roomID, claims.UserID, req.UserID); err != nil {
		if errors.Is(err, service.ErrNotMember) {
			respondError(w, http.StatusForbidden, "You are not a member of this room")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to add member")
		return
	}

	// Get room info and send WebSocket notification to invited user
	room, err := h.roomService.GetByID(r.Context(), roomID, claims.UserID)
	if err == nil && h.hub != nil {
		h.hub.SendRoomInvite(req.UserID, room)
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Member added successfully"})
}

func (h *RoomHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
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

	userID, err := strconv.ParseUint(vars["userId"], 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	if err := h.roomService.RemoveMember(r.Context(), roomID, claims.UserID, userID); err != nil {
		if errors.Is(err, service.ErrNotOwner) {
			respondError(w, http.StatusForbidden, "Only room owner can remove members")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to remove member")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *RoomHandler) Leave(w http.ResponseWriter, r *http.Request) {
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

	if err := h.roomService.Leave(r.Context(), roomID, claims.UserID); err != nil {
		if errors.Is(err, service.ErrOwnerCannotLeave) {
			respondError(w, http.StatusBadRequest, "Owner cannot leave the room. Delete it instead.")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to leave room")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Left room successfully"})
}
