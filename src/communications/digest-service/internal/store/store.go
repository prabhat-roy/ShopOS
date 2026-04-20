package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/shopos/digest-service/internal/domain"
)

// Storer defines all data-access operations for the digest-service.
type Storer interface {
	CreateConfig(ctx context.Context, cfg domain.DigestConfig) error
	GetConfig(ctx context.Context, id uuid.UUID) (domain.DigestConfig, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]domain.DigestConfig, error)
	ListConfigs(ctx context.Context, status domain.DigestStatus, frequency domain.DigestFrequency) ([]domain.DigestConfig, error)
	UpdateNextSend(ctx context.Context, id uuid.UUID, nextSendAt time.Time) error
	UpdateLastSent(ctx context.Context, id uuid.UUID, sentAt time.Time) error
	PauseConfig(ctx context.Context, id uuid.UUID) error
	ResumeConfig(ctx context.Context, id uuid.UUID) error
	DeleteConfig(ctx context.Context, id uuid.UUID) error
	SaveRun(ctx context.Context, run domain.DigestRun) error
	ListRuns(ctx context.Context, configID uuid.UUID, limit int) ([]domain.DigestRun, error)
	ListDueConfigs(ctx context.Context, now time.Time) ([]domain.DigestConfig, error)
}

// PostgresStore implements Storer backed by PostgreSQL.
type PostgresStore struct {
	db *sql.DB
}

// New opens a PostgreSQL connection and returns a PostgresStore.
func New(databaseURL string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %w", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)
	return &PostgresStore{db: db}, nil
}

// Ping verifies the database connection.
func (s *PostgresStore) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

// Close closes the underlying database connection pool.
func (s *PostgresStore) Close() error {
	return s.db.Close()
}

// CreateConfig inserts a new digest configuration row.
func (s *PostgresStore) CreateConfig(ctx context.Context, cfg domain.DigestConfig) error {
	const q = `
INSERT INTO digest_configs (id, user_id, email, frequency, status, last_sent_at, next_send_at, timezone, created_at, updated_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`
	_, err := s.db.ExecContext(ctx, q,
		cfg.ID, cfg.UserID, cfg.Email, cfg.Frequency, cfg.Status,
		cfg.LastSentAt, cfg.NextSendAt, cfg.Timezone, cfg.CreatedAt, cfg.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("CreateConfig exec: %w", err)
	}
	return nil
}

// GetConfig retrieves a single digest configuration by primary key.
func (s *PostgresStore) GetConfig(ctx context.Context, id uuid.UUID) (domain.DigestConfig, error) {
	const q = `
SELECT id, user_id, email, frequency, status, last_sent_at, next_send_at, timezone, created_at, updated_at
FROM digest_configs WHERE id = $1`
	row := s.db.QueryRowContext(ctx, q, id)
	return scanConfig(row)
}

// GetByUserID returns all digest configurations for a given user.
func (s *PostgresStore) GetByUserID(ctx context.Context, userID uuid.UUID) ([]domain.DigestConfig, error) {
	const q = `
SELECT id, user_id, email, frequency, status, last_sent_at, next_send_at, timezone, created_at, updated_at
FROM digest_configs WHERE user_id = $1 ORDER BY created_at DESC`
	rows, err := s.db.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("GetByUserID query: %w", err)
	}
	defer rows.Close()
	return scanConfigs(rows)
}

// ListConfigs returns configurations filtered by optional status and frequency.
// Empty string values mean "no filter".
func (s *PostgresStore) ListConfigs(ctx context.Context, status domain.DigestStatus, frequency domain.DigestFrequency) ([]domain.DigestConfig, error) {
	q := `SELECT id, user_id, email, frequency, status, last_sent_at, next_send_at, timezone, created_at, updated_at
FROM digest_configs WHERE 1=1`
	args := []interface{}{}
	idx := 1
	if status != "" {
		q += fmt.Sprintf(" AND status = $%d", idx)
		args = append(args, status)
		idx++
	}
	if frequency != "" {
		q += fmt.Sprintf(" AND frequency = $%d", idx)
		args = append(args, frequency)
		idx++
	}
	q += " ORDER BY created_at DESC"
	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("ListConfigs query: %w", err)
	}
	defer rows.Close()
	return scanConfigs(rows)
}

// UpdateNextSend sets the next_send_at timestamp for a config.
func (s *PostgresStore) UpdateNextSend(ctx context.Context, id uuid.UUID, nextSendAt time.Time) error {
	const q = `UPDATE digest_configs SET next_send_at=$1, updated_at=NOW() WHERE id=$2`
	_, err := s.db.ExecContext(ctx, q, nextSendAt, id)
	if err != nil {
		return fmt.Errorf("UpdateNextSend: %w", err)
	}
	return nil
}

