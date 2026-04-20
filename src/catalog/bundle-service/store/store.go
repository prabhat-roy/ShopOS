package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/shopos/bundle-service/domain"
)

// Store handles all Postgres persistence for the bundle service.
type Store struct {
	db *sql.DB
}

// New opens a Postgres connection and returns a ready Store.
func New(dsn string) (*Store, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("store: open: %w", err)
	}
	return &Store{db: db}, nil
}

// Close releases the underlying database connection pool.
func (s *Store) Close() error {
	return s.db.Close()
}

// Create persists a new bundle and returns it with generated ID and timestamps.
func (s *Store) Create(ctx context.Context, b *domain.Bundle) (*domain.Bundle, error) {
	b.ID = uuid.New().String()
	now := time.Now().UTC()
	b.CreatedAt = now
	b.UpdatedAt = now

	itemsJSON, err := json.Marshal(b.Items)
	if err != nil {
		return nil, fmt.Errorf("store: Create: marshal items: %w", err)
	}

	const q = `
		INSERT INTO bundles (id, name, description, price, currency, items, active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	if _, err := s.db.ExecContext(ctx, q,
		b.ID, b.Name, b.Description, b.Price, b.Currency,
		string(itemsJSON), b.Active, b.CreatedAt, b.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("store: Create: %w", err)
	}
	return b, nil
}

// GetByID returns a bundle by primary key. Returns domain.ErrNotFound when absent.
func (s *Store) GetByID(ctx context.Context, id string) (*domain.Bundle, error) {
	const q = `
		SELECT id, name, description, price, currency, items, active, created_at, updated_at
		FROM bundles
		WHERE id = $1`

	row := s.db.QueryRowContext(ctx, q, id)
	b, err := scanBundle(row)
	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("store: GetByID: %w", err)
	}
	return b, nil
}

// List returns all bundles. If activeOnly is true only active bundles are returned.
func (s *Store) List(ctx context.Context, activeOnly bool) ([]*domain.Bundle, error) {
	q := `
		SELECT id, name, description, price, currency, items, active, created_at, updated_at
		FROM bundles`
	args := []interface{}{}
	if activeOnly {
		q += " WHERE active = $1"
		args = append(args, true)
	}
	q += " ORDER BY created_at DESC"

	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("store: List: %w", err)
	}
	defer rows.Close()

	var bundles []*domain.Bundle
	for rows.Next() {
		b, err := scanBundleRows(rows)
		if err != nil {
			return nil, fmt.Errorf("store: List scan: %w", err)
		}
		bundles = append(bundles, b)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: List rows: %w", err)
	}
	return bundles, nil
}

// Update applies partial changes to a bundle. Only non-zero fields are written.
func (s *Store) Update(ctx context.Context, id string, patch *domain.Bundle) (*domain.Bundle, error) {
	existing, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if patch.Name != "" {
		existing.Name = patch.Name
	}
	if patch.Description != "" {
		existing.Description = patch.Description
	}
	if patch.Price != 0 {
		existing.Price = patch.Price
	}
	if patch.Currency != "" {
		existing.Currency = patch.Currency
	}
	if patch.Items != nil {
		existing.Items = patch.Items
	}
	// Active is a bool — always apply the patched value.
	existing.Active = patch.Active
	existing.UpdatedAt = time.Now().UTC()

	itemsJSON, err := json.Marshal(existing.Items)
	if err != nil {
		return nil, fmt.Errorf("store: Update: marshal items: %w", err)
	}

	const q = `
		UPDATE bundles
		SET name        = $1,
		    description = $2,
		    price       = $3,
		    currency    = $4,
		    items       = $5,
		    active      = $6,
		    updated_at  = $7
		WHERE id = $8`
	if _, err := s.db.ExecContext(ctx, q,
		existing.Name, existing.Description, existing.Price, existing.Currency,
		string(itemsJSON), existing.Active, existing.UpdatedAt, id,
	); err != nil {
		return nil, fmt.Errorf("store: Update: %w", err)
	}
	return existing, nil
}

// Delete performs a soft delete by setting active = false.
func (s *Store) Delete(ctx context.Context, id string) error {
	const q = `UPDATE bundles SET active = FALSE, updated_at = NOW() WHERE id = $1`
	res, err := s.db.ExecContext(ctx, q, id)
	if err != nil {
		return fmt.Errorf("store: Delete: %w", err)
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// scanBundle scans a *sql.Row into a Bundle.
func scanBundle(row *sql.Row) (*domain.Bundle, error) {
	b := &domain.Bundle{}
	var itemsRaw string
	err := row.Scan(
		&b.ID, &b.Name, &b.Description, &b.Price, &b.Currency,
		&itemsRaw, &b.Active, &b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(itemsRaw), &b.Items); err != nil {
		return nil, fmt.Errorf("unmarshal items: %w", err)
	}
	return b, nil
}

// scanBundleRows scans a *sql.Rows into a Bundle.
func scanBundleRows(rows *sql.Rows) (*domain.Bundle, error) {
	b := &domain.Bundle{}
	var itemsRaw string
	err := rows.Scan(
		&b.ID, &b.Name, &b.Description, &b.Price, &b.Currency,
		&itemsRaw, &b.Active, &b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(itemsRaw), &b.Items); err != nil {
		return nil, fmt.Errorf("unmarshal items: %w", err)
	}
	return b, nil
}
