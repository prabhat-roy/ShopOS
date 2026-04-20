package service

import (
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/shopos/category-service/domain"
)

// slugRE validates lowercase alphanumeric + hyphens, no leading/trailing hyphens.
var slugRE = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// Storer is the persistence interface required by the service.
type Storer interface {
	Create(c *domain.Category) error
	GetByID(id string) (*domain.Category, error)
	GetBySlug(slug string) (*domain.Category, error)
	List(parentID *string, activeOnly bool) ([]*domain.Category, error)
	Update(c *domain.Category) error
	Delete(id string) error
}

// Service contains the business logic for categories.
type Service struct {
	store Storer
}

// New returns a new Service backed by the given Storer.
func New(s Storer) *Service {
	return &Service{store: s}
}

// CreateRequest carries the fields required to create a new category.
type CreateRequest struct {
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	ParentID    *string `json:"parent_id,omitempty"`
	Description string  `json:"description"`
	ImageURL    string  `json:"image_url"`
	SortOrder   int     `json:"sort_order"`
}

// UpdateRequest carries the mutable fields for a PATCH operation.
// A nil pointer means "leave unchanged".
type UpdateRequest struct {
	Name        *string `json:"name,omitempty"`
	Slug        *string `json:"slug,omitempty"`
	ParentID    *string `json:"parent_id,omitempty"`
	Description *string `json:"description,omitempty"`
	ImageURL    *string `json:"image_url,omitempty"`
	SortOrder   *int    `json:"sort_order,omitempty"`
	Active      *bool   `json:"active,omitempty"`
}

// Create validates inputs and persists a new category, returning the created entity.
func (s *Service) Create(req CreateRequest) (*domain.Category, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if !slugRE.MatchString(req.Slug) {
		return nil, fmt.Errorf("slug must be lowercase alphanumeric with hyphens (e.g. my-category)")
	}
	now := time.Now().UTC()
	c := &domain.Category{
		ID:          uuid.NewString(),
		Name:        req.Name,
		Slug:        req.Slug,
		ParentID:    req.ParentID,
		Description: req.Description,
		ImageURL:    req.ImageURL,
		SortOrder:   req.SortOrder,
		Active:      true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.store.Create(c); err != nil {
		return nil, err
	}
	return c, nil
}

// GetByID retrieves a single category by its ID.
func (s *Service) GetByID(id string) (*domain.Category, error) {
	return s.store.GetByID(id)
}

// GetBySlug retrieves a single category by its URL slug.
func (s *Service) GetBySlug(slug string) (*domain.Category, error) {
	return s.store.GetBySlug(slug)
}

// List returns all categories, optionally filtered by parent and/or active state.
func (s *Service) List(parentID *string, activeOnly bool) ([]*domain.Category, error) {
	return s.store.List(parentID, activeOnly)
}

// Update applies partial changes to an existing category.
func (s *Service) Update(id string, req UpdateRequest) (*domain.Category, error) {
	c, err := s.store.GetByID(id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		if *req.Name == "" {
			return nil, fmt.Errorf("name cannot be empty")
		}
		c.Name = *req.Name
	}
	if req.Slug != nil {
		if !slugRE.MatchString(*req.Slug) {
			return nil, fmt.Errorf("slug must be lowercase alphanumeric with hyphens")
		}
		c.Slug = *req.Slug
	}
	if req.ParentID != nil {
		c.ParentID = req.ParentID
	}
	if req.Description != nil {
		c.Description = *req.Description
	}
	if req.ImageURL != nil {
		c.ImageURL = *req.ImageURL
	}
	if req.SortOrder != nil {
		c.SortOrder = *req.SortOrder
	}
	if req.Active != nil {
		c.Active = *req.Active
	}
	c.UpdatedAt = time.Now().UTC()

	if err := s.store.Update(c); err != nil {
		return nil, err
	}
	return c, nil
}

// Delete soft-deletes a category by its ID.
func (s *Service) Delete(id string) error {
	return s.store.Delete(id)
}
