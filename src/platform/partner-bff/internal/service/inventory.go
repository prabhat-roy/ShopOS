package service

import (
	"context"
	"github.com/shopos/partner-bff/internal/domain"
)

type InventoryService interface {
	GetStock(ctx context.Context, productID string) (*domain.StockLevel, error)
	GetBulkStock(ctx context.Context, productIDs []string) ([]*domain.StockLevel, error)
}

type GRPCInventoryService struct{ addr string }

func NewGRPCInventoryService(addr string) InventoryService { return &GRPCInventoryService{addr: addr} }

func (s *GRPCInventoryService) GetStock(_ context.Context, _ string) (*domain.StockLevel, error) {
	return nil, ErrNotImplemented
}
func (s *GRPCInventoryService) GetBulkStock(_ context.Context, _ []string) ([]*domain.StockLevel, error) {
	return nil, ErrNotImplemented
}
