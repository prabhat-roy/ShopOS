package service

import (
	"context"
	"github.com/shopos/partner-bff/internal/domain"
)

type OrderService interface {
	ListOrders(ctx context.Context, partnerID string, page, pageSize int) ([]*domain.Order, error)
	GetOrder(ctx context.Context, orderID, partnerID string) (*domain.Order, error)
	PlaceOrder(ctx context.Context, partnerID string, req *domain.PlaceOrderRequest) (*domain.Order, error)
}

type GRPCOrderService struct{ addr string }

func NewGRPCOrderService(addr string) OrderService { return &GRPCOrderService{addr: addr} }

func (s *GRPCOrderService) ListOrders(_ context.Context, _ string, _, _ int) ([]*domain.Order, error) {
	return nil, ErrNotImplemented
}
func (s *GRPCOrderService) GetOrder(_ context.Context, _, _ string) (*domain.Order, error) {
	return nil, ErrNotImplemented
}
func (s *GRPCOrderService) PlaceOrder(_ context.Context, _ string, _ *domain.PlaceOrderRequest) (*domain.Order, error) {
	return nil, ErrNotImplemented
}
