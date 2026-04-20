package domain

import (
	"errors"
	"time"
)

// MessageStatus represents the lifecycle state of a dead-lettered message.
type MessageStatus string

const (
	// StatusPending means the message has been received but not yet acted upon.
	StatusPending MessageStatus = "pending"
	// StatusRetried means the message has been sent back for reprocessing.
	StatusRetried MessageStatus = "retried"
	// StatusDiscarded means the message has been permanently discarded.
	StatusDiscarded MessageStatus = "discarded"
)

// ErrNotFound is returned when a requested message does not exist in the store.
var ErrNotFound = errors.New("not found")

// DeadMessage is the core domain entity representing a failed Kafka message that
// has been routed to a dead-letter queue.
type DeadMessage struct {
	ID          string        `json:"id"`
	Topic       string        `json:"topic"`
	Key         string        `json:"key"`
	Partition   int           `json:"partition"`
	Offset      int64         `json:"offset"`
	Payload     []byte        `json:"payload"`
	ErrorReason string        `json:"error_reason"`
	Status      MessageStatus `json:"status"`
	RetryCount  int           `json:"retry_count"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}
