package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/shopos/worker-job-queue/internal/domain"
)

// Queuer is the interface every consumer of the queue layer depends on.
// All concrete and mock implementations must satisfy this interface.
type Queuer interface {
	Enqueue(ctx context.Context, req domain.EnqueueRequest) (*domain.Job, error)
	Dequeue(ctx context.Context, queue string) (*domain.Job, error)
	Complete(ctx context.Context, id string) error
	Fail(ctx context.Context, id string, jobErr error) error
	GetJob(ctx context.Context, id string) (*domain.Job, error)
	ListDead(ctx context.Context, queue string) ([]*domain.Job, error)
}

// compile-time assertion
var _ Queuer = (*RedisQueue)(nil)

// RedisQueue implements Queuer using Redis lists for priority queues and hashes
// for job metadata storage.
type RedisQueue struct {
	client *redis.Client
}

// NewRedisQueue constructs a RedisQueue backed by the supplied redis.Client.
func NewRedisQueue(client *redis.Client) *RedisQueue {
	return &RedisQueue{client: client}
}

// ----- key helpers ----------------------------------------------------------

func listKey(queue string, p domain.Priority) string {
	return fmt.Sprintf("queue:%s:%s", queue, p)
}

func dlqKey(queue string) string {
	return fmt.Sprintf("dlq:%s", queue)
}

func jobKey(id string) string {
	return fmt.Sprintf("job:%s", id)
}

// ----- Enqueue --------------------------------------------------------------

// Enqueue creates a new job from req, persists it as a Redis hash, and pushes
// its ID onto the appropriate priority list.
func (q *RedisQueue) Enqueue(ctx context.Context, req domain.EnqueueRequest) (*domain.Job, error) {
	if req.Priority == "" {
		req.Priority = domain.PriorityNormal
	}
	if !domain.IsValidPriority(req.Priority) {
		return nil, fmt.Errorf("invalid priority %q", req.Priority)
	}
	if req.MaxRetries <= 0 {
		req.MaxRetries = 3
	}

	job := &domain.Job{
		ID:          uuid.New().String(),
		Queue:       req.Queue,
		Priority:    req.Priority,
		Payload:     req.Payload,
		CallbackURL: req.CallbackURL,
		MaxRetries:  req.MaxRetries,
		Attempt:     0,
		Status:      domain.StatusPending,
		EnqueuedAt:  time.Now().UTC(),
	}

	if err := q.saveJob(ctx, job); err != nil {
		return nil, fmt.Errorf("save job: %w", err)
	}

	if err := q.client.LPush(ctx, listKey(req.Queue, req.Priority), job.ID).Err(); err != nil {
		return nil, fmt.Errorf("lpush: %w", err)
	}

	return job, nil
}

// ----- Dequeue --------------------------------------------------------------

// Dequeue performs a blocking pop across priority lists in order
// high → normal → low. It updates the job status to running and returns the
// job. ErrQueueEmpty is returned when the context deadline is exceeded before
// any item is available.
func (q *RedisQueue) Dequeue(ctx context.Context, queue string) (*domain.Job, error) {
	keys := []string{
		listKey(queue, domain.PriorityHigh),
		listKey(queue, domain.PriorityNormal),
		listKey(queue, domain.PriorityLow),
	}

	// Use a short timeout so the loop in processLoop can re-check context
	// cancellation regularly. BRPOP returns redis.Nil when timeout expires.
	result, err := q.client.BRPop(ctx, 2*time.Second, keys...).Result()
	if err == redis.Nil {
		return nil, domain.ErrQueueEmpty
	}
	if err != nil {
		return nil, fmt.Errorf("brpop: %w", err)
	}

	// BRPop returns [key, value]
	if len(result) < 2 {
		return nil, domain.ErrQueueEmpty
	}
	id := result[1]

	job, err := q.GetJob(ctx, id)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	job.Status = domain.StatusRunning
	job.ProcessedAt = &now
	if err := q.saveJob(ctx, job); err != nil {
		return nil, fmt.Errorf("update job running: %w", err)
	}

	return job, nil
}

// ----- Complete -------------------------------------------------------------

// Complete marks the job identified by id as completed.
func (q *RedisQueue) Complete(ctx context.Context, id string) error {
	job, err := q.GetJob(ctx, id)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	job.Status = domain.StatusCompleted
	job.ProcessedAt = &now

	return q.saveJob(ctx, job)
}

// ----- Fail -----------------------------------------------------------------

