package domain

import (
	"errors"
	"time"
)

var (
	ErrNotFound      = errors.New("flag not found")
	ErrAlreadyExists = errors.New("flag already exists")
	ErrInvalidInput  = errors.New("invalid input")
)

// RolloutStrategy controls how a flag is enabled for users/contexts.
type RolloutStrategy string

const (
	StrategyAll        RolloutStrategy = "all"        // enabled for everyone
	StrategyPercentage RolloutStrategy = "percentage" // enabled for X% of requests
	StrategyUserList   RolloutStrategy = "user_list"  // enabled for specific user IDs
	StrategyContext    RolloutStrategy = "context"    // enabled based on context key/value
)

// Flag is a feature flag definition stored in Postgres.
type Flag struct {
	ID          string          `json:"id"`
	Key         string          `json:"key"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Enabled     bool            `json:"enabled"`
	Strategy    RolloutStrategy `json:"strategy"`
	Percentage  int             `json:"percentage,omitempty"`  // 0–100, used with StrategyPercentage
	UserIDs     []string        `json:"user_ids,omitempty"`    // used with StrategyUserList
	ContextKey  string          `json:"context_key,omitempty"` // used with StrategyContext
	ContextVal  string          `json:"context_val,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// EvalRequest carries the context needed to evaluate a flag.
type EvalRequest struct {
	Key       string            // flag key to evaluate
	UserID    string            // caller's user ID (optional)
	Context   map[string]string // arbitrary context key/value pairs
}

// CreateFlagRequest is the payload for creating a new flag.
type CreateFlagRequest struct {
	Key         string          `json:"key"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Enabled     bool            `json:"enabled"`
	Strategy    RolloutStrategy `json:"strategy"`
	Percentage  int             `json:"percentage,omitempty"`
	UserIDs     []string        `json:"user_ids,omitempty"`
	ContextKey  string          `json:"context_key,omitempty"`
	ContextVal  string          `json:"context_val,omitempty"`
}

// UpdateFlagRequest is the payload for updating an existing flag.
type UpdateFlagRequest struct {
	Name        *string          `json:"name,omitempty"`
	Description *string          `json:"description,omitempty"`
	Enabled     *bool            `json:"enabled,omitempty"`
	Strategy    *RolloutStrategy `json:"strategy,omitempty"`
	Percentage  *int             `json:"percentage,omitempty"`
	UserIDs     []string         `json:"user_ids,omitempty"`
	ContextKey  *string          `json:"context_key,omitempty"`
	ContextVal  *string          `json:"context_val,omitempty"`
}
