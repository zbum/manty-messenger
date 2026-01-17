package models

import (
	"database/sql"
	"time"
)

type UserStatus string

const (
	UserStatusOnline  UserStatus = "online"
	UserStatusOffline UserStatus = "offline"
	UserStatusAway    UserStatus = "away"
)

type User struct {
	ID           uint64         `json:"id"`
	KeycloakID   sql.NullString `json:"keycloak_id,omitempty"`
	Email        string         `json:"email"`
	Username     string         `json:"username"`
	PasswordHash string         `json:"-"`
	AvatarURL    sql.NullString `json:"avatar_url"`
	Status       UserStatus     `json:"status"`
	LastSeenAt   sql.NullTime   `json:"last_seen_at"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

type UserResponse struct {
	ID        uint64     `json:"id"`
	Email     string     `json:"email"`
	Username  string     `json:"username"`
	AvatarURL *string    `json:"avatar_url"`
	Status    UserStatus `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
}

func (u *User) ToResponse() *UserResponse {
	var avatarURL *string
	if u.AvatarURL.Valid {
		avatarURL = &u.AvatarURL.String
	}

	return &UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Username:  u.Username,
		AvatarURL: avatarURL,
		Status:    u.Status,
		CreatedAt: u.CreatedAt,
	}
}

type CreateUserRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	User         *UserResponse `json:"user"`
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token,omitempty"`
}