// Fail records a failure on the job. If the job has remaining retry attempts
// it is re-queued at its original priority. Once attempts are exhausted it is
// moved to the dead-letter queue with status StatusDead.
func (q *RedisQueue) Fail(ctx context.Context, id string, jobErr error) error {
	job, err := q.GetJob(ctx, id)
	if err != nil {
		return err
	}

	job.Attempt++
	job.Error = jobErr.Error()

	if job.Attempt < job.MaxRetries {
		job.Status = domain.StatusFailed
		if err := q.saveJob(ctx, job); err != nil {
			return fmt.Errorf("save failed job: %w", err)
		}
		// Re-queue with exponential backoff by pushing to the back (LPUSH).
		if err := q.client.LPush(ctx, listKey(job.Queue, job.Priority), job.ID).Err(); err != nil {
			return fmt.Errorf("requeue failed job: %w", err)
		}
		return nil
	}

	// Move to dead-letter queue.
	job.Status = domain.StatusDead
	if err := q.saveJob(ctx, job); err != nil {
		return fmt.Errorf("save dead job: %w", err)
	}
	if err := q.client.LPush(ctx, dlqKey(job.Queue), job.ID).Err(); err != nil {
		return fmt.Errorf("push to dlq: %w", err)
	}

	return nil
}

// ----- GetJob ---------------------------------------------------------------

// GetJob retrieves the job metadata stored in the Redis hash job:{id}.
func (q *RedisQueue) GetJob(ctx context.Context, id string) (*domain.Job, error) {
	data, err := q.client.HGetAll(ctx, jobKey(id)).Result()
	if err != nil {
		return nil, fmt.Errorf("hgetall: %w", err)
	}
	if len(data) == 0 {
		return nil, domain.ErrNotFound
	}
	return unmarshalJob(data)
}

// ----- ListDead -------------------------------------------------------------

// ListDead returns up to 100 jobs from the dead-letter queue for the given
// queue name.
func (q *RedisQueue) ListDead(ctx context.Context, queue string) ([]*domain.Job, error) {
	ids, err := q.client.LRange(ctx, dlqKey(queue), 0, 99).Result()
	if err != nil {
		return nil, fmt.Errorf("lrange dlq: %w", err)
	}

	jobs := make([]*domain.Job, 0, len(ids))
	for _, id := range ids {
		job, err := q.GetJob(ctx, id)
		if err != nil {
			// Skip missing entries caused by external cleanup.
			continue
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}

// ----- internal helpers -----------------------------------------------------

// saveJob serialises a Job into a flat string map and stores it as a Redis hash.
func (q *RedisQueue) saveJob(ctx context.Context, job *domain.Job) error {
	m, err := marshalJob(job)
	if err != nil {
		return err
	}

	args := make([]interface{}, 0, len(m)*2)
	for k, v := range m {
		args = append(args, k, v)
	}

	return q.client.HSet(ctx, jobKey(job.ID), args...).Err()
}

// marshalJob converts a Job to a string map suitable for HSet.
func marshalJob(job *domain.Job) (map[string]string, error) {
	payloadJSON, err := json.Marshal(job.Payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	processedAt := ""
	if job.ProcessedAt != nil {
		processedAt = job.ProcessedAt.Format(time.RFC3339Nano)
	}

	return map[string]string{
		"id":           job.ID,
		"queue":        job.Queue,
		"priority":     string(job.Priority),
		"payload":      string(payloadJSON),
		"callback_url": job.CallbackURL,
		"max_retries":  fmt.Sprintf("%d", job.MaxRetries),
		"attempt":      fmt.Sprintf("%d", job.Attempt),
		"status":       string(job.Status),
		"error":        job.Error,
		"enqueued_at":  job.EnqueuedAt.Format(time.RFC3339Nano),
		"processed_at": processedAt,
	}, nil
}

// unmarshalJob rebuilds a Job from the flat string map returned by HGetAll.
func unmarshalJob(data map[string]string) (*domain.Job, error) {
	maxRetries, err := parseInt(data["max_retries"])
	if err != nil {
		return nil, fmt.Errorf("parse max_retries: %w", err)
	}

	attempt, err := parseInt(data["attempt"])
	if err != nil {
		return nil, fmt.Errorf("parse attempt: %w", err)
	}

	enqueuedAt, err := time.Parse(time.RFC3339Nano, data["enqueued_at"])
	if err != nil {
		return nil, fmt.Errorf("parse enqueued_at: %w", err)
	}

	var processedAt *time.Time
	if data["processed_at"] != "" {
		t, err := time.Parse(time.RFC3339Nano, data["processed_at"])
		if err != nil {
			return nil, fmt.Errorf("parse processed_at: %w", err)
		}
		processedAt = &t
	}

	// payload was stored as a JSON-encoded byte slice string; decode it back.
	var payload []byte
	if err := json.Unmarshal([]byte(data["payload"]), &payload); err != nil {
		// Fallback: treat it as raw bytes.
		payload = []byte(data["payload"])
	}

	return &domain.Job{
		ID:          data["id"],
		Queue:       data["queue"],
		Priority:    domain.Priority(data["priority"]),
		Payload:     payload,
		CallbackURL: data["callback_url"],
		MaxRetries:  maxRetries,
		Attempt:     attempt,
		Status:      domain.JobStatus(data["status"]),
		Error:       data["error"],
		EnqueuedAt:  enqueuedAt,
		ProcessedAt: processedAt,
	}, nil
}

func parseInt(s string) (int, error) {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}
