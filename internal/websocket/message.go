package websocket

import (
	"time"

	"Mmessenger/internal/models"
)

type MessageType string

const (
	// Client -> Server
	TypeJoinRoom    MessageType = "join_room"
	TypeLeaveRoom   MessageType = "leave_room"
	TypeSendMessage MessageType = "send_message"
	TypeTyping      MessageType = "typing"
	TypeMarkRead    MessageType = "mark_read"
	TypePing        MessageType = "ping"

	// Server -> Client
	TypeNewMessage     MessageType = "new_message"
	TypeMessageRead    MessageType = "message_read"
	TypeUserJoined     MessageType = "user_joined"
	TypeUserLeft       MessageType = "user_left"
	TypeUserTyping     MessageType = "user_typing"
	TypePresenceUpdate MessageType = "presence_update"
	TypeError          MessageType = "error"
	TypePong           MessageType = "pong"
	TypeRoomJoined     MessageType = "room_joined"
	TypeRoomLeft       MessageType = "room_left"
	TypeRoomInvited    MessageType = "room_invited"
)

type RoomInvitedPayload struct {
	Room *RoomInfo `json:"room"`
}

type RoomInfo struct {
	ID          uint64  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	RoomType    string  `json:"room_type"`
	MemberCount int     `json:"member_count"`
}

type WSMessage struct {
	Type      MessageType `json:"type"`
	Payload   interface{} `json:"payload,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
	RequestID string      `json:"request_id,omitempty"`
}

// Payload types for client messages
type JoinRoomPayload struct {
	RoomID uint64 `json:"room_id"`
}

type LeaveRoomPayload struct {
	RoomID uint64 `json:"room_id"`
}

type SendMessagePayload struct {
	RoomID       uint64             `json:"room_id"`
	Content      string             `json:"content"`
	MessageType  models.MessageType `json:"message_type"`
	FileURL      string             `json:"file_url,omitempty"`
	ThumbnailURL string             `json:"thumbnail_url,omitempty"`
}

type TypingPayload struct {
	RoomID   uint64 `json:"room_id"`
	IsTyping bool   `json:"is_typing"`
}

type MarkReadPayload struct {
	RoomID    uint64 `json:"room_id"`
	MessageID uint64 `json:"message_id"`
}

// Payload types for server messages
type NewMessagePayload struct {
	ID           uint64               `json:"id"`
	RoomID       uint64               `json:"room_id"`
	Sender       *models.UserResponse `json:"sender"`
	Content      string               `json:"content"`
	MessageType  models.MessageType   `json:"message_type"`
	FileURL      *string              `json:"file_url,omitempty"`
	ThumbnailURL *string              `json:"thumbnail_url,omitempty"`
	CreatedAt    time.Time            `json:"created_at"`
	UnreadCount  int                  `json:"unread_count"`
}

type MessageReadPayload struct {
	RoomID      uint64 `json:"room_id"`
	UserID      uint64 `json:"user_id"`
	Username    string `json:"username"`
}

type UserJoinedPayload struct {
	RoomID      uint64               `json:"room_id"`
	User        *models.UserResponse `json:"user"`
	MemberCount int                  `json:"member_count"`
}

type UserLeftPayload struct {
	RoomID      uint64 `json:"room_id"`
	UserID      uint64 `json:"user_id"`
	Username    string `json:"username"`
	MemberCount int    `json:"member_count"`
}

type UserTypingPayload struct {
	RoomID   uint64 `json:"room_id"`
	UserID   uint64 `json:"user_id"`
	Username string `json:"username"`
	IsTyping bool   `json:"is_typing"`
}

type PresenceUpdatePayload struct {
	UserID uint64            `json:"user_id"`
	Status models.UserStatus `json:"status"`
}

type ErrorPayload struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
}

type RoomJoinedPayload struct {
	RoomID uint64 `json:"room_id"`
}

type RoomLeftPayload struct {
	RoomID uint64 `json:"room_id"`
}
