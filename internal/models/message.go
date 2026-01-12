package models

import (
	"database/sql"
	"time"
)

type MessageType string

const (
	MessageTypeText   MessageType = "text"
	MessageTypeImage  MessageType = "image"
	MessageTypeFile   MessageType = "file"
	MessageTypeSystem MessageType = "system"
)

type Message struct {
	ID          uint64         `json:"id"`
	RoomID      uint64         `json:"room_id"`
	SenderID    uint64         `json:"sender_id"`
	Content     string         `json:"content"`
	MessageType MessageType    `json:"message_type"`
	FileURL     sql.NullString `json:"file_url"`
	IsEdited    bool           `json:"is_edited"`
	IsDeleted   bool           `json:"is_deleted"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

type MessageResponse struct {
	ID          uint64        `json:"id"`
	RoomID      uint64        `json:"room_id"`
	Sender      *UserResponse `json:"sender"`
	Content     string        `json:"content"`
	MessageType MessageType   `json:"message_type"`
	FileURL     *string       `json:"file_url,omitempty"`
	IsEdited    bool          `json:"is_edited"`
	CreatedAt   time.Time     `json:"created_at"`
	UnreadCount int           `json:"unread_count"`
}

func (m *Message) ToResponse(sender *UserResponse, unreadCount int) *MessageResponse {
	var fileURL *string
	if m.FileURL.Valid {
		fileURL = &m.FileURL.String
	}

	content := m.Content
	if m.IsDeleted {
		content = "This message has been deleted"
	}

	return &MessageResponse{
		ID:          m.ID,
		RoomID:      m.RoomID,
		Sender:      sender,
		Content:     content,
		MessageType: m.MessageType,
		FileURL:     fileURL,
		IsEdited:    m.IsEdited,
		CreatedAt:   m.CreatedAt,
		UnreadCount: unreadCount,
	}
}

type SendMessageRequest struct {
	Content     string      `json:"content"`
	MessageType MessageType `json:"message_type"`
	FileURL     string      `json:"file_url,omitempty"`
}

type UpdateMessageRequest struct {
	Content string `json:"content"`
}
