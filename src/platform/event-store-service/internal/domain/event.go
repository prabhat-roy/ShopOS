package domain

import (
	"errors"
	"time"
)

var (
	ErrNotFound        = errors.New("event not found")
	ErrInvalidInput    = errors.New("invalid input")
	ErrVersionConflict = errors.New("optimistic concurrency conflict: expected version mismatch")
)

// Event is an immutable fact that something happened in the system.
// Events are append-only and never updated or deleted.
type Event struct {
	ID            string    `json:"id"`
	StreamID      string    `json:"stream_id"`      // aggregate identity (e.g. "order-123")
	StreamType    string    `json:"stream_type"`    // aggregate type (e.g. "Order")
	EventType     string    `json:"event_type"`     // e.g. "OrderPlaced", "OrderCancelled"
	Version       int64     `json:"version"`        // monotonically increasing per stream
	GlobalSeq     int64     `json:"global_seq"`     // global append-order sequence
	Payload       []byte    `json:"payload"`        // JSON-encoded event data
	Metadata      []byte    `json:"metadata"`       // JSON-encoded meta (correlation_id, causation_id, etc.)
	OccurredAt    time.Time `json:"occurred_at"`
	RecordedAt    time.Time `json:"recorded_at"`
}

// AppendRequest carries one or more events to append to a stream.
type AppendRequest struct {
	StreamID        string   `json:"stream_id"`
	StreamType      string   `json:"stream_type"`
	Events          []NewEvent `json:"events"`
	ExpectedVersion int64    `json:"expected_version"` // -1 = any, 0 = new stream
}

// NewEvent is the input shape for a single event to be appended.
type NewEvent struct {
	EventType  string `json:"event_type"`
	Payload    []byte `json:"payload"`
	Metadata   []byte `json:"metadata,omitempty"`
	OccurredAt *time.Time `json:"occurred_at,omitempty"`
}

// ReadRequest specifies how to read events from a stream.
type ReadRequest struct {
	StreamID      string `json:"stream_id"`
	FromVersion   int64  `json:"from_version"`    // inclusive
	ToVersion     int64  `json:"to_version"`      // 0 = unbounded
	MaxCount      int    `json:"max_count"`       // 0 = use server default
}

// ReadAllRequest reads events globally ordered by global_seq.
type ReadAllRequest struct {
	FromGlobalSeq int64  `json:"from_global_seq"` // inclusive
	MaxCount      int    `json:"max_count"`
	StreamType    string `json:"stream_type,omitempty"` // optional filter
	EventType     string `json:"event_type,omitempty"`  // optional filter
}

// Snapshot captures the projected state of an aggregate at a given version.
type Snapshot struct {
	StreamID   string    `json:"stream_id"`
	StreamType string    `json:"stream_type"`
	Version    int64     `json:"version"`
	State      []byte    `json:"state"`
	CreatedAt  time.Time `json:"created_at"`
}
