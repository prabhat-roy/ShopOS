package domain

import (
	"errors"
	"time"
)

// Priority represents the processing priority of a job.
type Priority string

const (
	PriorityHigh   Priority = "high"
	PriorityNormal Priority = "normal"
	PriorityLow    Priority = "low"
)

// JobStatus represents the current lifecycle state of a job.
type JobStatus string

const (
	StatusPending   JobStatus = "pending"
	StatusRunning   JobStatus = "running"
	StatusCompleted JobStatus = "completed"
	StatusFailed    JobStatus = "failed"
	StatusDead      JobStatus = "dead" // exceeded max retries
)

// Job is the core domain entity representing a unit of work.
type Job struct {
	ID          string     `json:"id"`
	Queue       string     `json:"queue"`
	Priority    Priority   `json:"priority"`
	Payload     []byte     `json:"payload"`
	CallbackURL string     `json:"callback_url"`
	MaxRetries  int        `json:"max_retries"`
	Attempt     int        `json:"attempt"`
	Status      JobStatus  `json:"status"`
	Error       string     `json:"error,omitempty"`
	EnqueuedAt  time.Time  `json:"enqueued_at"`
	ProcessedAt *time.Time `json:"processed_at,omitempty"`
}

// EnqueueRequest carries the data needed to create and enqueue a new job.
type EnqueueRequest struct {
	Queue       string   `json:"queue"`
	Priority    Priority `json:"priority"`
	Payload     []byte   `json:"payload"`
	CallbackURL string   `json:"callback_url"`
	MaxRetries  int      `json:"max_retries"`
}

// Sentinel errors returned by queue and handler layers.
var (
	ErrNotFound   = errors.New("not found")
	ErrQueueEmpty = errors.New("queue empty")
)

// IsValidPriority reports whether p is one of the recognised priority values.
func IsValidPriority(p Priority) bool {
	switch p {
	case PriorityHigh, PriorityNormal, PriorityLow:
		return true
	}
	return false
}

// IsValidStatus reports whether s is one of the recognised status values.
func IsValidStatus(s JobStatus) bool {
	switch s {
	case StatusPending, StatusRunning, StatusCompleted, StatusFailed, StatusDead:
		return true
	}
	return false
}
