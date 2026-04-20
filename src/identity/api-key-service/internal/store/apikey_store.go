package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lib/pq"
	"github.com/shopos/api-key-service/internal/domain"
)

// Store implements all Postgres persistence operations for API keys.
type Store struct {
	db *sql.DB
}

// New opens a Postgres connection and verifies it with a ping.
func New(databaseURL string) (*Store, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}
	return &Store{db: db}, nil
}

// Close releases the underlying database connection pool.
func (s *Store) Close() error {
	return s.db.Close()
}

// Create inserts a new APIKey row and returns the stored record.
func (s *Store) Create(ctx context.Context, key *domain.APIKey) (*domain.APIKey, error) {
	const q = `
		INSERT INTO api_keys
			(id, owner_id, owner_type, name, key_prefix, key_hash, scopes, active, expires_at, created_at, updated_at)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
		RETURNING id, owner_id, owner_type, name, key_prefix, key_hash, scopes, active,
		          last_used_at, expires_at, created_at, updated_at`

	row := s.db.QueryRowContext(ctx, q,
		key.ID,
		key.OwnerID,
		key.OwnerType,
		key.Name,
		key.KeyPrefix,
		key.KeyHash,
		pq.Array(key.Scopes),
		key.Active,
		key.ExpiresAt,
	)
	return scanKey(row)
}

// GetByID fetches a single API key by its primary key.
func (s *Store) GetByID(ctx context.Context, id string) (*domain.APIKey, error) {
	const q = `
		SELECT id, owner_id, owner_type, name, key_prefix, key_hash, scopes, active,
		       last_used_at, expires_at, created_at, updated_at
		FROM api_keys WHERE id = $1`

	row := s.db.QueryRowContext(ctx, q, id)
	key, err := scanKey(row)
	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	return key, err
}

// GetByHash fetches a single API key by its SHA-256 hash (used during validation).
func (s *Store) GetByHash(ctx context.Context, hash string) (*domain.APIKey, error) {
	const q = `
		SELECT id, owner_id, owner_type, name, key_prefix, key_hash, scopes, active,
		       last_used_at, expires_at, created_at, updated_at
		FROM api_keys WHERE key_hash = $1`

	row := s.db.QueryRowContext(ctx, q, hash)
	key, err := scanKey(row)
	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	return key, err
}

// List returns all API keys that belong to a given owner.
func (s *Store) List(ctx context.Context, ownerID string) ([]*domain.APIKey, error) {
	const q = `
		SELECT id, owner_id, owner_type, name, key_prefix, key_hash, scopes, active,
		       last_used_at, expires_at, created_at, updated_at
		FROM api_keys WHERE owner_id = $1
		ORDER BY created_at DESC`

	rows, err := s.db.QueryContext(ctx, q, ownerID)
	if err != nil {
		return nil, fmt.Errorf("querying api_keys by owner: %w", err)
	}
	defer rows.Close()

	var keys []*domain.APIKey
	for rows.Next() {
		key, err := scanKeyRow(rows)
		if err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating api_keys rows: %w", err)
	}
	return keys, nil
}

// Deactivate sets active = FALSE for the given key ID.
func (s *Store) Deactivate(ctx context.Context, id string) error {
	const q = `UPDATE api_keys SET active = FALSE, updated_at = NOW() WHERE id = $1`
	res, err := s.db.ExecContext(ctx, q, id)
	if err != nil {
		return fmt.Errorf("deactivating key: %w", err)
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

// TouchLastUsed updates last_used_at to the current time for the given key ID.
func (s *Store) TouchLastUsed(ctx context.Context, id string) error {
	const q = `UPDATE api_keys SET last_used_at = NOW() WHERE id = $1`
	_, err := s.db.ExecContext(ctx, q, id)
	if err != nil {
		return fmt.Errorf("touching last_used_at: %w", err)
	}
	return nil
}

// Delete permanently removes a key row from the database.
func (s *Store) Delete(ctx context.Context, id string) error {
	const q = `DELETE FROM api_keys WHERE id = $1`
	res, err := s.db.ExecContext(ctx, q, id)
	if err != nil {
		return fmt.Errorf("deleting key: %w", err)
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

// ---- helpers ----------------------------------------------------------------

// rowScanner abstracts *sql.Row and *sql.Rows so scanKey can serve both.
type rowScanner interface {
	Scan(dest ...any) error
}

func scanKey(r rowScanner) (*domain.APIKey, error) {
	return scanKeyRow(r)
}

func scanKeyRow(r rowScanner) (*domain.APIKey, error) {
	var k domain.APIKey
	err := r.Scan(
		&k.ID,
		&k.OwnerID,
		&k.OwnerType,
		&k.Name,
		&k.KeyPrefix,
		&k.KeyHash,
		pq.Array(&k.Scopes),
		&k.Active,
		&k.LastUsedAt,
		&k.ExpiresAt,
		&k.CreatedAt,
		&k.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &k, nil
}
