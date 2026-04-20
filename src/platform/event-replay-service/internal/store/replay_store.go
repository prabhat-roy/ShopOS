package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/shopos/event-replay-service/internal/domain"
)

// ReplayStore persists and retrieves ReplayJob records using PostgreSQL.
type ReplayStore struct {
	db *sql.DB
}

// New creates a new ReplayStore backed by the supplied *sql.DB.
func New(db *sql.DB) *ReplayStore {
	return &ReplayStore{db: db}
}

// Create inserts a new replay job row. The job must already have its ID set.
func (s *ReplayStore) Create(ctx context.Context, job *domain.ReplayJob) error {
	const q = `
		INSERT INTO replay_jobs
			(id, stream_id, stream_type, event_type,
			 from_seq, to_seq, from_time, to_time,
			 target, target_topic, status, events_replayed,
			 error_message, started_at, completed_at, created_at)
		VALUES
			($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`

	_, err := s.db.ExecContext(ctx, q,
		job.ID, job.StreamID, job.StreamType, job.EventType,
		job.FromSeq, job.ToSeq, job.FromTime, job.ToTime,
		job.Target, job.TargetTopic, job.Status, job.EventsReplayed,
		job.ErrorMessage, job.StartedAt, job.CompletedAt, job.CreatedAt,
	)
	return err
}

// Get retrieves a single replay job by its ID.
func (s *ReplayStore) Get(ctx context.Context, id string) (*domain.ReplayJob, error) {
	const q = `
		SELECT id, stream_id, stream_type, event_type,
		       from_seq, to_seq, from_time, to_time,
		       target, target_topic, status, events_replayed,
		       error_message, started_at, completed_at, created_at
		FROM replay_jobs
		WHERE id = $1`

	row := s.db.QueryRowContext(ctx, q, id)
	job, err := scanJob(row)
	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	return job, err
}

// List returns all replay jobs ordered by creation time (newest first).
func (s *ReplayStore) List(ctx context.Context) ([]*domain.ReplayJob, error) {
	const q = `
		SELECT id, stream_id, stream_type, event_type,
		       from_seq, to_seq, from_time, to_time,
		       target, target_topic, status, events_replayed,
		       error_message, started_at, completed_at, created_at
		FROM replay_jobs
		ORDER BY created_at DESC`

	rows, err := s.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []*domain.ReplayJob
	for rows.Next() {
		job, err := scanJobRow(rows)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	return jobs, rows.Err()
}

// UpdateStatus updates the mutable outcome fields of a replay job.
func (s *ReplayStore) UpdateStatus(
	ctx context.Context,
	id string,
	status domain.ReplayStatus,
	eventsReplayed int64,
	errMsg string,
) error {
	now := time.Now().UTC()

	var startedAt *time.Time
	var completedAt *time.Time

	switch status {
	case domain.StatusRunning:
		startedAt = &now
	case domain.StatusCompleted, domain.StatusFailed:
		completedAt = &now
	}

	const q = `
		UPDATE replay_jobs
		SET status          = $2,
		    events_replayed = $3,
		    error_message   = $4,
		    started_at      = COALESCE(started_at, $5),
		    completed_at    = $6
		WHERE id = $1`

	res, err := s.db.ExecContext(ctx, q, id, status, eventsReplayed, errMsg, startedAt, completedAt)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// Cancel sets a job's status to cancelled when it is still pending or running.
func (s *ReplayStore) Cancel(ctx context.Context, id string) error {
	const q = `
		UPDATE replay_jobs
		SET status = $2
		WHERE id = $1 AND status IN ('pending','running')`

	res, err := s.db.ExecContext(ctx, q, id, domain.StatusCancelled)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		// Either not found or already in a terminal state — treat both as not found
		// so the handler can return 404 consistently.
		return domain.ErrNotFound
	}
	return nil
}

// ---- helpers ----------------------------------------------------------------

type scanner interface {
	Scan(dest ...any) error
}

func scanJob(s scanner) (*domain.ReplayJob, error) {
	return scanJobFields(s)
}

func scanJobRow(rows *sql.Rows) (*domain.ReplayJob, error) {
	return scanJobFields(rows)
}

func scanJobFields(s scanner) (*domain.ReplayJob, error) {
	var j domain.ReplayJob
	err := s.Scan(
		&j.ID, &j.StreamID, &j.StreamType, &j.EventType,
		&j.FromSeq, &j.ToSeq, &j.FromTime, &j.ToTime,
		&j.Target, &j.TargetTopic, &j.Status, &j.EventsReplayed,
		&j.ErrorMessage, &j.StartedAt, &j.CompletedAt, &j.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &j, nil
}
