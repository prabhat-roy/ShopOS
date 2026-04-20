package store

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/shopos/brand-service/domain"
	_ "github.com/lib/pq"
)

// Store handles all database interactions for brands.
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
func (s *Store) Close() error { return s.db.Close() }

// Create inserts a new brand row. Returns ErrSlugTaken on unique constraint violation.
func (s *Store) Create(b *domain.Brand) error {
	q := `
		INSERT INTO brands
			(id, name, slug, description, logo_url, website, active, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`
	_, err := s.db.Exec(q,
		b.ID, b.Name, b.Slug, b.Description, b.LogoURL,
		b.Website, b.Active, b.CreatedAt, b.UpdatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrSlugTaken
		}
		return fmt.Errorf("store: create: %w", err)
	}
	return nil
}

// GetByID returns the brand with the given ID or ErrNotFound.
func (s *Store) GetByID(id string) (*domain.Brand, error) {
	q := `SELECT id, name, slug, description, logo_url, website, active, created_at, updated_at
		  FROM brands WHERE id = $1`
	row := s.db.QueryRow(q, id)
	return scanBrand(row)
}

// GetBySlug returns the brand with the given slug or ErrNotFound.
func (s *Store) GetBySlug(slug string) (*domain.Brand, error) {
	q := `SELECT id, name, slug, description, logo_url, website, active, created_at, updated_at
		  FROM brands WHERE slug = $1`
	row := s.db.QueryRow(q, slug)
	return scanBrand(row)
}

// List returns brands, filtered by optional activeOnly flag.
func (s *Store) List(activeOnly bool) ([]*domain.Brand, error) {
	q := `SELECT id, name, slug, description, logo_url, website, active, created_at, updated_at
		  FROM brands WHERE 1=1`
	args := []interface{}{}

	if activeOnly {
		q += " AND active = $1"
		args = append(args, true)
	}
	q += " ORDER BY name ASC"

	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil, fmt.Errorf("store: list: %w", err)
	}
	defer rows.Close()

	var brands []*domain.Brand
	for rows.Next() {
		b, err := scanBrandRow(rows)
		if err != nil {
			return nil, err
		}
		brands = append(brands, b)
	}
	return brands, rows.Err()
}

// Update overwrites mutable fields on an existing brand row.
// Returns ErrNotFound if the row does not exist, ErrSlugTaken on slug conflict.
func (s *Store) Update(b *domain.Brand) error {
	q := `
		UPDATE brands
		SET name=$1, slug=$2, description=$3, logo_url=$4, website=$5, active=$6, updated_at=$7
		WHERE id=$8`
	res, err := s.db.Exec(q,
		b.Name, b.Slug, b.Description, b.LogoURL, b.Website, b.Active, b.UpdatedAt, b.ID,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrSlugTaken
		}
		return fmt.Errorf("store: update: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// Delete performs a soft-delete by setting active=false.
// Returns ErrNotFound if the row does not exist.
func (s *Store) Delete(id string) error {
	q := `UPDATE brands SET active=false, updated_at=NOW() WHERE id=$1`
	res, err := s.db.Exec(q, id)
	if err != nil {
		return fmt.Errorf("store: delete: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// ---- helpers ----------------------------------------------------------------

func scanBrand(row *sql.Row) (*domain.Brand, error) {
	var b domain.Brand
	err := row.Scan(
		&b.ID, &b.Name, &b.Slug, &b.Description,
		&b.LogoURL, &b.Website, &b.Active, &b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("store: scan: %w", err)
	}
	return &b, nil
}

func scanBrandRow(rows *sql.Rows) (*domain.Brand, error) {
	var b domain.Brand
	err := rows.Scan(
		&b.ID, &b.Name, &b.Slug, &b.Description,
		&b.LogoURL, &b.Website, &b.Active, &b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("store: scan row: %w", err)
	}
	return &b, nil
}

// isUniqueViolation checks for Postgres error code 23505 (unique_violation).
func isUniqueViolation(err error) bool {
	return err != nil && len(err.Error()) > 0 &&
		(contains(err.Error(), "23505") || contains(err.Error(), "unique constraint") || contains(err.Error(), "unique_violation"))
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStr(s, sub))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
