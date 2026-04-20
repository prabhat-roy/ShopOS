package queue_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shopos/worker-job-queue/internal/domain"
	"github.com/shopos/worker-job-queue/internal/queue"
)

// ---------------------------------------------------------------------------
// mockQueuer — satisfies queue.Queuer without a real Redis connection.
// ---------------------------------------------------------------------------

type mockQueuer struct {
	jobs   map[string]*domain.Job
	lists  map[string][]string // queue:priority → []id
	dlqs   map[string][]string // dlq:queue      → []id
	nextID int
}

func newMockQueuer() *mockQueuer {
	return &mockQueuer{
		jobs:  make(map[string]*domain.Job),
		lists: make(map[string][]string),
		dlqs:  make(map[string][]string),
	}
}

func (m *mockQueuer) Enqueue(ctx context.Context, req domain.EnqueueRequest) (*domain.Job, error) {
	if req.Queue == "" {
		return nil, errors.New("queue name required")
	}
	if req.Priority == "" {
		req.Priority = domain.PriorityNormal
	}
	if !domain.IsValidPriority(req.Priority) {
		return nil, errors.New("invalid priority")
	}
	if req.MaxRetries <= 0 {
		req.MaxRetries = 3
	}

	m.nextID++
	id := string(rune('A' + m.nextID - 1)) // simple deterministic IDs: A, B, C …

	job := &domain.Job{
		ID:          id,
		Queue:       req.Queue,
		Priority:    req.Priority,
		Payload:     req.Payload,
		CallbackURL: req.CallbackURL,
		MaxRetries:  req.MaxRetries,
		Attempt:     0,
		Status:      domain.StatusPending,
		EnqueuedAt:  time.Now().UTC(),
	}
	m.jobs[id] = job

	key := string(req.Priority)
	m.lists[key] = append(m.lists[key], id)

	return job, nil
}

func (m *mockQueuer) Dequeue(ctx context.Context, queueName string) (*domain.Job, error) {
	for _, p := range []domain.Priority{domain.PriorityHigh, domain.PriorityNormal, domain.PriorityLow} {
		key := string(p)
		if len(m.lists[key]) == 0 {
			continue
		}
		// pop from the end (BRPOP behaviour)
		n := len(m.lists[key])
		id := m.lists[key][n-1]
		m.lists[key] = m.lists[key][:n-1]

		job, ok := m.jobs[id]
		if !ok {
			return nil, domain.ErrNotFound
		}
		now := time.Now().UTC()
		job.Status = domain.StatusRunning
		job.ProcessedAt = &now
		return job, nil
	}
	return nil, domain.ErrQueueEmpty
}

func (m *mockQueuer) Complete(ctx context.Context, id string) error {
	job, ok := m.jobs[id]
	if !ok {
		return domain.ErrNotFound
	}
	job.Status = domain.StatusCompleted
	return nil
}

func (m *mockQueuer) Fail(ctx context.Context, id string, jobErr error) error {
	job, ok := m.jobs[id]
	if !ok {
		return domain.ErrNotFound
	}
	job.Attempt++
	job.Error = jobErr.Error()

	if job.Attempt < job.MaxRetries {
		job.Status = domain.StatusFailed
		key := string(job.Priority)
		m.lists[key] = append(m.lists[key], id)
		return nil
	}

	job.Status = domain.StatusDead
	m.dlqs[job.Queue] = append(m.dlqs[job.Queue], id)
	return nil
}

