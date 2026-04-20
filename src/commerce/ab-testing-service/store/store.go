package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/lib/pq"

	"github.com/shopos/ab-testing-service/domain"
)

const migrationSQL = `
CREATE TABLE IF NOT EXISTS experiments (
    id               TEXT PRIMARY KEY,
    name             TEXT NOT NULL,
    description      TEXT NOT NULL DEFAULT '',
    variants         JSONB NOT NULL DEFAULT '[]',
    active           BOOLEAN NOT NULL DEFAULT TRUE,
    traffic_percent  INTEGER NOT NULL DEFAULT 100,
    created_at       TIMESTAMPTZ NOT NULL,
    updated_at       TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS assignments (
    experiment_id  TEXT NOT NULL,
    user_id        TEXT NOT NULL,
    variant        TEXT NOT NULL,
    assigned_at    TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (experiment_id, user_id)
);

CREATE TABLE IF NOT EXISTS conversions (
    id             BIGSERIAL PRIMARY KEY,
    experiment_id  TEXT NOT NULL,
    user_id        TEXT NOT NULL,
    metric         TEXT NOT NULL,
    value          DOUBLE PRECISION NOT NULL DEFAULT 0,
    recorded_at    TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_conversions_experiment ON conversions (experiment_id);
`

// Store handles all persistence for the ab-testing-service.
type Store struct {
	db *sql.DB
}

// New opens a Postgres connection, runs migrations, and returns a Store.
func New(databaseURL string) (*Store, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	db.Exec(migrationSQL) //nolint:errcheck — best-effort at startup; fails gracefully when DB is unavailable
	return &Store{db: db}, nil
}

// Close shuts down the database pool.
func (s *Store) Close() error { return s.db.Close() }

// SaveExperiment inserts or updates an experiment.
func (s *Store) SaveExperiment(ctx context.Context, e *domain.Experiment) error {
	variantsJSON, err := json.Marshal(e.Variants)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO experiments (id, name, description, variants, active, traffic_percent, created_at, updated_at)
		VALUES ($1,$2,$3,$4::jsonb,$5,$6,$7,$8)
		ON CONFLICT (id) DO UPDATE SET
		    name=EXCLUDED.name, description=EXCLUDED.description,
		    variants=EXCLUDED.variants, active=EXCLUDED.active,
		    traffic_percent=EXCLUDED.traffic_percent, updated_at=EXCLUDED.updated_at`,
		e.ID, e.Name, e.Description, string(variantsJSON),
		e.Active, e.TrafficPercent, e.CreatedAt, e.UpdatedAt,
	)
	return err
}

// GetExperiment retrieves an experiment by ID.
func (s *Store) GetExperiment(ctx context.Context, id string) (*domain.Experiment, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, name, description, variants, active, traffic_percent, created_at, updated_at
		FROM experiments WHERE id=$1`, id)
	return scanExperiment(row)
}

// ListExperiments returns all experiments ordered by created_at desc.
func (s *Store) ListExperiments(ctx context.Context) ([]*domain.Experiment, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, description, variants, active, traffic_percent, created_at, updated_at
		FROM experiments ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*domain.Experiment
	for rows.Next() {
		e, err := scanExperiment(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, rows.Err()
}

type scanner interface {
	Scan(dest ...any) error
}

func scanExperiment(s scanner) (*domain.Experiment, error) {
	var e domain.Experiment
	var variantsJSON string
	err := s.Scan(
		&e.ID, &e.Name, &e.Description, &variantsJSON,
		&e.Active, &e.TrafficPercent, &e.CreatedAt, &e.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(variantsJSON), &e.Variants); err != nil {
		return nil, err
	}
	return &e, nil
}

// SaveAssignment inserts or updates a user-to-variant assignment.
func (s *Store) SaveAssignment(ctx context.Context, a *domain.Assignment) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO assignments (experiment_id, user_id, variant, assigned_at)
		VALUES ($1,$2,$3,$4)
		ON CONFLICT (experiment_id, user_id) DO NOTHING`,
		a.ExperimentID, a.UserID, a.Variant, a.AssignedAt,
	)
	return err
}

// GetAssignment retrieves an existing assignment.
func (s *Store) GetAssignment(ctx context.Context, experimentID, userID string) (*domain.Assignment, error) {
	var a domain.Assignment
	err := s.db.QueryRowContext(ctx, `
		SELECT experiment_id, user_id, variant, assigned_at
		FROM assignments WHERE experiment_id=$1 AND user_id=$2`,
		experimentID, userID,
	).Scan(&a.ExperimentID, &a.UserID, &a.Variant, &a.AssignedAt)
	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	return &a, err
}

// RecordConversion saves a conversion event.
func (s *Store) RecordConversion(ctx context.Context, c *domain.Conversion) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO conversions (experiment_id, user_id, metric, value, recorded_at)
		VALUES ($1,$2,$3,$4,$5)`,
		c.ExperimentID, c.UserID, c.Metric, c.Value, c.RecordedAt,
	)
	return err
}
