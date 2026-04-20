package domain

import (
	"errors"
	"time"
)

// Category represents a product category in the catalog.
type Category struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	ParentID    *string   `json:"parent_id,omitempty"` // nil = root
	Description string    `json:"description"`
	ImageURL    string    `json:"image_url"`
	SortOrder   int       `json:"sort_order"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

var ErrNotFound = errors.New("not found")
var ErrSlugTaken = errors.New("slug already taken")
