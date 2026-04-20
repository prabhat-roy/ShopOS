package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"github.com/shopos/scheduler-service/internal/domain"
)

// Storer is the persistence interface.
type Storer interface {
	Create(ctx context.Context, req *domain.CreateJobRequest, nextRun time.Time) (*domain.Job, error)
	Get(ctx context.Context, id string) (*domain.Job, error)
	List(ctx context.Context) ([]*domain.Job, error)
	DueJobs(ctx context.Context) ([]*domain.Job, error)
	Update(ctx context.Context, id string, req *domain.UpdateJobRequest, nextRun *time.Time) (*domain.Job, error)
	UpdateNextRun(ctx context.Context, id string, lastRun, nextRun time.Time) error
	Delete(ctx context.Context, id string) error
	SaveRun(ctx context.Context, run *domain.JobRun) error
	ListRuns(ctx context.Context, jobID string, limit int) ([]*domain.JobRun, error)
}

type SchedulerService struct {
	store  Storer
	parser cron.Parser
	client *http.Client
}

func New(store Storer) *SchedulerService {
	return &SchedulerService{
		store:  store,
		parser: cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow),
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (s *SchedulerService) CreateJob(ctx context.Context, req *domain.CreateJobRequest) (*domain.Job, error) {
	next, err := s.nextRun(req.CronExpr)
	if err != nil {
		return nil, domain.ErrInvalidCron
	}
	return s.store.Create(ctx, req, next)
}

func (s *SchedulerService) GetJob(ctx context.Context, id string) (*domain.Job, error) {
	return s.store.Get(ctx, id)
}

func (s *SchedulerService) ListJobs(ctx context.Context) ([]*domain.Job, error) {
	return s.store.List(ctx)
}

func (s *SchedulerService) UpdateJob(ctx context.Context, id string, req *domain.UpdateJobRequest) (*domain.Job, error) {
	var nextRun *time.Time
	if req.CronExpr != nil {
		t, err := s.nextRun(*req.CronExpr)
		if err != nil {
			return nil, domain.ErrInvalidCron
		}
		nextRun = &t
	}
	return s.store.Update(ctx, id, req, nextRun)
}

func (s *SchedulerService) DeleteJob(ctx context.Context, id string) error {
	return s.store.Delete(ctx, id)
}

func (s *SchedulerService) ListRuns(ctx context.Context, jobID string, limit int) ([]*domain.JobRun, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.store.ListRuns(ctx, jobID, limit)
}

// Tick checks for due jobs and executes them. Called by the scheduler loop.
func (s *SchedulerService) Tick(ctx context.Context) {
	jobs, err := s.store.DueJobs(ctx)
	if err != nil {
		slog.Error("DueJobs query failed", "err", err)
		return
	}
	for _, j := range jobs {
		go s.execute(ctx, j)
	}
}

func (s *SchedulerService) execute(ctx context.Context, j *domain.Job) {
	startedAt := time.Now().UTC()
	output, execErr := s.dispatch(ctx, j)
	finishedAt := time.Now().UTC()

	status := domain.RunSuccess
	if execErr != nil {
		status = domain.RunFailed
		output = execErr.Error()
	}

	run := &domain.JobRun{
		ID:         uuid.NewString(),
		JobID:      j.ID,
		Status:     status,
		Output:     output,
		StartedAt:  startedAt,
		FinishedAt: finishedAt,
	}
	_ = s.store.SaveRun(ctx, run)

	next, _ := s.nextRun(j.CronExpr)
	_ = s.store.UpdateNextRun(ctx, j.ID, startedAt, next)

	slog.Info("job executed", "job_id", j.ID, "status", status)
}

func (s *SchedulerService) dispatch(ctx context.Context, j *domain.Job) (string, error) {
	method := j.HTTPMethod
	if method == "" {
		method = http.MethodPost
	}
	var body io.Reader
	if j.HTTPBody != "" {
		body = bytes.NewBufferString(j.HTTPBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, j.HTTPURL, body)
	if err != nil {
		return "", err
	}
	if j.HTTPBody != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode >= 300 {
		return string(b), fmt.Errorf("non-2xx status %d", resp.StatusCode)
	}
	return string(b), nil
}

func (s *SchedulerService) nextRun(expr string) (time.Time, error) {
	sched, err := s.parser.Parse(expr)
	if err != nil {
		return time.Time{}, err
	}
	return sched.Next(time.Now().UTC()), nil
}
