package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/lib/pq"
	"github.com/shopos/tenant-service/internal/domain"
)

const pgUniqueViolation = "23505"

// TenantStore is a Postgres-backed implementation of the tenant persistence layer.
type TenantStore struct {
	db *sql.DB
}

// New returns a TenantStore backed by the provided *sql.DB.
func New(db *sql.DB) *TenantStore {
	return &TenantStore{db: db}
}

// Create inserts a new tenant row. Returns ErrSlugTaken if the slug is already used.
func (s *TenantStore) Create(ctx context.Context, t *domain.Tenant) error {
	settingsJSON, err := marshalSettings(t.Settings)
	if err != nil {
		return err
	}

	const q = `
		INSERT INTO tenants (id, name, slug, plan, status, owner_email, settings, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err = s.db.ExecContext(ctx, q,
		t.ID, t.Name, t.Slug, string(t.Plan), string(t.Status),
		t.OwnerEmail, settingsJSON, t.CreatedAt, t.UpdatedAt,
	)
	if err != nil {
		return mapPQError(err)
	}
	return nil
}

// Get fetches a tenant by primary key ID.
func (s *TenantStore) Get(ctx context.Context, id string) (*domain.Tenant, error) {
	const q = `
		SELECT id, name, slug, plan, status, owner_email, settings, created_at, updated_at
		FROM tenants
		WHERE id = $1 AND status != 'deleted'`

	row := s.db.QueryRowContext(ctx, q, id)
	return scanTenant(row)
}

// GetBySlug fetches a tenant by its unique slug.
func (s *TenantStore) GetBySlug(ctx context.Context, slug string) (*domain.Tenant, error) {
	const q = `
		SELECT id, name, slug, plan, status, owner_email, settings, created_at, updated_at
		FROM tenants
		WHERE slug = $1 AND status != 'deleted'`

	row := s.db.QueryRowContext(ctx, q, slug)
	return scanTenant(row)
}

// List returns a page of tenants, optionally filtered by status.
// Pass an empty string to list all non-deleted tenants.
func (s *TenantStore) List(ctx context.Context, status string, limit, offset int) ([]*domain.Tenant, error) {
	var (
		rows *sql.Rows
		err  error
	)

	if limit <= 0 {
		limit = 20
	}

	if status != "" {
		const q = `
			SELECT id, name, slug, plan, status, owner_email, settings, created_at, updated_at
			FROM tenants
			WHERE status = $1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3`
		rows, err = s.db.QueryContext(ctx, q, status, limit, offset)
	} else {
		const q = `
			SELECT id, name, slug, plan, status, owner_email, settings, created_at, updated_at
			FROM tenants
			WHERE status != 'deleted'
			ORDER BY created_at DESC
			LIMIT $1 OFFSET $2`
		rows, err = s.db.QueryContext(ctx, q, limit, offset)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tenants []*domain.Tenant
	for rows.Next() {
		t, err := scanTenantRow(rows)
		if err != nil {
			return nil, err
		}
		tenants = append(tenants, t)
	}
	return tenants, rows.Err()
}

// Update applies partial updates to an existing tenant.
func (s *TenantStore) Update(ctx context.Context, id string, req domain.UpdateTenantRequest) (*domain.Tenant, error) {
	// Fetch current state first so we can merge settings.
	current, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		current.Name = *req.Name
	}
	if req.Plan != nil {
		current.Plan = *req.Plan
	}
	if req.Status != nil {
		current.Status = *req.Status
	}
	if req.Settings != nil {
		// Merge: incoming keys overwrite existing ones.
		for k, v := range req.Settings {
			current.Settings[k] = v
		}
	}
	current.UpdatedAt = time.Now().UTC()

	settingsJSON, err := marshalSettings(current.Settings)
	if err != nil {
		return nil, err
	}

	const q = `
		UPDATE tenants
		SET name = $1, plan = $2, status = $3, settings = $4, updated_at = $5
		WHERE id = $6`

	result, err := s.db.ExecContext(ctx, q,
		current.Name, string(current.Plan), string(current.Status),
		settingsJSON, current.UpdatedAt, id,
	)
	if err != nil {
		return nil, mapPQError(err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if affected == 0 {
		return nil, domain.ErrNotFound
	}
	return current, nil
}

// Delete soft-deletes a tenant by setting its status to "deleted".
func (s *TenantStore) Delete(ctx context.Context, id string) error {
	const q = `
		UPDATE tenants
		SET status = 'deleted', updated_at = $1
		WHERE id = $2 AND status != 'deleted'`

	result, err := s.db.ExecContext(ctx, q, time.Now().UTC(), id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// — helpers —

func scanTenant(row *sql.Row) (*domain.Tenant, error) {
	var (
		t            domain.Tenant
		plan, status string
		settingsRaw  []byte
	)
	err := row.Scan(&t.ID, &t.Name, &t.Slug, &plan, &status, &t.OwnerEmail, &settingsRaw, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	t.Plan = domain.Plan(plan)
	t.Status = domain.TenantStatus(status)
	if err := json.Unmarshal(settingsRaw, &t.Settings); err != nil {
		t.Settings = map[string]string{}
	}
	return &t, nil
}

func scanTenantRow(rows *sql.Rows) (*domain.Tenant, error) {
	var (
		t            domain.Tenant
		plan, status string
		settingsRaw  []byte
	)
	if err := rows.Scan(&t.ID, &t.Name, &t.Slug, &plan, &status, &t.OwnerEmail, &settingsRaw, &t.CreatedAt, &t.UpdatedAt); err != nil {
		return nil, err
	}
	t.Plan = domain.Plan(plan)
	t.Status = domain.TenantStatus(status)
	if err := json.Unmarshal(settingsRaw, &t.Settings); err != nil {
		t.Settings = map[string]string{}
	}
	return &t, nil
}

func marshalSettings(s map[string]string) ([]byte, error) {
	if s == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(s)
}

func mapPQError(err error) error {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == pgUniqueViolation {
		return domain.ErrSlugTaken
	}
	return err
}
