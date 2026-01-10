package models

import (
	"database/sql"
	"time"
)

type MemberRole string

const (
	MemberRoleOwner  MemberRole = "owner"
	MemberRoleAdmin  MemberRole = "admin"
	MemberRoleMember MemberRole = "member"
)

type RoomMember struct {
	ID         uint64       `json:"id"`
	RoomID     uint64       `json:"room_id"`
	UserID     uint64       `json:"user_id"`
	Role       MemberRole   `json:"role"`
	JoinedAt   time.Time    `json:"joined_at"`
	LastReadAt sql.NullTime `json:"last_read_at"`
}

type RoomMemberResponse struct {
	User     *UserResponse `json:"user"`
	Role     MemberRole    `json:"role"`
	JoinedAt time.Time     `json:"joined_at"`
}

type AddMemberRequest struct {
	UserID uint64 `json:"user_id"`
}
