package service

import (
	"context"
	"github.com/shopos/partner-bff/internal/domain"
)

type CatalogService interface {
	ListProducts(ctx context.Context, page, pageSize int, categoryID string) (*domain.ProductList, error)
	GetProduct(ctx context.Context, id string) (*domain.Product, error)
	ListCategories(ctx context.Context) ([]*domain.Category, error)
}

type GRPCCatalogService struct{ addr string }

func NewGRPCCatalogService(addr string) CatalogService { return &GRPCCatalogService{addr: addr} }

func (s *GRPCCatalogService) ListProducts(_ context.Context, _, _ int, _ string) (*domain.ProductList, error) {
	return nil, ErrNotImplemented
}
func (s *GRPCCatalogService) GetProduct(_ context.Context, _ string) (*domain.Product, error) {
	return nil, ErrNotImplemented
}
func (s *GRPCCatalogService) ListCategories(_ context.Context) ([]*domain.Category, error) {
	return nil, ErrNotImplemented
}
