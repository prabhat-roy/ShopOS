package service

import (
	"context"

	"github.com/shopos/web-bff/internal/domain"
)

type OrderService interface {
	ListOrders(ctx context.Context, userID string, page, pageSize int) ([]*domain.Order, error)
	GetOrder(ctx context.Context, orderID, userID string) (*domain.Order, error)
	PlaceOrder(ctx context.Context, userID string, req *domain.PlaceOrderRequest) (*domain.Order, error)
}

// GRPCOrderService connects to commerce/order-service (port 50082).
// TODO: Replace stub bodies with generated proto client calls once proto/commerce/ is compiled.
type GRPCOrderService struct{ addr string }

func NewGRPCOrderService(addr string) OrderService {
	return &GRPCOrderService{addr: addr}
}

func (s *GRPCOrderService) ListOrders(_ context.Context, _ string, _, _ int) ([]*domain.Order, error) {
	return nil, ErrNotImplemented
}

func (s *GRPCOrderService) GetOrder(_ context.Context, _, _ string) (*domain.Order, error) {
	return nil, ErrNotImplemented
}

func (s *GRPCOrderService) PlaceOrder(_ context.Context, _ string, _ *domain.PlaceOrderRequest) (*domain.Order, error) {
	return nil, ErrNotImplemented
}
