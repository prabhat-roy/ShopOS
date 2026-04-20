package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/shopos/worker-job-queue/internal/domain"
)

// Queuer is the subset of queue.Queuer that JobService depends on.
// Using a locally-defined interface keeps the service package decoupled from
// the concrete queue implementation.
type Queuer interface {
	Enqueue(ctx context.Context, req domain.EnqueueRequest) (*domain.Job, error)
	GetJob(ctx context.Context, id string) (*domain.Job, error)
	ListDead(ctx context.Context, queue string) ([]*domain.Job, error)
	Fail(ctx context.Context, id string, err error) error
}

// JobService implements the business logic layer between the HTTP handler and
// the queue backend.
type JobService struct {
	queue Queuer
}

// New constructs a JobService backed by the supplied Queuer.
func New(q Queuer) *JobService {
	return &JobService{queue: q}
}

// Enqueue validates req, then delegates to the queue backend.
//
// Validation rules:
//   - Queue name must be non-empty and contain only safe characters.
//   - Priority (when provided) must be a recognised value.
//   - CallbackURL must be non-empty.
//   - MaxRetries defaults to 3 when zero or negative.
func (s *JobService) Enqueue(ctx context.Context, req *domain.EnqueueRequest) (*domain.Job, error) {
	if err := validateEnqueueRequest(req); err != nil {
		return nil, err
	}
	return s.queue.Enqueue(ctx, *req)
}

// GetJob fetches a single job by ID. Returns domain.ErrNotFound when the ID
// does not exist.
func (s *JobService) GetJob(ctx context.Context, id string) (*domain.Job, error) {
	if id == "" {
		return nil, errors.New("job id must not be empty")
	}
	return s.queue.GetJob(ctx, id)
}

// ListDead returns all jobs sitting in the dead-letter queue for queue.
func (s *JobService) ListDead(ctx context.Context, queue string) ([]*domain.Job, error) {
	if queue == "" {
		return nil, errors.New("queue name must not be empty")
	}
	return s.queue.ListDead(ctx, queue)
}

// Retry re-enqueues a dead job so it will be picked up again by a worker.
// The job must currently have StatusDead; its attempt counter and error string
// are reset to give it a clean slate.
func (s *JobService) Retry(ctx context.Context, queue, id string) (*domain.Job, error) {
	if queue == "" {
		return nil, errors.New("queue name must not be empty")
	}
	if id == "" {
		return nil, errors.New("job id must not be empty")
	}

	job, err := s.queue.GetJob(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get job: %w", err)
	}

	if job.Status != domain.StatusDead {
		return nil, fmt.Errorf("cannot retry job with status %q: only dead jobs may be retried", job.Status)
	}

	if job.Queue != queue {
		return nil, fmt.Errorf("job %q belongs to queue %q, not %q", id, job.Queue, queue)
	}

	// Re-enqueue with the same configuration but fresh counters.
	req := domain.EnqueueRequest{
		Queue:       job.Queue,
		Priority:    job.Priority,
		Payload:     job.Payload,
		CallbackURL: job.CallbackURL,
		MaxRetries:  job.MaxRetries,
	}

	newJob, err := s.queue.Enqueue(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("re-enqueue: %w", err)
	}

	return newJob, nil
}

// ---------------------------------------------------------------------------
// validation helpers
// ---------------------------------------------------------------------------

func validateEnqueueRequest(req *domain.EnqueueRequest) error {
	if req.Queue == "" {
		return errors.New("queue name must not be empty")
	}
	if !isValidQueueName(req.Queue) {
		return fmt.Errorf("queue name %q contains invalid characters (use letters, digits, hyphens, underscores)", req.Queue)
	}
	if req.CallbackURL == "" {
		return errors.New("callback_url must not be empty")
	}
	if req.Priority != "" && !domain.IsValidPriority(req.Priority) {
		return fmt.Errorf("invalid priority %q: must be one of high, normal, low", req.Priority)
	}
	if req.MaxRetries < 0 {
		return errors.New("max_retries must be >= 0")
	}
	return nil
}

// isValidQueueName returns true when name consists only of letters, digits,
// hyphens, and underscores (no dots, slashes, spaces, etc.).
func isValidQueueName(name string) bool {
	for _, r := range name {
		if !isAlphaNum(r) && r != '-' && r != '_' {
			return false
		}
	}
	return len(name) > 0
}

func isAlphaNum(r rune) bool {
	return (r >= 'a' && r <= 'z') ||
		(r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9')
}
