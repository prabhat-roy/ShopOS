package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/shopos/permission-service/internal/domain"
)

// Store defines the data-access contract for the permission service.
type Store interface {
	CreateRole(ctx context.Context, role *domain.Role) (*domain.Role, error)
	GetRole(ctx context.Context, id string) (*domain.Role, error)
	ListRoles(ctx context.Context) ([]*domain.Role, error)
	DeleteRole(ctx context.Context, id string) error

	AssignRole(ctx context.Context, userID, roleID string) error
	RevokeRole(ctx context.Context, userID, roleID string) error
	GetUserRoles(ctx context.Context, userID string) ([]*domain.UserRole, error)
	GetUserPermissions(ctx context.Context, userID string) ([]string, error)
}

// PostgresStore is a Postgres-backed implementation of Store.
type PostgresStore struct {
	db *sql.DB
}

// NewPostgresStore opens a connection pool to Postgres and verifies connectivity.
func NewPostgresStore(databaseURL string, maxConns int) (*PostgresStore, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	db.SetMaxOpenConns(maxConns)
	db.SetMaxIdleConns(maxConns / 2)
	db.SetConnMaxLifetime(5 * time.Minute)
	return &PostgresStore{db: db}, nil
}

// Close closes the underlying connection pool.
func (s *PostgresStore) Close() error {
	return s.db.Close()
}

// CreateRole inserts a new role row and returns the persisted role.
func (s *PostgresStore) CreateRole(ctx context.Context, role *domain.Role) (*domain.Role, error) {
	const q = `
		INSERT INTO roles (id, name, description, permissions, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, name, description, permissions, created_at`

	row := s.db.QueryRowContext(ctx, q,
		role.ID,
		role.Name,
		role.Description,
		pq.Array(role.Permissions),
		role.CreatedAt,
	)
	return scanRole(row)
}

// GetRole fetches a single role by primary key.
func (s *PostgresStore) GetRole(ctx context.Context, id string) (*domain.Role, error) {
	const q = `SELECT id, name, description, permissions, created_at FROM roles WHERE id = $1`
	row := s.db.QueryRowContext(ctx, q, id)
	r, err := scanRole(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return r, err
}

// ListRoles returns all roles ordered by name.
func (s *PostgresStore) ListRoles(ctx context.Context) ([]*domain.Role, error) {
	const q = `SELECT id, name, description, permissions, created_at FROM roles ORDER BY name`
	rows, err := s.db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("list roles query: %w", err)
	}
	defer rows.Close()

	var roles []*domain.Role
	for rows.Next() {
		r, err := scanRoleRow(rows)
		if err != nil {
			return nil, err
		}
		roles = append(roles, r)
	}
	return roles, rows.Err()
}

// DeleteRole removes a role by primary key. Cascades to user_roles via FK.
func (s *PostgresStore) DeleteRole(ctx context.Context, id string) error {
	const q = `DELETE FROM roles WHERE id = $1`
	res, err := s.db.ExecContext(ctx, q, id)
	if err != nil {
		return fmt.Errorf("delete role: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// AssignRole creates a user_roles row binding userID to roleID.
func (s *PostgresStore) AssignRole(ctx context.Context, userID, roleID string) error {
	const q = `
		INSERT INTO user_roles (user_id, role_id, assigned_at)
		VALUES ($1, $2, NOW())`

	_, err := s.db.ExecContext(ctx, q, userID, roleID)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case "23505": // unique_violation — primary key duplicate
				return domain.ErrAlreadyAssigned
			case "23503": // foreign_key_violation — role does not exist
				return domain.ErrNotFound
			}
		}
		return fmt.Errorf("assign role: %w", err)
	}
	return nil
}

// RevokeRole removes the user_roles row for the given pair.
func (s *PostgresStore) RevokeRole(ctx context.Context, userID, roleID string) error {
	const q = `DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2`
	res, err := s.db.ExecContext(ctx, q, userID, roleID)
	if err != nil {
		return fmt.Errorf("revoke role: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// GetUserRoles returns all roles assigned to a user.
func (s *PostgresStore) GetUserRoles(ctx context.Context, userID string) ([]*domain.UserRole, error) {
	const q = `
		SELECT ur.user_id, ur.role_id, r.name, ur.assigned_at
		FROM user_roles ur
		JOIN roles r ON r.id = ur.role_id
		WHERE ur.user_id = $1
		ORDER BY r.name`

	rows, err := s.db.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("get user roles: %w", err)
	}
	defer rows.Close()

	var out []*domain.UserRole
	for rows.Next() {
		ur := &domain.UserRole{}
		if err := rows.Scan(&ur.UserID, &ur.RoleID, &ur.RoleName, &ur.AssignedAt); err != nil {
			return nil, fmt.Errorf("scan user role: %w", err)
		}
		out = append(out, ur)
	}
	return out, rows.Err()
}

// GetUserPermissions returns the de-duplicated set of permissions held by a user
// across all their assigned roles.
func (s *PostgresStore) GetUserPermissions(ctx context.Context, userID string) ([]string, error) {
	const q = `
		SELECT DISTINCT unnest(r.permissions) AS perm
		FROM user_roles ur
		JOIN roles r ON r.id = ur.role_id
		WHERE ur.user_id = $1
		ORDER BY perm`

	rows, err := s.db.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("get user permissions: %w", err)
	}
	defer rows.Close()

	var perms []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return nil, fmt.Errorf("scan permission: %w", err)
		}
		perms = append(perms, p)
	}
	return perms, rows.Err()
}

// ---- helpers ----------------------------------------------------------------

func scanRole(row *sql.Row) (*domain.Role, error) {
	r := &domain.Role{}
	var perms pq.StringArray
	if err := row.Scan(&r.ID, &r.Name, &r.Description, &perms, &r.CreatedAt); err != nil {
		return nil, err
	}
	r.Permissions = []string(perms)
	return r, nil
}

func scanRoleRow(rows *sql.Rows) (*domain.Role, error) {
	r := &domain.Role{}
	var perms pq.StringArray
	if err := rows.Scan(&r.ID, &r.Name, &r.Description, &perms, &r.CreatedAt); err != nil {
		return nil, fmt.Errorf("scan role row: %w", err)
	}
	r.Permissions = []string(perms)
	return r, nil
}
