package models

import (
	"database/sql"
	"time"
)

type RoomType string

const (
	RoomTypePrivate RoomType = "private"
	RoomTypeGroup   RoomType = "group"
)

type Room struct {
	ID          uint64         `json:"id"`
	Name        string         `json:"name"`
	Description sql.NullString `json:"description"`
	RoomType    RoomType       `json:"room_type"`
	OwnerID     uint64         `json:"owner_id"`
	AvatarURL   sql.NullString `json:"avatar_url"`
	MaxMembers  int            `json:"max_members"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

type RoomResponse struct {
	ID          uint64    `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	RoomType    RoomType  `json:"room_type"`
	OwnerID     uint64    `json:"owner_id"`
	AvatarURL   *string   `json:"avatar_url"`
	MaxMembers  int       `json:"max_members"`
	MemberCount int       `json:"member_count,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

func (r *Room) ToResponse() *RoomResponse {
	var description, avatarURL *string
	if r.Description.Valid {
		description = &r.Description.String
	}
	if r.AvatarURL.Valid {
		avatarURL = &r.AvatarURL.String
	}

	return &RoomResponse{
		ID:          r.ID,
		Name:        r.Name,
		Description: description,
		RoomType:    r.RoomType,
		OwnerID:     r.OwnerID,
		AvatarURL:   avatarURL,
		MaxMembers:  r.MaxMembers,
		CreatedAt:   r.CreatedAt,
	}
}

// Getter methods for RoomResponse (used by websocket hub)
func (r *RoomResponse) GetID() uint64          { return r.ID }
func (r *RoomResponse) GetName() string        { return r.Name }
func (r *RoomResponse) GetDescription() *string { return r.Description }
func (r *RoomResponse) GetRoomType() string    { return string(r.RoomType) }
func (r *RoomResponse) GetMemberCount() int    { return r.MemberCount }

type CreateRoomRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	RoomType    RoomType `json:"room_type"`
	MemberIDs   []uint64 `json:"member_ids,omitempty"`
}

type UpdateRoomRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}
