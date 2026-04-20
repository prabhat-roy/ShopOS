package domain

import (
	"errors"
	"time"
)

// ReplayStatus represents the lifecycle state of a replay job.
type ReplayStatus string

const (
	StatusPending   ReplayStatus = "pending"
	StatusRunning   ReplayStatus = "running"
	StatusCompleted ReplayStatus = "completed"
	StatusFailed    ReplayStatus = "failed"
	StatusCancelled ReplayStatus = "cancelled"
)

// ReplayTarget represents where replayed events are sent.
type ReplayTarget string

const (
	TargetKafka ReplayTarget = "kafka"
	TargetHTTP  ReplayTarget = "http"
)

// ReplayJob is the aggregate root for an event replay operation.
type ReplayJob struct {
	ID          string       `json:"id"`
	StreamID    string       `json:"stream_id"`
	StreamType  string       `json:"stream_type"`
	EventType   string       `json:"event_type"`
	FromSeq     int64        `json:"from_seq"`
	ToSeq       int64        `json:"to_seq"`
	FromTime    *time.Time   `json:"from_time,omitempty"`
	ToTime      *time.Time   `json:"to_time,omitempty"`
	Target      ReplayTarget `json:"target"`
	TargetTopic string       `json:"target_topic"`
	Status      ReplayStatus `json:"status"`

	EventsReplayed int64  `json:"events_replayed"`
	ErrorMessage   string `json:"error_message,omitempty"`

	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// CreateReplayRequest carries the parameters needed to create a new replay job.
type CreateReplayRequest struct {
	StreamID    string       `json:"stream_id"`
	StreamType  string       `json:"stream_type"`
	EventType   string       `json:"event_type"`
	FromSeq     int64        `json:"from_seq"`
	ToSeq       int64        `json:"to_seq"`
	Target      ReplayTarget `json:"target"`
	TargetTopic string       `json:"target_topic"`
}

// ErrNotFound is returned when a requested replay job does not exist.
var ErrNotFound = errors.New("not found")

// ErrInvalidTransition is returned when a status change is not permitted.
var ErrInvalidTransition = errors.New("invalid status transition")