func (m *mockQueuer) GetJob(ctx context.Context, id string) (*domain.Job, error) {
	job, ok := m.jobs[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return job, nil
}

func (m *mockQueuer) ListDead(ctx context.Context, queueName string) ([]*domain.Job, error) {
	ids := m.dlqs[queueName]
	out := make([]*domain.Job, 0, len(ids))
	for _, id := range ids {
		if j, ok := m.jobs[id]; ok {
			out = append(out, j)
		}
	}
	return out, nil
}

// compile-time assertion: mockQueuer must satisfy queue.Queuer
var _ queue.Queuer = (*mockQueuer)(nil)

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestEnqueue_DefaultsApplied(t *testing.T) {
	ctx := context.Background()
	q := newMockQueuer()

	req := domain.EnqueueRequest{
		Queue:       "emails",
		Payload:     []byte(`{"to":"user@example.com"}`),
		CallbackURL: "http://localhost/cb",
		// Priority and MaxRetries intentionally omitted → defaults
	}

	job, err := q.Enqueue(ctx, req)
	if err != nil {
		t.Fatalf("Enqueue returned unexpected error: %v", err)
	}
	if job.ID == "" {
		t.Error("expected non-empty job ID")
	}
	if job.Priority != domain.PriorityNormal {
		t.Errorf("expected default priority %q, got %q", domain.PriorityNormal, job.Priority)
	}
	if job.MaxRetries != 3 {
		t.Errorf("expected default MaxRetries=3, got %d", job.MaxRetries)
	}
	if job.Status != domain.StatusPending {
		t.Errorf("expected status %q, got %q", domain.StatusPending, job.Status)
	}
	if job.Queue != "emails" {
		t.Errorf("expected queue %q, got %q", "emails", job.Queue)
	}
}

func TestEnqueue_InvalidPriority(t *testing.T) {
	ctx := context.Background()
	q := newMockQueuer()

	_, err := q.Enqueue(ctx, domain.EnqueueRequest{
		Queue:    "emails",
		Priority: domain.Priority("urgent"), // not a valid constant
	})
	if err == nil {
		t.Fatal("expected error for invalid priority, got nil")
	}
}

func TestEnqueue_EmptyQueue(t *testing.T) {
	ctx := context.Background()
	q := newMockQueuer()

	_, err := q.Enqueue(ctx, domain.EnqueueRequest{Queue: ""})
	if err == nil {
		t.Fatal("expected error for empty queue name, got nil")
	}
}

func TestDequeue_ReturnsEnqueuedJob(t *testing.T) {
	ctx := context.Background()
	q := newMockQueuer()

	req := domain.EnqueueRequest{
		Queue:       "default",
		Priority:    domain.PriorityNormal,
		Payload:     []byte(`{}`),
		CallbackURL: "http://localhost/cb",
		MaxRetries:  3,
	}
	enqueued, _ := q.Enqueue(ctx, req)

	job, err := q.Dequeue(ctx, "default")
	if err != nil {
		t.Fatalf("Dequeue returned unexpected error: %v", err)
	}
	if job.ID != enqueued.ID {
		t.Errorf("expected job ID %q, got %q", enqueued.ID, job.ID)
	}
	if job.Status != domain.StatusRunning {
		t.Errorf("expected status %q after dequeue, got %q", domain.StatusRunning, job.Status)
	}
	if job.ProcessedAt == nil {
		t.Error("expected ProcessedAt to be set after dequeue")
	}
}

func TestDequeue_EmptyQueue(t *testing.T) {
	ctx := context.Background()
	q := newMockQueuer()

	_, err := q.Dequeue(ctx, "nonexistent")
	if !errors.Is(err, domain.ErrQueueEmpty) {
		t.Errorf("expected ErrQueueEmpty, got %v", err)
	}
}

func TestDequeue_PriorityOrder(t *testing.T) {
	ctx := context.Background()
	q := newMockQueuer()

	// Enqueue low then high; high must be dequeued first.
	lowJob, _ := q.Enqueue(ctx, domain.EnqueueRequest{
		Queue:      "work",
		Priority:   domain.PriorityLow,
		Payload:    []byte(`{"order":"low"}`),
		MaxRetries: 1,
	})
	highJob, _ := q.Enqueue(ctx, domain.EnqueueRequest{
		Queue:      "work",
		Priority:   domain.PriorityHigh,
		Payload:    []byte(`{"order":"high"}`),
		MaxRetries: 1,
	})

	first, err := q.Dequeue(ctx, "work")
	if err != nil {
		t.Fatalf("first Dequeue error: %v", err)
	}
	if first.ID != highJob.ID {
		t.Errorf("expected high-priority job %q first, got %q", highJob.ID, first.ID)
	}

	second, err := q.Dequeue(ctx, "work")
	if err != nil {
		t.Fatalf("second Dequeue error: %v", err)
	}
	if second.ID != lowJob.ID {
		t.Errorf("expected low-priority job %q second, got %q", lowJob.ID, second.ID)
	}
}

func TestComplete_SetsStatusCompleted(t *testing.T) {
	ctx := context.Background()
	q := newMockQueuer()

	job, _ := q.Enqueue(ctx, domain.EnqueueRequest{
		Queue:      "default",
		Priority:   domain.PriorityNormal,
		MaxRetries: 3,
	})
	// Simulate dequeue first so status becomes running
	_, _ = q.Dequeue(ctx, "default")

	if err := q.Complete(ctx, job.ID); err != nil {
		t.Fatalf("Complete returned unexpected error: %v", err)
	}

	stored, _ := q.GetJob(ctx, job.ID)
	if stored.Status != domain.StatusCompleted {
		t.Errorf("expected status %q, got %q", domain.StatusCompleted, stored.Status)
	}
}

func TestComplete_NotFound(t *testing.T) {
	ctx := context.Background()
	q := newMockQueuer()

	err := q.Complete(ctx, "nonexistent-id")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestFail_IncrementsAttemptAndRequeues(t *testing.T) {
	ctx := context.Background()
	q := newMockQueuer()

	job, _ := q.Enqueue(ctx, domain.EnqueueRequest{
		Queue:      "tasks",
		Priority:   domain.PriorityNormal,
		MaxRetries: 3,
	})
	_, _ = q.Dequeue(ctx, "tasks")

	jobErr := errors.New("connection timeout")
	if err := q.Fail(ctx, job.ID, jobErr); err != nil {
		t.Fatalf("Fail returned unexpected error: %v", err)
	}

	stored, _ := q.GetJob(ctx, job.ID)
	if stored.Attempt != 1 {
		t.Errorf("expected attempt=1 after first failure, got %d", stored.Attempt)
	}
	if stored.Status != domain.StatusFailed {
		t.Errorf("expected status %q, got %q", domain.StatusFailed, stored.Status)
	}
	if stored.Error != "connection timeout" {
		t.Errorf("expected error message %q, got %q", "connection timeout", stored.Error)
	}

	// Job should be back on the queue for retry
	requeued, err := q.Dequeue(ctx, "tasks")
	if err != nil {
		t.Fatalf("expected requeued job to be dequeueable, got: %v", err)
	}
	if requeued.ID != job.ID {
		t.Errorf("expected requeued job ID %q, got %q", job.ID, requeued.ID)
	}
}

func TestFail_ExceedsMaxRetries_MovesToDead(t *testing.T) {
	ctx := context.Background()
	q := newMockQueuer()

	job, _ := q.Enqueue(ctx, domain.EnqueueRequest{
		Queue:      "tasks",
		Priority:   domain.PriorityNormal,
		MaxRetries: 2,
	})
	_, _ = q.Dequeue(ctx, "tasks")

	jobErr := errors.New("permanent failure")

	// First failure — still under max
	if err := q.Fail(ctx, job.ID, jobErr); err != nil {
		t.Fatalf("first Fail error: %v", err)
	}
	// Re-dequeue for second attempt
	_, _ = q.Dequeue(ctx, "tasks")

	// Second failure — hits max, goes to DLQ
	if err := q.Fail(ctx, job.ID, jobErr); err != nil {
		t.Fatalf("second Fail error: %v", err)
	}

	stored, _ := q.GetJob(ctx, job.ID)
	if stored.Status != domain.StatusDead {
		t.Errorf("expected status %q after exhausting retries, got %q", domain.StatusDead, stored.Status)
	}
}

func TestFail_NotFound(t *testing.T) {
	ctx := context.Background()
	q := newMockQueuer()

	err := q.Fail(ctx, "ghost-id", errors.New("boom"))
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestListDead_ReturnsDeadJobs(t *testing.T) {
	ctx := context.Background()
	q := newMockQueuer()

	// Exhaust retries for two jobs so they end up in the DLQ
	for i := 0; i < 2; i++ {
		job, _ := q.Enqueue(ctx, domain.EnqueueRequest{
			Queue:      "alerts",
			Priority:   domain.PriorityHigh,
			MaxRetries: 1,
		})
		_, _ = q.Dequeue(ctx, "alerts")
		_ = q.Fail(ctx, job.ID, errors.New("dead"))
	}

	dead, err := q.ListDead(ctx, "alerts")
	if err != nil {
		t.Fatalf("ListDead error: %v", err)
	}
	if len(dead) != 2 {
		t.Errorf("expected 2 dead jobs, got %d", len(dead))
	}
	for _, j := range dead {
		if j.Status != domain.StatusDead {
			t.Errorf("expected dead status, got %q for job %q", j.Status, j.ID)
		}
	}
}

func TestListDead_EmptyQueue(t *testing.T) {
	ctx := context.Background()
	q := newMockQueuer()

	dead, err := q.ListDead(ctx, "empty-queue")
	if err != nil {
		t.Fatalf("ListDead error: %v", err)
	}
	if len(dead) != 0 {
		t.Errorf("expected 0 dead jobs for empty queue, got %d", len(dead))
	}
}

func TestGetJob_NotFound(t *testing.T) {
	ctx := context.Background()
	q := newMockQueuer()

	_, err := q.GetJob(ctx, "missing")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
