package service

import (
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/shopos/brand-service/domain"
)

// slugRE validates lowercase alphanumeric + hyphens, no leading/trailing hyphens.
var slugRE = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// Storer is the persistence interface required by the service.
type Storer interface {
	Create(b *domain.Brand) error
	GetByID(id string) (*domain.Brand, error)
	GetBySlug(slug string) (*domain.Brand, error)
	List(activeOnly bool) ([]*domain.Brand, error)
	Update(b *domain.Brand) error
	Delete(id string) error
}

// Service contains the business logic for brands.
type Service struct {
	store Storer
}

// New returns a new Service backed by the given Storer.
func New(s Storer) *Service {
	return &Service{store: s}
}

// CreateRequest carries the fields required to create a new brand.
type CreateRequest struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	LogoURL     string `json:"logo_url"`
	Website     string `json:"website"`
}

// UpdateRequest carries the mutable fields for a PATCH operation.
// A nil pointer means "leave unchanged".
type UpdateRequest struct {
	Name        *string `json:"name,omitempty"`
	Slug        *string `json:"slug,omitempty"`
	Description *string `json:"description,omitempty"`
	LogoURL     *string `json:"logo_url,omitempty"`
	Website     *string `json:"website,omitempty"`
	Active      *bool   `json:"active,omitempty"`
}

// Create validates inputs and persists a new brand, returning the created entity.
func (s *Service) Create(req CreateRequest) (*domain.Brand, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if !slugRE.MatchString(req.Slug) {
		return nil, fmt.Errorf("slug must be lowercase alphanumeric with hyphens (e.g. my-brand)")
	}
	now := time.Now().UTC()
	b := &domain.Brand{
		ID:          uuid.NewString(),
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		LogoURL:     req.LogoURL,
		Website:     req.Website,
		Active:      true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.store.Create(b); err != nil {
		return nil, err
	}
	return b, nil
}

// GetByID retrieves a single brand by its ID.
func (s *Service) GetByID(id string) (*domain.Brand, error) {
	return s.store.GetByID(id)
}

// GetBySlug retrieves a single brand by its URL slug.
func (s *Service) GetBySlug(slug string) (*domain.Brand, error) {
	return s.store.GetBySlug(slug)
}

// List returns all brands, optionally filtered by active state.
func (s *Service) List(activeOnly bool) ([]*domain.Brand, error) {
	return s.store.List(activeOnly)
}

// Update applies partial changes to an existing brand.
func (s *Service) Update(id string, req UpdateRequest) (*domain.Brand, error) {
	b, err := s.store.GetByID(id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		if *req.Name == "" {
			return nil, fmt.Errorf("name cannot be empty")
		}
		b.Name = *req.Name
	}
	if req.Slug != nil {
		if !slugRE.MatchString(*req.Slug) {
			return nil, fmt.Errorf("slug must be lowercase alphanumeric with hyphens")
		}
		b.Slug = *req.Slug
	}
	if req.Description != nil {
		b.Description = *req.Description
	}
	if req.LogoURL != nil {
		b.LogoURL = *req.LogoURL
	}
	if req.Website != nil {
		b.Website = *req.Website
	}
	if req.Active != nil {
		b.Active = *req.Active
	}
	b.UpdatedAt = time.Now().UTC()

	if err := s.store.Update(b); err != nil {
		return nil, err
	}
	return b, nil
}

// Delete soft-deletes a brand by its ID.
func (s *Service) Delete(id string) error {
	return s.store.Delete(id)
}
