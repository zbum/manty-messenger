package websocket

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"

	"Mmessenger/internal/models"
	"Mmessenger/internal/repository"
	"Mmessenger/internal/service"
	"Mmessenger/pkg/jwt"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

type Handler struct {
	hub            *Hub
	jwtService     *jwt.Service
	messageService *service.MessageService
	memberRepo     *repository.RoomMemberRepository
	userRepo       *repository.UserRepository
	roomRepo       *repository.RoomRepository
}

func NewHandler(hub *Hub, jwtService *jwt.Service, messageService *service.MessageService, memberRepo *repository.RoomMemberRepository, userRepo *repository.UserRepository, roomRepo *repository.RoomRepository) *Handler {
	return &Handler{
		hub:            hub,
		jwtService:     jwtService,
		messageService: messageService,
		memberRepo:     memberRepo,
		userRepo:       userRepo,
		roomRepo:       roomRepo,
	}
}

func (h *Handler) ServeWS(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Token required", http.StatusUnauthorized)
		return
	}

	claims, err := h.jwtService.ValidateToken(token)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("websocket upgrade error: %v", err)
		return
	}

	client := NewClient(h.hub, conn, claims.UserID, claims.Username, h)
	h.hub.register <- client

	// Broadcast online status
	h.hub.BroadcastPresence(claims.UserID, "online")

	go client.WritePump()
	go client.ReadPump()
}

func (h *Handler) HandleMessage(client *Client, msg *WSMessage) {
	switch msg.Type {
	case TypeJoinRoom:
		h.handleJoinRoom(client, msg)
	case TypeLeaveRoom:
		h.handleLeaveRoom(client, msg)
	case TypeSendMessage:
		h.handleSendMessage(client, msg)
	case TypeTyping:
		h.handleTyping(client, msg)
	case TypeMarkRead:
		h.handleMarkRead(client, msg)
	case TypePing:
		h.handlePing(client)
	default:
		client.sendError("UNKNOWN_TYPE", "Unknown message type", msg.RequestID)
	}
}

func (h *Handler) handleJoinRoom(client *Client, msg *WSMessage) {
	payloadBytes, _ := json.Marshal(msg.Payload)
	var payload JoinRoomPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		client.sendError("INVALID_PAYLOAD", "Invalid payload", msg.RequestID)
		return
	}

	// Check if user is a member of the room
	isMember, err := h.memberRepo.IsMember(context.Background(), payload.RoomID, client.UserID)
	if err != nil || !isMember {
		client.sendError("NOT_MEMBER", "You are not a member of this room", msg.RequestID)
		return
	}

	h.hub.JoinRoom(client, payload.RoomID)

	// Send confirmation
	client.Send(&WSMessage{
		Type: TypeRoomJoined,
		Payload: RoomJoinedPayload{
			RoomID: payload.RoomID,
		},
		Timestamp: time.Now(),
	})

	// Get member count
	memberCount, _ := h.roomRepo.GetMemberCount(context.Background(), payload.RoomID)

	// Notify other members
	user, _ := h.userRepo.GetByID(context.Background(), client.UserID)
	notification := &WSMessage{
		Type: TypeUserJoined,
		Payload: UserJoinedPayload{
			RoomID:      payload.RoomID,
			User:        user.ToResponse(),
			MemberCount: memberCount,
		},
		Timestamp: time.Now(),
	}

	if data, err := marshalMessage(notification); err == nil {
		h.hub.BroadcastToRoom(payload.RoomID, data, nil) // nil로 변경하여 본인 포함 모두에게 전송
	}
}

func (h *Handler) handleLeaveRoom(client *Client, msg *WSMessage) {
	payloadBytes, _ := json.Marshal(msg.Payload)
	var payload LeaveRoomPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		client.sendError("INVALID_PAYLOAD", "Invalid payload", msg.RequestID)
		return
	}

	h.hub.LeaveRoom(client, payload.RoomID)

	// Send confirmation
	client.Send(&WSMessage{
		Type: TypeRoomLeft,
		Payload: RoomLeftPayload{
			RoomID: payload.RoomID,
		},
		Timestamp: time.Now(),
	})

	// Get member count
	memberCount, _ := h.roomRepo.GetMemberCount(context.Background(), payload.RoomID)

	// Notify other members
	notification := &WSMessage{
		Type: TypeUserLeft,
		Payload: UserLeftPayload{
			RoomID:      payload.RoomID,
			UserID:      client.UserID,
			Username:    client.Username,
			MemberCount: memberCount,
		},
		Timestamp: time.Now(),
	}

	if data, err := marshalMessage(notification); err == nil {
		h.hub.BroadcastToRoom(payload.RoomID, data, nil) // 본인 포함 모두에게 전송
	}
}

func (h *Handler) handleSendMessage(client *Client, msg *WSMessage) {
	payloadBytes, _ := json.Marshal(msg.Payload)
	var payload SendMessagePayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		client.sendError("INVALID_PAYLOAD", "Invalid payload", msg.RequestID)
		return
	}

	if payload.Content == "" {
		client.sendError("EMPTY_CONTENT", "Message content cannot be empty", msg.RequestID)
		return
	}

	// Save message to database
	req := &models.SendMessageRequest{
		Content:     payload.Content,
		MessageType: payload.MessageType,
	}

	savedMsg, err := h.messageService.Create(context.Background(), payload.RoomID, client.UserID, req)
	if err != nil {
		client.sendError("SEND_FAILED", "Failed to send message", msg.RequestID)
		return
	}

	// Broadcast to room members
	notification := &WSMessage{
		Type: TypeNewMessage,
		Payload: NewMessagePayload{
			ID:          savedMsg.ID,
			RoomID:      payload.RoomID,
			Sender:      savedMsg.Sender,
			Content:     savedMsg.Content,
			MessageType: savedMsg.MessageType,
			CreatedAt:   savedMsg.CreatedAt,
			UnreadCount: savedMsg.UnreadCount,
		},
		Timestamp: time.Now(),
	}

	if data, err := marshalMessage(notification); err == nil {
		// Send to all room members including sender
		h.hub.BroadcastToRoom(payload.RoomID, data, nil)
	}
}

func (h *Handler) handleTyping(client *Client, msg *WSMessage) {
	payloadBytes, _ := json.Marshal(msg.Payload)
	var payload TypingPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return
	}

	notification := &WSMessage{
		Type: TypeUserTyping,
		Payload: UserTypingPayload{
			RoomID:   payload.RoomID,
			UserID:   client.UserID,
			Username: client.Username,
			IsTyping: payload.IsTyping,
		},
		Timestamp: time.Now(),
	}

	if data, err := marshalMessage(notification); err == nil {
		h.hub.BroadcastToRoom(payload.RoomID, data, client)
	}
}

func (h *Handler) handleMarkRead(client *Client, msg *WSMessage) {
	payloadBytes, _ := json.Marshal(msg.Payload)
	var payload MarkReadPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return
	}

	h.memberRepo.UpdateLastRead(context.Background(), payload.RoomID, client.UserID)

	// Broadcast read status to room members
	notification := &WSMessage{
		Type: TypeMessageRead,
		Payload: MessageReadPayload{
			RoomID:   payload.RoomID,
			UserID:   client.UserID,
			Username: client.Username,
		},
		Timestamp: time.Now(),
	}

	if data, err := marshalMessage(notification); err == nil {
		h.hub.BroadcastToRoom(payload.RoomID, data, client)
	}
}

func (h *Handler) handlePing(client *Client) {
	client.Send(&WSMessage{
		Type:      TypePong,
		Timestamp: time.Now(),
	})
}
