package service

import (
	"context"

	"github.com/shopos/bundle-service/domain"
)

// Storer is the persistence contract required by the service layer.
type Storer interface {
	Create(ctx context.Context, b *domain.Bundle) (*domain.Bundle, error)
	GetByID(ctx context.Context, id string) (*domain.Bundle, error)
	List(ctx context.Context, activeOnly bool) ([]*domain.Bundle, error)
	Update(ctx context.Context, id string, patch *domain.Bundle) (*domain.Bundle, error)
	Delete(ctx context.Context, id string) error
}

// Service implements bundle business logic.
type Service struct {
	store Storer
}

// New returns a Service backed by the given Storer.
func New(s Storer) *Service {
	return &Service{store: s}
}

// Create persists a new bundle.
func (s *Service) Create(ctx context.Context, b *domain.Bundle) (*domain.Bundle, error) {
	return s.store.Create(ctx, b)
}

// GetByID returns a bundle by ID.
func (s *Service) GetByID(ctx context.Context, id string) (*domain.Bundle, error) {
	return s.store.GetByID(ctx, id)
}

// List returns bundles, optionally filtered to active-only.
func (s *Service) List(ctx context.Context, activeOnly bool) ([]*domain.Bundle, error) {
	return s.store.List(ctx, activeOnly)
}

// Update applies a partial update to an existing bundle.
func (s *Service) Update(ctx context.Context, id string, patch *domain.Bundle) (*domain.Bundle, error) {
	return s.store.Update(ctx, id, patch)
}

// Delete soft-deletes a bundle.
func (s *Service) Delete(ctx context.Context, id string) error {
	return s.store.Delete(ctx, id)
}
