package models

import "time"

type DeviceToken struct {
	ID        uint64    `json:"id"`
	UserID    uint64    `json:"user_id"`
	DeviceID  string    `json:"device_id"`
	Platform  string    `json:"platform"`
	Token     string    `json:"token"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type RegisterDeviceRequest struct {
	DeviceID string `json:"device_id"`
	Platform string `json:"platform"`
	Token    string `json:"token"`
}
