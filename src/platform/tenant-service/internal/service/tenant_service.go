package service

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/shopos/tenant-service/internal/domain"
)

// slugRe validates that a slug is 3-63 lowercase alphanumeric characters or hyphens,
// does not start or end with a hyphen, and contains no consecutive hyphens.
var slugRe = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{1,61}[a-z0-9])?$`)

// Storer is the persistence interface required by TenantService.
type Storer interface {
	Create(ctx context.Context, t *domain.Tenant) error
	Get(ctx context.Context, id string) (*domain.Tenant, error)
	GetBySlug(ctx context.Context, slug string) (*domain.Tenant, error)
	List(ctx context.Context, status string, limit, offset int) ([]*domain.Tenant, error)
	Update(ctx context.Context, id string, req domain.UpdateTenantRequest) (*domain.Tenant, error)
	Delete(ctx context.Context, id string) error
}

// TenantService implements business logic for the tenant domain.
type TenantService struct {
	store Storer
}

// New returns a TenantService wired to the provided Storer.
func New(store Storer) *TenantService {
	return &TenantService{store: store}
}

// Create validates the request and creates a new tenant.
func (s *TenantService) Create(ctx context.Context, req domain.CreateTenantRequest) (*domain.Tenant, error) {
	if err := validateSlug(req.Slug); err != nil {
		return nil, err
	}
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if req.OwnerEmail == "" {
		return nil, fmt.Errorf("owner_email is required")
	}

	plan := req.Plan
	if plan == "" {
		plan = domain.PlanStarter
	}

	now := time.Now().UTC()
	t := &domain.Tenant{
		ID:         uuid.NewString(),
		Name:       req.Name,
		Slug:       req.Slug,
		OwnerEmail: req.OwnerEmail,
		Plan:       plan,
		Status:     domain.StatusActive,
		Settings:   map[string]string{},
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := s.store.Create(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

// Get fetches a tenant by its ID.
func (s *TenantService) Get(ctx context.Context, id string) (*domain.Tenant, error) {
	return s.store.Get(ctx, id)
}

// GetBySlug fetches a tenant by its unique slug.
func (s *TenantService) GetBySlug(ctx context.Context, slug string) (*domain.Tenant, error) {
	return s.store.GetBySlug(ctx, slug)
}

// List returns a paginated list of tenants, optionally filtered by status.
func (s *TenantService) List(ctx context.Context, status string, limit, offset int) ([]*domain.Tenant, error) {
	return s.store.List(ctx, status, limit, offset)
}

// Update applies partial updates to the tenant identified by id.
func (s *TenantService) Update(ctx context.Context, id string, req domain.UpdateTenantRequest) (*domain.Tenant, error) {
	return s.store.Update(ctx, id, req)
}

// Delete soft-deletes the tenant identified by id.
func (s *TenantService) Delete(ctx context.Context, id string) error {
	return s.store.Delete(ctx, id)
}

// GetSettings returns the settings map for a tenant.
func (s *TenantService) GetSettings(ctx context.Context, id string) (map[string]string, error) {
	t, err := s.store.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return t.Settings, nil
}

// UpdateSettings replaces/merges the settings for a tenant.
func (s *TenantService) UpdateSettings(ctx context.Context, id string, settings map[string]string) (*domain.Tenant, error) {
	req := domain.UpdateTenantRequest{Settings: settings}
	return s.store.Update(ctx, id, req)
}

// — helpers —

func validateSlug(slug string) error {
	if len(slug) < 3 || len(slug) > 63 {
		return fmt.Errorf("slug must be between 3 and 63 characters")
	}
	if !slugRe.MatchString(slug) {
		return fmt.Errorf("slug must contain only lowercase alphanumeric characters and hyphens, and must not start or end with a hyphen")
	}
	return nil
}
