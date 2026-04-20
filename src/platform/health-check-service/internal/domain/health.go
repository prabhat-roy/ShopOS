package domain

import "time"

// Status represents the health of a single target.
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusUnhealthy Status = "unhealthy"
	StatusUnknown   Status = "unknown"
)

// TargetHealth holds the latest probe result for one service.
type TargetHealth struct {
	Name        string        `json:"name"`
	URL         string        `json:"url"`
	Status      Status        `json:"status"`
	StatusCode  int           `json:"status_code,omitempty"`
	Latency     time.Duration `json:"latency_ms"`
	Message     string        `json:"message,omitempty"`
	Failures    int           `json:"consecutive_failures"`
	LastChecked time.Time     `json:"last_checked"`
}

// OverallHealth is the aggregated view of all targets.
type OverallHealth struct {
	Status  Status                  `json:"status"` // healthy only if ALL targets healthy
	Targets map[string]TargetHealth `json:"targets"`
	CheckedAt time.Time             `json:"checked_at"`
}
