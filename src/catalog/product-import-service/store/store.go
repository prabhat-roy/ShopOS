package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopos/product-import-service/domain"
)

// Storer defines the persistence contract for ImportJob.
type Storer interface {
	Create(fileName string, format domain.ImportFormat) (*domain.ImportJob, error)
	Get(id string) (*domain.ImportJob, error)
	List() ([]*domain.ImportJob, error)
	UpdateProgress(id string, processed, errorRows int, errs []domain.ImportError) error
	Complete(id string) error
	Fail(id string, errMsg string) error
}

// PostgresStore is the production Postgres implementation of Storer.
type PostgresStore struct {
	db *sql.DB
}

// NewPostgresStore creates a PostgresStore backed by db.
func NewPostgresStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

// Create inserts a new ImportJob with PENDING status and returns it.
func (s *PostgresStore) Create(fileName string, format domain.ImportFormat) (*domain.ImportJob, error) {
	id := uuid.New().String()
	now := time.Now().UTC()
	errorsJSON := []byte("[]")

	const q = `
		INSERT INTO import_jobs (id, file_name, format, status, total_rows, processed_rows, error_rows, errors, created_at)
		VALUES ($1, $2, $3, $4, 0, 0, 0, $5, $6)
	`
	_, err := s.db.Exec(q, id, fileName, string(format), string(domain.ImportPending), errorsJSON, now)
	if err != nil {
		return nil, fmt.Errorf("store.Create: %w", err)
	}
	return s.Get(id)
}

// Get retrieves an ImportJob by ID.
func (s *PostgresStore) Get(id string) (*domain.ImportJob, error) {
	const q = `
		SELECT id, file_name, format, status, total_rows, processed_rows, error_rows, errors, created_at, completed_at
		FROM import_jobs
		WHERE id = $1
	`
	row := s.db.QueryRow(q, id)
	return scanJob(row)
}

// List returns all ImportJobs ordered by creation time descending.
func (s *PostgresStore) List() ([]*domain.ImportJob, error) {
	const q = `
		SELECT id, file_name, format, status, total_rows, processed_rows, error_rows, errors, created_at, completed_at
		FROM import_jobs
		ORDER BY created_at DESC
	`
	rows, err := s.db.Query(q)
	if err != nil {
		return nil, fmt.Errorf("store.List: %w", err)
	}
	defer rows.Close()

	var jobs []*domain.ImportJob
	for rows.Next() {
		job, err := scanJob(rows)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store.List rows: %w", err)
	}
	return jobs, nil
}

// UpdateProgress updates processed/error counts and appends new errors to the JSONB column.
func (s *PostgresStore) UpdateProgress(id string, processed, errorRows int, errs []domain.ImportError) error {
	errorsJSON, err := json.Marshal(errs)
	if err != nil {
		return fmt.Errorf("store.UpdateProgress marshal: %w", err)
	}
	const q = `
		UPDATE import_jobs
		SET status = $1, processed_rows = $2, error_rows = $3, errors = $4
		WHERE id = $5
	`
	res, err := s.db.Exec(q, string(domain.ImportProcessing), processed, errorRows, errorsJSON, id)
	if err != nil {
		return fmt.Errorf("store.UpdateProgress: %w", err)
	}
	return expectOneRow(res, id)
}

// Complete marks a job as COMPLETED and records the completion timestamp.
func (s *PostgresStore) Complete(id string) error {
	now := time.Now().UTC()
	const q = `UPDATE import_jobs SET status = $1, completed_at = $2 WHERE id = $3`
	res, err := s.db.Exec(q, string(domain.ImportCompleted), now, id)
	if err != nil {
		return fmt.Errorf("store.Complete: %w", err)
	}
	return expectOneRow(res, id)
}

// Fail marks a job as FAILED and stores the failure reason as a single error entry.
func (s *PostgresStore) Fail(id string, errMsg string) error {
	now := time.Now().UTC()
	errs := []domain.ImportError{{Row: 0, Field: "job", Message: errMsg}}
	errorsJSON, _ := json.Marshal(errs)
	const q = `UPDATE import_jobs SET status = $1, completed_at = $2, errors = $3 WHERE id = $4`
	res, err := s.db.Exec(q, string(domain.ImportFailed), now, errorsJSON, id)
	if err != nil {
		return fmt.Errorf("store.Fail: %w", err)
	}
	return expectOneRow(res, id)
}

// ---- helpers ----------------------------------------------------------------

type scanner interface {
	Scan(dest ...any) error
}

func scanJob(s scanner) (*domain.ImportJob, error) {
	var (
		job         domain.ImportJob
		formatStr   string
		statusStr   string
		errorsJSON  []byte
		completedAt sql.NullTime
	)
	err := s.Scan(
		&job.ID,
		&job.FileName,
		&formatStr,
		&statusStr,
		&job.TotalRows,
		&job.ProcessedRows,
		&job.ErrorRows,
		&errorsJSON,
		&job.CreatedAt,
		&completedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("import job not found")
	}
	if err != nil {
		return nil, fmt.Errorf("store.scanJob: %w", err)
	}
	job.Format = domain.ImportFormat(formatStr)
	job.Status = domain.ImportStatus(statusStr)
	if completedAt.Valid {
		t := completedAt.Time
		job.CompletedAt = &t
	}
	if len(errorsJSON) > 0 {
		if err := json.Unmarshal(errorsJSON, &job.Errors); err != nil {
			return nil, fmt.Errorf("store.scanJob unmarshal errors: %w", err)
		}
	}
	if job.Errors == nil {
		job.Errors = []domain.ImportError{}
	}
	return &job, nil
}

func expectOneRow(res sql.Result, id string) error {
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("import job not found: %s", id)
	}
	return nil
}
