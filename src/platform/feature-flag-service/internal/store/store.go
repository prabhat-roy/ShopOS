package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/shopos/feature-flag-service/internal/domain"
)

// Store handles all Postgres interactions for feature flags.
type Store struct {
	db *sql.DB
}

func New(db *sql.DB) *Store { return &Store{db: db} }

func (s *Store) GetByKey(ctx context.Context, key string) (*domain.Flag, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, key, name, description, enabled, strategy,
		       percentage, user_ids, context_key, context_val, created_at, updated_at
		FROM feature_flags WHERE key = $1`, key)
	return scanFlag(row)
}

func (s *Store) GetByID(ctx context.Context, id string) (*domain.Flag, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, key, name, description, enabled, strategy,
		       percentage, user_ids, context_key, context_val, created_at, updated_at
		FROM feature_flags WHERE id = $1`, id)
	return scanFlag(row)
}

func (s *Store) List(ctx context.Context) ([]*domain.Flag, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, key, name, description, enabled, strategy,
		       percentage, user_ids, context_key, context_val, created_at, updated_at
		FROM feature_flags ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var flags []*domain.Flag
	for rows.Next() {
		f, err := scanFlagRow(rows)
		if err != nil {
			return nil, err
		}
		flags = append(flags, f)
	}
	return flags, rows.Err()
}

func (s *Store) Create(ctx context.Context, req *domain.CreateFlagRequest) (*domain.Flag, error) {
	f := &domain.Flag{
		Key:         req.Key,
		Name:        req.Name,
		Description: req.Description,
		Enabled:     req.Enabled,
		Strategy:    req.Strategy,
		Percentage:  req.Percentage,
		UserIDs:     req.UserIDs,
		ContextKey:  req.ContextKey,
		ContextVal:  req.ContextVal,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	err := s.db.QueryRowContext(ctx, `
		INSERT INTO feature_flags
		  (key, name, description, enabled, strategy, percentage, user_ids, context_key, context_val, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		RETURNING id`,
		f.Key, f.Name, f.Description, f.Enabled, string(f.Strategy),
		f.Percentage, pq.Array(f.UserIDs), f.ContextKey, f.ContextVal,
		f.CreatedAt, f.UpdatedAt,
	).Scan(&f.ID)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return nil, domain.ErrAlreadyExists
		}
		return nil, fmt.Errorf("insert flag: %w", err)
	}
	return f, nil
}

func (s *Store) Update(ctx context.Context, id string, req *domain.UpdateFlagRequest) (*domain.Flag, error) {
	setClauses := []string{"updated_at = NOW()"}
	args := []any{}
	i := 1

	if req.Name != nil {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", i))
		args = append(args, *req.Name)
		i++
	}
	if req.Description != nil {
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", i))
		args = append(args, *req.Description)
		i++
	}
	if req.Enabled != nil {
		setClauses = append(setClauses, fmt.Sprintf("enabled = $%d", i))
		args = append(args, *req.Enabled)
		i++
	}
	if req.Strategy != nil {
		setClauses = append(setClauses, fmt.Sprintf("strategy = $%d", i))
		args = append(args, string(*req.Strategy))
		i++
	}
	if req.Percentage != nil {
		setClauses = append(setClauses, fmt.Sprintf("percentage = $%d", i))
		args = append(args, *req.Percentage)
		i++
	}
	if req.UserIDs != nil {
		setClauses = append(setClauses, fmt.Sprintf("user_ids = $%d", i))
		args = append(args, pq.Array(req.UserIDs))
		i++
	}
	if req.ContextKey != nil {
		setClauses = append(setClauses, fmt.Sprintf("context_key = $%d", i))
		args = append(args, *req.ContextKey)
		i++
	}
	if req.ContextVal != nil {
		setClauses = append(setClauses, fmt.Sprintf("context_val = $%d", i))
		args = append(args, *req.ContextVal)
		i++
	}

	args = append(args, id)
	query := fmt.Sprintf(
		"UPDATE feature_flags SET %s WHERE id = $%d RETURNING id, key, name, description, enabled, strategy, percentage, user_ids, context_key, context_val, created_at, updated_at",
		strings.Join(setClauses, ", "), i,
	)

	row := s.db.QueryRowContext(ctx, query, args...)
	return scanFlag(row)
}

func (s *Store) Delete(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, "DELETE FROM feature_flags WHERE id = $1", id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// scanFlag scans a *sql.Row into a Flag.
func scanFlag(row *sql.Row) (*domain.Flag, error) {
	f := &domain.Flag{}
	err := row.Scan(
		&f.ID, &f.Key, &f.Name, &f.Description, &f.Enabled,
		&f.Strategy, &f.Percentage, pq.Array(&f.UserIDs),
		&f.ContextKey, &f.ContextVal, &f.CreatedAt, &f.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return f, err
}

// scanFlagRow scans a *sql.Rows (plural) into a Flag.
func scanFlagRow(rows *sql.Rows) (*domain.Flag, error) {
	f := &domain.Flag{}
	err := rows.Scan(
		&f.ID, &f.Key, &f.Name, &f.Description, &f.Enabled,
		&f.Strategy, &f.Percentage, pq.Array(&f.UserIDs),
		&f.ContextKey, &f.ContextVal, &f.CreatedAt, &f.UpdatedAt,
	)
	return f, err
}
