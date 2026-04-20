package domain

import (
	"errors"
	"time"
)

// Session status constants.
const (
	StatusWaiting = "waiting"
	StatusActive  = "active"
	StatusClosed  = "closed"
)

// Sender type constants.
const (
	SenderCustomer = "customer"
	SenderAgent    = "agent"
	SenderSystem   = "system"
)

// ErrNotFound is returned when a session or message does not exist.
var ErrNotFound = errors.New("not found")

// ErrSessionClosed is returned when trying to operate on a closed session.
var ErrSessionClosed = errors.New("session is closed")

// ChatSession represents a live-chat session between a customer and an agent.
type ChatSession struct {
	ID         string    `json:"id"`
	CustomerID string    `json:"customerId"`
	AgentID    string    `json:"agentId,omitempty"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// ChatMessage represents a single message within a chat session.
type ChatMessage struct {
	ID         string    `json:"id"`
	SessionID  string    `json:"sessionId"`
	SenderID   string    `json:"senderId"`
	SenderType string    `json:"senderType"`
	Body       string    `json:"body"`
	SentAt     time.Time `json:"sentAt"`
}
