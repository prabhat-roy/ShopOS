package domain

import (
	"errors"
	"time"
)

// BundleItem is a single product entry within a bundle.
type BundleItem struct {
	ProductID string  `json:"product_id"`
	SKU       string  `json:"sku"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"` // item price within bundle
}

// Bundle is a curated collection of products sold together at a bundle price.
type Bundle struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Price       float64      `json:"price"`    // bundle price (may differ from sum of parts)
	Currency    string       `json:"currency"`
	Items       []BundleItem `json:"items"`
	Active      bool         `json:"active"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

var ErrNotFound = errors.New("not found")
