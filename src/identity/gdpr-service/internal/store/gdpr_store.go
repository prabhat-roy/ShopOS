package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/lib/pq"

	"github.com/shopos/gdpr-service/internal/domain"
)

// Store defines the persistence contract for GDPR data.
type Store interface {
	CreateRequest(ctx context.Context, req *domain.DataRequest) (*domain.DataRequest, error)
	GetRequest(ctx context.Context, id string) (*domain.DataRequest, error)
	ListRequests(ctx context.Context, userID string) ([]*domain.DataRequest, error)
	UpdateRequestStatus(ctx context.Context, id string, status domain.RequestStatus, notes string) error
	UpsertConsent(ctx context.Context, consent *domain.Consent) error
	GetConsents(ctx context.Context, userID string) ([]*domain.Consent, error)
	GetConsent(ctx context.Context, userID string, consentType domain.ConsentType) (*domain.Consent, error)
}

// postgresStore is the Postgres-backed implementation of Store.
type postgresStore struct {
	db *sql.DB
}

// New opens a Postgres connection pool and returns a Store.
func New(databaseURL string) (Store, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	return &postgresStore{db: db}, nil
}

// CreateRequest inserts a new data subject request row and returns it with
// the timestamps assigned by the database.
func (s *postgresStore) CreateRequest(ctx context.Context, req *domain.DataRequest) (*domain.DataRequest, error) {
	const q = `
		INSERT INTO data_requests (id, user_id, type, status, reason, notes)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, user_id, type, status, reason, notes, completed_at, created_at, updated_at`

	row := s.db.QueryRowContext(ctx, q,
		req.ID, req.UserID, req.Type, req.Status, req.Reason, req.Notes,
	)
	return scanRequest(row)
}

// GetRequest fetches a single data subject request by its primary key.
func (s *postgresStore) GetRequest(ctx context.Context, id string) (*domain.DataRequest, error) {
	const q = `
		SELECT id, user_id, type, status, reason, notes, completed_at, created_at, updated_at
		FROM data_requests WHERE id = $1`

	row := s.db.QueryRowContext(ctx, q, id)
	r, err := scanRequest(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return r, err
}

// ListRequests returns all data subject requests belonging to a user, newest first.
func (s *postgresStore) ListRequests(ctx context.Context, userID string) ([]*domain.DataRequest, error) {
	const q = `
		SELECT id, user_id, type, status, reason, notes, completed_at, created_at, updated_at
		FROM data_requests WHERE user_id = $1 ORDER BY created_at DESC`

	rows, err := s.db.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("list requests: %w", err)
	}
	defer rows.Close()

	var results []*domain.DataRequest
	for rows.Next() {
		r, err := scanRequestRow(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

// UpdateRequestStatus atomically sets a new status, notes, and optionally
// the completed_at timestamp on a data subject request.
func (s *postgresStore) UpdateRequestStatus(ctx context.Context, id string, status domain.RequestStatus, notes string) error {
	var q string
	var args []interface{}

	if status == domain.StatusCompleted {
		q = `UPDATE data_requests
			 SET status = $1, notes = $2, completed_at = NOW(), updated_at = NOW()
			 WHERE id = $3`
		args = []interface{}{status, notes, id}
	} else {
		q = `UPDATE data_requests
			 SET status = $1, notes = $2, updated_at = NOW()
			 WHERE id = $3`
		args = []interface{}{status, notes, id}
	}

	res, err := s.db.ExecContext(ctx, q, args...)
	if err != nil {
		return fmt.Errorf("update request status: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// UpsertConsent inserts or updates a consent record for a (user_id, type) pair.
func (s *postgresStore) UpsertConsent(ctx context.Context, consent *domain.Consent) error {
	const q = `
		INSERT INTO consents (user_id, type, granted, ip_address, updated_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (user_id, type) DO UPDATE
		SET granted = EXCLUDED.granted,
		    ip_address = EXCLUDED.ip_address,
		    updated_at = NOW()`

	_, err := s.db.ExecContext(ctx, q,
		consent.UserID, consent.Type, consent.Granted, consent.IPAddress,
	)
	if err != nil {
		return fmt.Errorf("upsert consent: %w", err)
	}
	return nil
}

// GetConsents returns all consent records for a user.
func (s *postgresStore) GetConsents(ctx context.Context, userID string) ([]*domain.Consent, error) {
	const q = `
		SELECT user_id, type, granted, ip_address, updated_at
		FROM consents WHERE user_id = $1 ORDER BY type`

	rows, err := s.db.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("get consents: %w", err)
	}
	defer rows.Close()

	var results []*domain.Consent
	for rows.Next() {
		c := &domain.Consent{}
		if err := rows.Scan(&c.UserID, &c.Type, &c.Granted, &c.IPAddress, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan consent: %w", err)
		}
		results = append(results, c)
	}
	return results, rows.Err()
}

// GetConsent fetches a single consent record for a (user_id, consentType) pair.
func (s *postgresStore) GetConsent(ctx context.Context, userID string, consentType domain.ConsentType) (*domain.Consent, error) {
	const q = `
		SELECT user_id, type, granted, ip_address, updated_at
		FROM consents WHERE user_id = $1 AND type = $2`

	c := &domain.Consent{}
	err := s.db.QueryRowContext(ctx, q, userID, consentType).
		Scan(&c.UserID, &c.Type, &c.Granted, &c.IPAddress, &c.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get consent: %w", err)
	}
	return c, nil
}

// ---------- helpers ----------

func scanRequest(row *sql.Row) (*domain.DataRequest, error) {
	r := &domain.DataRequest{}
	err := row.Scan(
		&r.ID, &r.UserID, &r.Type, &r.Status,
		&r.Reason, &r.Notes, &r.CompletedAt,
		&r.CreatedAt, &r.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func scanRequestRow(rows *sql.Rows) (*domain.DataRequest, error) {
	r := &domain.DataRequest{}
	err := rows.Scan(
		&r.ID, &r.UserID, &r.Type, &r.Status,
		&r.Reason, &r.Notes, &r.CompletedAt,
		&r.CreatedAt, &r.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan request row: %w", err)
	}
	return r, nil
}
