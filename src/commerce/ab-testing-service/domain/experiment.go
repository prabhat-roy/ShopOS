package domain

import (
	"errors"
	"time"
)

// ErrNotFound is returned when a requested resource does not exist.
var ErrNotFound = errors.New("not found")

// Experiment represents an A/B test experiment.
type Experiment struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	Variants       []Variant `json:"variants"`
	Active         bool      `json:"active"`
	TrafficPercent int       `json:"traffic_percent"` // 1-100
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// Variant is one arm of an experiment.
type Variant struct {
	Name   string            `json:"name"`
	Weight int               `json:"weight"` // relative weight; sum of all weights = 100
	Config map[string]string `json:"config"`
}

// Assignment records which variant a user was placed in.
type Assignment struct {
	ExperimentID string    `json:"experiment_id"`
	UserID       string    `json:"user_id"`
	Variant      string    `json:"variant"`
	AssignedAt   time.Time `json:"assigned_at"`
}

// Conversion records a metric event for an assigned user.
type Conversion struct {
	ExperimentID string    `json:"experiment_id"`
	UserID       string    `json:"user_id"`
	Metric       string    `json:"metric"`
	Value        float64   `json:"value"`
	RecordedAt   time.Time `json:"recorded_at"`
}
