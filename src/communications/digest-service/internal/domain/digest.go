package domain

import (
	"time"

	"github.com/google/uuid"
)

// DigestFrequency controls how often a digest email is sent.
type DigestFrequency string

const (
	FrequencyDaily  DigestFrequency = "DAILY"
	FrequencyWeekly DigestFrequency = "WEEKLY"
)

// DigestStatus reflects whether a digest configuration is active or paused.
type DigestStatus string

const (
	StatusActive DigestStatus = "ACTIVE"
	StatusPaused DigestStatus = "PAUSED"
)

// DigestConfig represents a user's digest subscription configuration.
type DigestConfig struct {
	ID         uuid.UUID       `json:"id"`
	UserID     uuid.UUID       `json:"userId"`
	Email      string          `json:"email"`
	Frequency  DigestFrequency `json:"frequency"`
	Status     DigestStatus    `json:"status"`
	LastSentAt *time.Time      `json:"lastSentAt,omitempty"`
	NextSendAt time.Time       `json:"nextSendAt"`
	Timezone   string          `json:"timezone"`
	CreatedAt  time.Time       `json:"createdAt"`
	UpdatedAt  time.Time       `json:"updatedAt"`
}

// DigestRun records a single execution of a digest send.
type DigestRun struct {
	ID        uuid.UUID `json:"id"`
	ConfigID  uuid.UUID `json:"configId"`
	SentAt    time.Time `json:"sentAt"`
	ItemCount int       `json:"itemCount"`
	Status    string    `json:"status"`
	ErrorMsg  string    `json:"errorMsg,omitempty"`
}

// CreateConfigRequest is the payload for creating a new digest config.
type CreateConfigRequest struct {
	UserID    string          `json:"userId"`
	Email     string          `json:"email"`
	Frequency DigestFrequency `json:"frequency"`
	Timezone  string          `json:"timezone"`
}

// Validate returns a non-empty error message if the request is invalid.
func (r *CreateConfigRequest) Validate() string {
	if r.UserID == "" {
		return "userId is required"
	}
	if _, err := uuid.Parse(r.UserID); err != nil {
		return "userId must be a valid UUID"
	}
	if r.Email == "" {
		return "email is required"
	}
	switch r.Frequency {
	case FrequencyDaily, FrequencyWeekly:
		// valid
	default:
		return "frequency must be DAILY or WEEKLY"
	}
	if r.Timezone == "" {
		r.Timezone = "UTC"
	}
	return ""
}
