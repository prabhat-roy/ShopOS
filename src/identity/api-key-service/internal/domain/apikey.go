package domain

import (
	"errors"
	"time"
)

// APIKey represents a stored API key. The raw key is never stored — only the SHA-256 hex digest.
type APIKey struct {
	ID         string
	OwnerID    string     // user or service ID
	OwnerType  string     // "user" | "service"
	Name       string     // human-readable label
	KeyPrefix  string     // first 8 chars of the raw key, safe to display (e.g. "sk_live_")
	KeyHash    string     // SHA-256 hex hash of the full raw key
	Scopes     []string   // e.g. ["catalog:read", "orders:write"]
	Active     bool
	LastUsedAt *time.Time
	ExpiresAt  *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// CreateKeyRequest carries the fields needed to create a new API key.
type CreateKeyRequest struct {
	OwnerID   string
	OwnerType string
	Name      string
	Scopes    []string
	ExpiresAt *time.Time
}

// ValidateRequest carries the raw API key to be validated.
type ValidateRequest struct {
	Key string
}

// ValidateResponse is returned by the validate endpoint.
type ValidateResponse struct {
	Valid   bool
	KeyID   string
	OwnerID string
	Scopes  []string
	Reason  string
}

var ErrNotFound    = errors.New("not found")
var ErrKeyInactive = errors.New("key is inactive or expired")
