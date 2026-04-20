package domain

import (
	"errors"
	"time"
)

var (
	ErrNotFound     = errors.New("policy not found")
	ErrInvalidInput = errors.New("invalid input")
	ErrRateLimited  = errors.New("rate limit exceeded")
)

// Algorithm is the rate-limiting algorithm to use.
type Algorithm string

const (
	AlgoTokenBucket   Algorithm = "token_bucket"
	AlgoSlidingWindow Algorithm = "sliding_window"
	AlgoFixedWindow   Algorithm = "fixed_window"
)

// Policy defines a named rate-limit rule.
type Policy struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Key         string    `json:"key"`    // e.g. "service:endpoint" or "user:*"
	Algorithm   Algorithm `json:"algorithm"`
	Limit       int       `json:"limit"`        // max requests per window
	WindowSecs  int       `json:"window_secs"`  // window size in seconds
	BurstSize   int       `json:"burst_size"`   // token bucket burst (0 = same as Limit)
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CheckRequest is the payload for a rate-limit check.
type CheckRequest struct {
	PolicyKey string `json:"policy_key"` // matches Policy.Key
	Subject   string `json:"subject"`    // e.g. IP address, user ID, API key
	Cost      int    `json:"cost"`       // tokens to consume (default 1)
}

// CheckResponse is the result of a rate-limit check.
type CheckResponse struct {
	Allowed    bool  `json:"allowed"`
	Remaining  int64 `json:"remaining"`   // tokens/requests remaining in window
	ResetAfter int64 `json:"reset_after"` // seconds until window resets
	RetryAfter int64 `json:"retry_after"` // seconds to wait if not allowed
}

// CreatePolicyRequest is the payload for creating a new policy.
type CreatePolicyRequest struct {
	Name       string    `json:"name"`
	Key        string    `json:"key"`
	Algorithm  Algorithm `json:"algorithm"`
	Limit      int       `json:"limit"`
	WindowSecs int       `json:"window_secs"`
	BurstSize  int       `json:"burst_size"`
	Enabled    bool      `json:"enabled"`
}

// UpdatePolicyRequest carries partial updates for a policy.
type UpdatePolicyRequest struct {
	Enabled    *bool  `json:"enabled,omitempty"`
	Limit      *int   `json:"limit,omitempty"`
	WindowSecs *int   `json:"window_secs,omitempty"`
	BurstSize  *int   `json:"burst_size,omitempty"`
}
