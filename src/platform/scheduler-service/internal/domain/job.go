package domain

import (
	"errors"
	"time"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrInvalidCron   = errors.New("invalid cron expression")
	ErrJobDisabled   = errors.New("job is disabled")
)

type JobStatus string

const (
	StatusEnabled  JobStatus = "enabled"
	StatusDisabled JobStatus = "disabled"
)

type RunStatus string

const (
	RunSuccess RunStatus = "success"
	RunFailed  RunStatus = "failed"
)

// Job is a scheduled task identified by a cron expression.
type Job struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CronExpr    string    `json:"cron_expr"`
	HTTPMethod  string    `json:"http_method"` // GET, POST, etc.
	HTTPURL     string    `json:"http_url"`
	HTTPBody    string    `json:"http_body,omitempty"`
	Status      JobStatus `json:"status"`
	NextRunAt   time.Time `json:"next_run_at"`
	LastRunAt   *time.Time `json:"last_run_at,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateJobRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	CronExpr    string `json:"cron_expr"`
	HTTPMethod  string `json:"http_method"`
	HTTPURL     string `json:"http_url"`
	HTTPBody    string `json:"http_body"`
}

type UpdateJobRequest struct {
	CronExpr   *string    `json:"cron_expr"`
	HTTPMethod *string    `json:"http_method"`
	HTTPURL    *string    `json:"http_url"`
	HTTPBody   *string    `json:"http_body"`
	Status     *JobStatus `json:"status"`
}

// JobRun records each execution attempt.
type JobRun struct {
	ID         string    `json:"id"`
	JobID      string    `json:"job_id"`
	Status     RunStatus `json:"status"`
	Output     string    `json:"output"`
	StartedAt  time.Time `json:"started_at"`
	FinishedAt time.Time `json:"finished_at"`
}
