package domain

import (
	"errors"
	"time"
)

var (
	ErrNotFound    = errors.New("not found")
	ErrInvalidURL  = errors.New("invalid webhook URL")
)

type Webhook struct {
	ID        string    `json:"id"`
	OwnerID   string    `json:"owner_id"`
	URL       string    `json:"url"`
	Events    []string  `json:"events"`   // e.g. ["commerce.order.placed"]
	Secret    string    `json:"secret,omitempty"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateWebhookRequest struct {
	OwnerID string   `json:"owner_id"`
	URL     string   `json:"url"`
	Events  []string `json:"events"`
	Secret  string   `json:"secret"`
}

type UpdateWebhookRequest struct {
	URL    *string  `json:"url"`
	Events []string `json:"events"`
	Active *bool    `json:"active"`
}

// Delivery records a single HTTP dispatch attempt.
type Delivery struct {
	ID         string    `json:"id"`
	WebhookID  string    `json:"webhook_id"`
	EventTopic string    `json:"event_topic"`
	Payload    []byte    `json:"payload"`
	StatusCode int       `json:"status_code"`
	Attempt    int       `json:"attempt"`
	Success    bool      `json:"success"`
	CreatedAt  time.Time `json:"created_at"`
}
