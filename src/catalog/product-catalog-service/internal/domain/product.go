package domain

import (
	"errors"
	"time"
)

// ProductStatus represents the lifecycle state of a product.
type ProductStatus string

const (
	StatusActive   ProductStatus = "active"
	StatusDraft    ProductStatus = "draft"
	StatusArchived ProductStatus = "archived"
)

// Product is the core domain entity stored in MongoDB.
type Product struct {
	ID          string            `json:"id"          bson:"_id"`
	SKU         string            `json:"sku"         bson:"sku"`
	Name        string            `json:"name"        bson:"name"`
	Description string            `json:"description" bson:"description"`
	CategoryID  string            `json:"category_id" bson:"category_id"`
	BrandID     string            `json:"brand_id"    bson:"brand_id"`
	Price       float64           `json:"price"       bson:"price"`
	Currency    string            `json:"currency"    bson:"currency"`
	Status      ProductStatus     `json:"status"      bson:"status"`
	Tags        []string          `json:"tags"        bson:"tags"`
	Attributes  map[string]string `json:"attributes"  bson:"attributes"`
	ImageURLs   []string          `json:"image_urls"  bson:"image_urls"`
	Weight      float64           `json:"weight_kg"   bson:"weight_kg"`
	CreatedAt   time.Time         `json:"created_at"  bson:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"  bson:"updated_at"`
}

// CreateProductRequest carries the fields needed to create a new product.
type CreateProductRequest struct {
	SKU         string            `json:"sku"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	CategoryID  string            `json:"category_id"`
	BrandID     string            `json:"brand_id"`
	Price       float64           `json:"price"`
	Currency    string            `json:"currency"`
	Tags        []string          `json:"tags"`
	Attributes  map[string]string `json:"attributes"`
	ImageURLs   []string          `json:"image_urls"`
	Weight      float64           `json:"weight_kg"`
}

// UpdateProductRequest carries the fields that may be patched on an existing product.
// Pointer fields are optional — nil means "leave unchanged".
type UpdateProductRequest struct {
	Name        *string           `json:"name"`
	Description *string           `json:"description"`
	Price       *float64          `json:"price"`
	Status      *ProductStatus    `json:"status"`
	Tags        []string          `json:"tags"`
	Attributes  map[string]string `json:"attributes"`
}

// ListProductsRequest carries filter and pagination parameters for list queries.
type ListProductsRequest struct {
	CategoryID string
	BrandID    string
	Status     ProductStatus
	MinPrice   float64
	MaxPrice   float64
	Limit      int
	Offset     int
	Tags       []string
}

// ProductList is the paginated response for list queries.
type ProductList struct {
	Items  []*Product `json:"items"`
	Total  int64      `json:"total"`
	Limit  int        `json:"limit"`
	Offset int        `json:"offset"`
}

// Sentinel errors returned by the store and surfaced by the service.
var (
	ErrNotFound     = errors.New("product not found")
	ErrDuplicateSKU = errors.New("SKU already exists")
)
