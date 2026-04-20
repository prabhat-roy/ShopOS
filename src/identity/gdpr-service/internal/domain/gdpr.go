package domain

import (
	"errors"
	"time"
)

// RequestType represents the category of a GDPR data subject request.
type RequestType string

const (
	RequestExport          RequestType = "export"
	RequestErasure         RequestType = "erasure"
	RequestRectification   RequestType = "rectification"
)

// RequestStatus tracks the lifecycle of a data subject request.
type RequestStatus string

const (
	StatusPending    RequestStatus = "pending"
	StatusProcessing RequestStatus = "processing"
	StatusCompleted  RequestStatus = "completed"
	StatusRejected   RequestStatus = "rejected"
)

// DataRequest represents a GDPR data subject request (Art. 15, 17, 16 GDPR).
type DataRequest struct {
	ID          string        `json:"id"`
	UserID      string        `json:"user_id"`
	Type        RequestType   `json:"type"`
	Status      RequestStatus `json:"status"`
	Reason      string        `json:"reason"`
	Notes       string        `json:"notes"`
	CompletedAt *time.Time    `json:"completed_at,omitempty"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

// ConsentType identifies a category of data processing consent.
type ConsentType string

const (
	ConsentMarketing  ConsentType = "marketing"
	ConsentAnalytics  ConsentType = "analytics"
	ConsentNecessary  ConsentType = "necessary"
)

// Consent records a user's consent decision for a specific processing purpose.
type Consent struct {
	UserID    string      `json:"user_id"`
	Type      ConsentType `json:"type"`
	Granted   bool        `json:"granted"`
	IPAddress string      `json:"ip_address"`
	UpdatedAt time.Time   `json:"updated_at"`
}

// ErrNotFound is returned when a requested resource does not exist.
var ErrNotFound = errors.New("not found")
