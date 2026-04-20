package domain

import (
	"errors"
	"time"
)

// Brand represents a product brand in the catalog.
type Brand struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	LogoURL     string    `json:"logo_url"`
	Website     string    `json:"website"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

var ErrNotFound = errors.New("not found")
var ErrSlugTaken = errors.New("slug already taken")
