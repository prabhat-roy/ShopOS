package store

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/shopos/category-service/domain"
	_ "github.com/lib/pq"
)

// Store handles all database interactions for categories.
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

// Create inserts a new category row. Returns ErrSlugTaken on unique constraint violation.
func (s *Store) Create(c *domain.Category) error {
	q := `
		INSERT INTO categories
			(id, name, slug, parent_id, description, image_url, sort_order, active, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`
	_, err := s.db.Exec(q,
		c.ID, c.Name, c.Slug, c.ParentID, c.Description,
		c.ImageURL, c.SortOrder, c.Active, c.CreatedAt, c.UpdatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrSlugTaken
		}
		return fmt.Errorf("store: create: %w", err)
	}
	return nil
}

// GetByID returns the category with the given ID or ErrNotFound.
func (s *Store) GetByID(id string) (*domain.Category, error) {
	q := `SELECT id, name, slug, parent_id, description, image_url, sort_order, active, created_at, updated_at
		  FROM categories WHERE id = $1`
	row := s.db.QueryRow(q, id)
	return scanCategory(row)
}

// GetBySlug returns the category with the given slug or ErrNotFound.
func (s *Store) GetBySlug(slug string) (*domain.Category, error) {
	q := `SELECT id, name, slug, parent_id, description, image_url, sort_order, active, created_at, updated_at
		  FROM categories WHERE slug = $1`
	row := s.db.QueryRow(q, slug)
	return scanCategory(row)
}

// List returns categories filtered by optional parentID and/or activeOnly flag.
func (s *Store) List(parentID *string, activeOnly bool) ([]*domain.Category, error) {
	q := `SELECT id, name, slug, parent_id, description, image_url, sort_order, active, created_at, updated_at
		  FROM categories WHERE 1=1`
	args := []interface{}{}
	idx := 1

	if parentID != nil {
		q += fmt.Sprintf(" AND parent_id = $%d", idx)
		args = append(args, *parentID)
		idx++
	}
	if activeOnly {
		q += fmt.Sprintf(" AND active = $%d", idx)
		args = append(args, true)
	}
	q += " ORDER BY sort_order ASC, name ASC"

	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil, fmt.Errorf("store: list: %w", err)
	}
	defer rows.Close()

	var cats []*domain.Category
	for rows.Next() {
		c, err := scanCategoryRow(rows)
		if err != nil {
			return nil, err
		}
		cats = append(cats, c)
	}
	return cats, rows.Err()
}

// Update overwrites mutable fields on an existing category row.
// Returns ErrNotFound if the row does not exist, ErrSlugTaken on slug conflict.
func (s *Store) Update(c *domain.Category) error {
	q := `
		UPDATE categories
		SET name=$1, slug=$2, parent_id=$3, description=$4, image_url=$5,
		    sort_order=$6, active=$7, updated_at=$8
		WHERE id=$9`
	res, err := s.db.Exec(q,
		c.Name, c.Slug, c.ParentID, c.Description, c.ImageURL,
		c.SortOrder, c.Active, c.UpdatedAt, c.ID,
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
	q := `UPDATE categories SET active=false, updated_at=NOW() WHERE id=$1`
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

func scanCategory(row *sql.Row) (*domain.Category, error) {
	var c domain.Category
	err := row.Scan(
		&c.ID, &c.Name, &c.Slug, &c.ParentID, &c.Description,
		&c.ImageURL, &c.SortOrder, &c.Active, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("store: scan: %w", err)
	}
	return &c, nil
}

func scanCategoryRow(rows *sql.Rows) (*domain.Category, error) {
	var c domain.Category
	err := rows.Scan(
		&c.ID, &c.Name, &c.Slug, &c.ParentID, &c.Description,
		&c.ImageURL, &c.SortOrder, &c.Active, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("store: scan row: %w", err)
	}
	return &c, nil
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
