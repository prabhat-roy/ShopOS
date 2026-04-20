package worker

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/shopos/worker-job-queue/internal/domain"
)

// Queuer is the subset of queue.Queuer that the Worker depends on.
// Defining it here keeps the worker package free of a direct import of the
// queue package and makes it trivially mockable in tests.
type Queuer interface {
	Dequeue(ctx context.Context, queue string) (*domain.Job, error)
	Complete(ctx context.Context, id string) error
	Fail(ctx context.Context, id string, err error) error
}

// Worker polls one or more named queues and dispatches each job to its
// CallbackURL via an HTTP POST.
type Worker struct {
	queue       Queuer
	client      *http.Client
	concurrency int
	queues      []string
}

// New constructs a Worker.
//
//   - q            — queue backend (must satisfy Queuer)
//   - client       — HTTP client used for callback requests; caller controls timeout
//   - concurrency  — number of goroutines spawned by Run
//   - queues       — ordered list of queue names to poll (round-robin across goroutines)
func New(q Queuer, client *http.Client, concurrency int, queues []string) *Worker {
	if concurrency <= 0 {
		concurrency = 1
	}
	if len(queues) == 0 {
		queues = []string{"default"}
	}
	return &Worker{
		queue:       q,
		client:      client,
		concurrency: concurrency,
		queues:      queues,
	}
}

// Run launches concurrency goroutines, each running processLoop, and blocks
// until ctx is cancelled. All goroutines are guaranteed to have exited before
// Run returns.
func (w *Worker) Run(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Add(w.concurrency)

	for i := 0; i < w.concurrency; i++ {
		go func(workerID int) {
			defer wg.Done()
			w.processLoop(ctx, workerID)
		}(i)
	}

	wg.Wait()
	slog.Info("all worker goroutines stopped")
}

// processLoop continuously dequeues jobs from every configured queue in
// round-robin order until ctx is cancelled.
func (w *Worker) processLoop(ctx context.Context, workerID int) {
	logger := slog.With("worker_id", workerID)
	logger.Info("worker started", "queues", w.queues)

	for {
		select {
		case <-ctx.Done():
			logger.Info("worker shutting down")
			return
		default:
		}

		processed := false
		for _, q := range w.queues {
			select {
			case <-ctx.Done():
				return
			default:
			}

			job, err := w.queue.Dequeue(ctx, q)
			if err != nil {
				if errors.Is(err, domain.ErrQueueEmpty) {
					// Nothing on this queue; try the next one.
					continue
				}
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					return
				}
				logger.Error("dequeue error", "queue", q, "error", err)
				continue
			}

			processed = true
			logger.Info("processing job", "job_id", job.ID, "queue", q, "attempt", job.Attempt)

			if err := w.dispatch(ctx, job); err != nil {
				logger.Warn("job failed", "job_id", job.ID, "error", err)
				if failErr := w.queue.Fail(ctx, job.ID, err); failErr != nil {
					logger.Error("failed to record job failure", "job_id", job.ID, "error", failErr)
				}
				continue
			}

			if err := w.queue.Complete(ctx, job.ID); err != nil {
				logger.Error("failed to mark job complete", "job_id", job.ID, "error", err)
			} else {
				logger.Info("job completed", "job_id", job.ID)
			}
		}

		// If no queue yielded a job, back off briefly to avoid a hot spin.
		if !processed {
			select {
			case <-ctx.Done():
				return
			case <-time.After(500 * time.Millisecond):
			}
		}
	}
}

// dispatch POSTs the job's Payload to its CallbackURL. A non-2xx HTTP status
// is treated as a failure.
func (w *Worker) dispatch(ctx context.Context, job *domain.Job) error {
	if job.CallbackURL == "" {
		return errors.New("job has no callback URL")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, job.CallbackURL, bytes.NewReader(job.Payload))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Job-ID", job.ID)
	req.Header.Set("X-Job-Queue", job.Queue)
	req.Header.Set("X-Job-Attempt", fmt.Sprintf("%d", job.Attempt))

	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("http post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("callback returned non-2xx status: %d", resp.StatusCode)
	}

	return nil
}
