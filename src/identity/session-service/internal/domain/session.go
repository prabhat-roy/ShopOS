package domain

import (
	"errors"
	"time"
)

// Session represents an authenticated user session.
type Session struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	DeviceInfo   string    `json:"device_info"`
	IPAddress    string    `json:"ip_address"`
	UserAgent    string    `json:"user_agent"`
	CreatedAt    time.Time `json:"created_at"`
	LastActiveAt time.Time `json:"last_active_at"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// CreateSessionRequest is the payload required to open a new session.
type CreateSessionRequest struct {
	UserID     string `json:"user_id"`
	DeviceInfo string `json:"device_info"`
	IPAddress  string `json:"ip_address"`
	UserAgent  string `json:"user_agent"`
}

// Sentinel errors returned by the store and service layers.
var (
	ErrNotFound = errors.New("session not found")
	ErrExpired  = errors.New("session expired")
)
