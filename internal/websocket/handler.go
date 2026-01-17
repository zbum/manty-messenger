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
	"Mmessenger/pkg/keycloak"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

type Handler struct {
	hub             *Hub
	keycloakService *keycloak.Service
	authService     *service.AuthService
	messageService  *service.MessageService
	pushService     *service.PushService
	memberRepo      *repository.RoomMemberRepository
	userRepo        *repository.UserRepository
	roomRepo        *repository.RoomRepository
}

func NewHandler(hub *Hub, keycloakService *keycloak.Service, authService *service.AuthService, messageService *service.MessageService, pushService *service.PushService, memberRepo *repository.RoomMemberRepository, userRepo *repository.UserRepository, roomRepo *repository.RoomRepository) *Handler {
	return &Handler{
		hub:             hub,
		keycloakService: keycloakService,
		authService:     authService,
		messageService:  messageService,
		pushService:     pushService,
		memberRepo:      memberRepo,
		userRepo:        userRepo,
		roomRepo:        roomRepo,
	}
}

func (h *Handler) ServeWS(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Token required", http.StatusUnauthorized)
		return
	}

	keycloakClaims, err := h.keycloakService.ValidateToken(token)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	user, err := h.authService.GetOrCreateUserFromKeycloak(r.Context(), keycloakClaims)
	if err != nil {
		http.Error(w, "Failed to lookup user", http.StatusInternalServerError)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("websocket upgrade error: %v", err)
		return
	}

	client := NewClient(h.hub, conn, user.ID, user.Username, h)
	h.hub.register <- client

	// Broadcast online status
	h.hub.BroadcastPresence(user.ID, "online")

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

	if payload.Content == "" && payload.FileURL == "" {
		client.sendError("EMPTY_CONTENT", "Message content cannot be empty", msg.RequestID)
		return
	}

	// Save message to database
	req := &models.SendMessageRequest{
		Content:      payload.Content,
		MessageType:  payload.MessageType,
		FileURL:      payload.FileURL,
		ThumbnailURL: payload.ThumbnailURL,
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
			ID:           savedMsg.ID,
			RoomID:       payload.RoomID,
			Sender:       savedMsg.Sender,
			Content:      savedMsg.Content,
			MessageType:  savedMsg.MessageType,
			FileURL:      savedMsg.FileURL,
			ThumbnailURL: savedMsg.ThumbnailURL,
			CreatedAt:    savedMsg.CreatedAt,
			UnreadCount:  savedMsg.UnreadCount,
		},
		Timestamp: time.Now(),
	}

	if data, err := marshalMessage(notification); err == nil {
		// Send to all room members including sender
		h.hub.BroadcastToRoom(payload.RoomID, data, nil)
	}

	// Send push notification to offline users
	if h.pushService != nil {
		go func() {
			room, err := h.roomRepo.GetByID(context.Background(), payload.RoomID)
			if err != nil {
				return
			}

			senderName := client.Username
			if savedMsg.Sender != nil && savedMsg.Sender.Username != "" {
				senderName = savedMsg.Sender.Username
			}

			content := savedMsg.Content
			if savedMsg.MessageType == "image" {
				content = "📷 이미지를 보냈습니다"
			} else if savedMsg.MessageType == "file" {
				content = "📎 파일을 보냈습니다"
			} else if len(content) > 100 {
				content = content[:100] + "..."
			}

			pushNotif := &models.PushNotification{
				Title: senderName + " - " + room.Name,
				Body:  content,
				Icon:  "/favicon.ico",
				Tag:   "message-" + string(rune(payload.RoomID)),
				Data: map[string]any{
					"roomId":    payload.RoomID,
					"messageId": savedMsg.ID,
				},
			}

			h.pushService.SendToRoomMembers(context.Background(), payload.RoomID, client.UserID, pushNotif)
		}()
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