// UpdateLastSent sets the last_sent_at timestamp for a config.
func (s *PostgresStore) UpdateLastSent(ctx context.Context, id uuid.UUID, sentAt time.Time) error {
	const q = `UPDATE digest_configs SET last_sent_at=$1, updated_at=NOW() WHERE id=$2`
	_, err := s.db.ExecContext(ctx, q, sentAt, id)
	if err != nil {
		return fmt.Errorf("UpdateLastSent: %w", err)
	}
	return nil
}

// PauseConfig sets status=PAUSED for a config.
func (s *PostgresStore) PauseConfig(ctx context.Context, id uuid.UUID) error {
	const q = `UPDATE digest_configs SET status='PAUSED', updated_at=NOW() WHERE id=$1`
	_, err := s.db.ExecContext(ctx, q, id)
	if err != nil {
		return fmt.Errorf("PauseConfig: %w", err)
	}
	return nil
}

// ResumeConfig sets status=ACTIVE for a config.
func (s *PostgresStore) ResumeConfig(ctx context.Context, id uuid.UUID) error {
	const q = `UPDATE digest_configs SET status='ACTIVE', updated_at=NOW() WHERE id=$1`
	_, err := s.db.ExecContext(ctx, q, id)
	if err != nil {
		return fmt.Errorf("ResumeConfig: %w", err)
	}
	return nil
}

// DeleteConfig removes a digest configuration and its associated runs.
func (s *PostgresStore) DeleteConfig(ctx context.Context, id uuid.UUID) error {
	const q = `DELETE FROM digest_configs WHERE id=$1`
	_, err := s.db.ExecContext(ctx, q, id)
	if err != nil {
		return fmt.Errorf("DeleteConfig: %w", err)
	}
	return nil
}

// SaveRun inserts a digest run record.
func (s *PostgresStore) SaveRun(ctx context.Context, run domain.DigestRun) error {
	const q = `
INSERT INTO digest_runs (id, config_id, sent_at, item_count, status, error_msg)
VALUES ($1,$2,$3,$4,$5,$6)`
	_, err := s.db.ExecContext(ctx, q, run.ID, run.ConfigID, run.SentAt, run.ItemCount, run.Status, run.ErrorMsg)
	if err != nil {
		return fmt.Errorf("SaveRun: %w", err)
	}
	return nil
}

// ListRuns returns the most recent runs for a config, limited to limit rows.
func (s *PostgresStore) ListRuns(ctx context.Context, configID uuid.UUID, limit int) ([]domain.DigestRun, error) {
	if limit <= 0 {
		limit = 20
	}
	const q = `
SELECT id, config_id, sent_at, item_count, status, error_msg
FROM digest_runs WHERE config_id=$1 ORDER BY sent_at DESC LIMIT $2`
	rows, err := s.db.QueryContext(ctx, q, configID, limit)
	if err != nil {
		return nil, fmt.Errorf("ListRuns query: %w", err)
	}
	defer rows.Close()
	var runs []domain.DigestRun
	for rows.Next() {
		var r domain.DigestRun
		if err := rows.Scan(&r.ID, &r.ConfigID, &r.SentAt, &r.ItemCount, &r.Status, &r.ErrorMsg); err != nil {
			return nil, fmt.Errorf("ListRuns scan: %w", err)
		}
		runs = append(runs, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ListRuns rows: %w", err)
	}
	if runs == nil {
		runs = []domain.DigestRun{}
	}
	return runs, nil
}

// ListDueConfigs returns all ACTIVE configs whose next_send_at is on or before now.
func (s *PostgresStore) ListDueConfigs(ctx context.Context, now time.Time) ([]domain.DigestConfig, error) {
	const q = `
SELECT id, user_id, email, frequency, status, last_sent_at, next_send_at, timezone, created_at, updated_at
FROM digest_configs
WHERE status = 'ACTIVE' AND next_send_at <= $1
ORDER BY next_send_at ASC`
	rows, err := s.db.QueryContext(ctx, q, now)
	if err != nil {
		return nil, fmt.Errorf("ListDueConfigs query: %w", err)
	}
	defer rows.Close()
	return scanConfigs(rows)
}

// ── helpers ──────────────────────────────────────────────────────────────────

type rowScanner interface {
	Scan(dest ...any) error
}

func scanConfig(row rowScanner) (domain.DigestConfig, error) {
	var c domain.DigestConfig
	err := row.Scan(
		&c.ID, &c.UserID, &c.Email, &c.Frequency, &c.Status,
		&c.LastSentAt, &c.NextSendAt, &c.Timezone, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return domain.DigestConfig{}, fmt.Errorf("scanConfig: %w", err)
	}
	return c, nil
}

func scanConfigs(rows *sql.Rows) ([]domain.DigestConfig, error) {
	var cfgs []domain.DigestConfig
	for rows.Next() {
		c, err := scanConfig(rows)
		if err != nil {
			return nil, err
		}
		cfgs = append(cfgs, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("scanConfigs rows: %w", err)
	}
	if cfgs == nil {
		cfgs = []domain.DigestConfig{}
	}
	return cfgs, nil
}
