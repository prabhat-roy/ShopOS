package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopos/scheduler-service/internal/domain"
)

type JobStore struct {
	db *sql.DB
}

func New(db *sql.DB) *JobStore {
	return &JobStore{db: db}
}

func (s *JobStore) Create(ctx context.Context, req *domain.CreateJobRequest, nextRun time.Time) (*domain.Job, error) {
	j := &domain.Job{
		ID:          uuid.NewString(),
		Name:        req.Name,
		Description: req.Description,
		CronExpr:    req.CronExpr,
		HTTPMethod:  req.HTTPMethod,
		HTTPURL:     req.HTTPURL,
		HTTPBody:    req.HTTPBody,
		Status:      domain.StatusEnabled,
		NextRunAt:   nextRun,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO jobs (id, name, description, cron_expr, http_method, http_url, http_body, status, next_run_at, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		j.ID, j.Name, j.Description, j.CronExpr, j.HTTPMethod, j.HTTPURL, j.HTTPBody,
		j.Status, j.NextRunAt, j.CreatedAt, j.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("store.Create: %w", err)
	}
	return j, nil
}

func (s *JobStore) Get(ctx context.Context, id string) (*domain.Job, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, name, description, cron_expr, http_method, http_url, http_body, status, next_run_at, last_run_at, created_at, updated_at
		 FROM jobs WHERE id = $1`, id)
	return scanJob(row)
}

func (s *JobStore) List(ctx context.Context) ([]*domain.Job, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name, description, cron_expr, http_method, http_url, http_body, status, next_run_at, last_run_at, created_at, updated_at
		 FROM jobs ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*domain.Job
	for rows.Next() {
		j, err := scanJobRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, j)
	}
	return out, rows.Err()
}

// DueJobs returns all enabled jobs whose next_run_at <= now.
func (s *JobStore) DueJobs(ctx context.Context) ([]*domain.Job, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name, description, cron_expr, http_method, http_url, http_body, status, next_run_at, last_run_at, created_at, updated_at
		 FROM jobs WHERE status = 'enabled' AND next_run_at <= $1`, time.Now().UTC())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*domain.Job
	for rows.Next() {
		j, err := scanJobRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, j)
	}
	return out, rows.Err()
}

func (s *JobStore) Update(ctx context.Context, id string, req *domain.UpdateJobRequest, nextRun *time.Time) (*domain.Job, error) {
	j, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.CronExpr != nil {
		j.CronExpr = *req.CronExpr
	}
	if req.HTTPMethod != nil {
		j.HTTPMethod = *req.HTTPMethod
	}
	if req.HTTPURL != nil {
		j.HTTPURL = *req.HTTPURL
	}
	if req.HTTPBody != nil {
		j.HTTPBody = *req.HTTPBody
	}
	if req.Status != nil {
		j.Status = *req.Status
	}
	if nextRun != nil {
		j.NextRunAt = *nextRun
	}
	j.UpdatedAt = time.Now().UTC()

	_, err = s.db.ExecContext(ctx,
		`UPDATE jobs SET cron_expr=$1, http_method=$2, http_url=$3, http_body=$4, status=$5, next_run_at=$6, updated_at=$7 WHERE id=$8`,
		j.CronExpr, j.HTTPMethod, j.HTTPURL, j.HTTPBody, j.Status, j.NextRunAt, j.UpdatedAt, id,
	)
	if err != nil {
		return nil, fmt.Errorf("store.Update: %w", err)
	}
	return j, nil
}

func (s *JobStore) UpdateNextRun(ctx context.Context, id string, lastRun, nextRun time.Time) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE jobs SET last_run_at=$1, next_run_at=$2, updated_at=$3 WHERE id=$4`,
		lastRun, nextRun, time.Now().UTC(), id,
	)
	return err
}

func (s *JobStore) Delete(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM jobs WHERE id = $1`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (s *JobStore) SaveRun(ctx context.Context, run *domain.JobRun) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO job_runs (id, job_id, status, output, started_at, finished_at)
		 VALUES ($1,$2,$3,$4,$5,$6)`,
		run.ID, run.JobID, run.Status, run.Output, run.StartedAt, run.FinishedAt,
	)
	return err
}

func (s *JobStore) ListRuns(ctx context.Context, jobID string, limit int) ([]*domain.JobRun, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, job_id, status, output, started_at, finished_at
		 FROM job_runs WHERE job_id=$1 ORDER BY started_at DESC LIMIT $2`, jobID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*domain.JobRun
	for rows.Next() {
		var r domain.JobRun
		if err := rows.Scan(&r.ID, &r.JobID, &r.Status, &r.Output, &r.StartedAt, &r.FinishedAt); err != nil {
			return nil, err
		}
		out = append(out, &r)
	}
	return out, rows.Err()
}

func scanJob(row *sql.Row) (*domain.Job, error) {
	var j domain.Job
	err := row.Scan(&j.ID, &j.Name, &j.Description, &j.CronExpr,
		&j.HTTPMethod, &j.HTTPURL, &j.HTTPBody, &j.Status,
		&j.NextRunAt, &j.LastRunAt, &j.CreatedAt, &j.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return &j, err
}

func scanJobRow(rows *sql.Rows) (*domain.Job, error) {
	var j domain.Job
	err := rows.Scan(&j.ID, &j.Name, &j.Description, &j.CronExpr,
		&j.HTTPMethod, &j.HTTPURL, &j.HTTPBody, &j.Status,
		&j.NextRunAt, &j.LastRunAt, &j.CreatedAt, &j.UpdatedAt)
	return &j, err
}
