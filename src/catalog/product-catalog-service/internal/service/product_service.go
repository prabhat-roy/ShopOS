package service

import (
	"context"

	"github.com/shopos/product-catalog-service/internal/domain"
)

// Storer defines the persistence operations the service depends on.
// Implemented by store.ProductStore; can be mocked in tests.
type Storer interface {
	Create(ctx context.Context, req *domain.CreateProductRequest) (*domain.Product, error)
	GetByID(ctx context.Context, id string) (*domain.Product, error)
	GetBySKU(ctx context.Context, sku string) (*domain.Product, error)
	List(ctx context.Context, req *domain.ListProductsRequest) (*domain.ProductList, error)
	Update(ctx context.Context, id string, req *domain.UpdateProductRequest) (*domain.Product, error)
	Delete(ctx context.Context, id string) error
}

// ProductService contains the business logic for the product catalog.
type ProductService struct {
	store Storer
}

// New creates a ProductService backed by the provided Storer.
func New(store Storer) *ProductService {
	return &ProductService{store: store}
}

// CreateProduct validates the request and persists a new product.
func (s *ProductService) CreateProduct(ctx context.Context, req *domain.CreateProductRequest) (*domain.Product, error) {
	if req.SKU == "" {
		return nil, &ValidationError{Field: "sku", Message: "SKU is required"}
	}
	if req.Name == "" {
		return nil, &ValidationError{Field: "name", Message: "name is required"}
	}
	if req.Price < 0 {
		return nil, &ValidationError{Field: "price", Message: "price must be non-negative"}
	}
	if req.Currency == "" {
		req.Currency = "USD"
	}

	return s.store.Create(ctx, req)
}

// GetProduct retrieves a product by its ID.
func (s *ProductService) GetProduct(ctx context.Context, id string) (*domain.Product, error) {
	if id == "" {
		return nil, &ValidationError{Field: "id", Message: "id is required"}
	}
	return s.store.GetByID(ctx, id)
}

// GetBySKU retrieves a product by its SKU.
func (s *ProductService) GetBySKU(ctx context.Context, sku string) (*domain.Product, error) {
	if sku == "" {
		return nil, &ValidationError{Field: "sku", Message: "sku is required"}
	}
	return s.store.GetBySKU(ctx, sku)
}

// ListProducts returns a filtered, paginated list of products.
func (s *ProductService) ListProducts(ctx context.Context, req *domain.ListProductsRequest) (*domain.ProductList, error) {
	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}
	if req.Offset < 0 {
		req.Offset = 0
	}
	return s.store.List(ctx, req)
}

// UpdateProduct applies a partial update to the product identified by id.
func (s *ProductService) UpdateProduct(ctx context.Context, id string, req *domain.UpdateProductRequest) (*domain.Product, error) {
	if id == "" {
		return nil, &ValidationError{Field: "id", Message: "id is required"}
	}
	if req.Price != nil && *req.Price < 0 {
		return nil, &ValidationError{Field: "price", Message: "price must be non-negative"}
	}
	return s.store.Update(ctx, id, req)
}

// DeleteProduct soft-deletes the product identified by id.
func (s *ProductService) DeleteProduct(ctx context.Context, id string) error {
	if id == "" {
		return &ValidationError{Field: "id", Message: "id is required"}
	}
	return s.store.Delete(ctx, id)
}

// ---------------------------------------------------------------------------
// ValidationError
// ---------------------------------------------------------------------------

// ValidationError is returned when a request fails domain validation.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}
