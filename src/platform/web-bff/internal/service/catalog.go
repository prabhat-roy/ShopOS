package service

import (
	"context"

	"github.com/shopos/web-bff/internal/domain"
)

type CatalogService interface {
	ListProducts(ctx context.Context, req *domain.ListProductsRequest) (*domain.ProductList, error)
	GetProduct(ctx context.Context, id string) (*domain.Product, error)
	ListCategories(ctx context.Context) ([]*domain.Category, error)
	Search(ctx context.Context, query string, page, pageSize int) (*domain.ProductList, error)
}

// GRPCCatalogService connects to catalog/product-catalog-service (port 50070).
// TODO: Replace stub bodies with generated proto client calls once proto/catalog/ is compiled.
type GRPCCatalogService struct{ addr string }

func NewGRPCCatalogService(addr string) CatalogService {
	return &GRPCCatalogService{addr: addr}
}

func (s *GRPCCatalogService) ListProducts(_ context.Context, _ *domain.ListProductsRequest) (*domain.ProductList, error) {
	return nil, ErrNotImplemented
}

func (s *GRPCCatalogService) GetProduct(_ context.Context, _ string) (*domain.Product, error) {
	return nil, ErrNotImplemented
}

func (s *GRPCCatalogService) ListCategories(_ context.Context) ([]*domain.Category, error) {
	return nil, ErrNotImplemented
}

func (s *GRPCCatalogService) Search(_ context.Context, _ string, _, _ int) (*domain.ProductList, error) {
	return nil, ErrNotImplemented
}
