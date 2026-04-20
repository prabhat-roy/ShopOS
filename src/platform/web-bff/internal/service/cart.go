package service

import (
	"context"

	"github.com/shopos/web-bff/internal/domain"
)

type CartService interface {
	GetCart(ctx context.Context, userID string) (*domain.Cart, error)
	AddItem(ctx context.Context, userID, productID string, qty int) (*domain.Cart, error)
	UpdateItem(ctx context.Context, userID, itemID string, qty int) (*domain.Cart, error)
	RemoveItem(ctx context.Context, userID, itemID string) (*domain.Cart, error)
	ClearCart(ctx context.Context, userID string) error
}

// GRPCCartService connects to commerce/cart-service (port 50080).
// TODO: Replace stub bodies with generated proto client calls once proto/commerce/ is compiled.
type GRPCCartService struct{ addr string }

func NewGRPCCartService(addr string) CartService {
	return &GRPCCartService{addr: addr}
}

func (s *GRPCCartService) GetCart(_ context.Context, _ string) (*domain.Cart, error) {
	return nil, ErrNotImplemented
}

func (s *GRPCCartService) AddItem(_ context.Context, _, _ string, _ int) (*domain.Cart, error) {
	return nil, ErrNotImplemented
}

func (s *GRPCCartService) UpdateItem(_ context.Context, _, _ string, _ int) (*domain.Cart, error) {
	return nil, ErrNotImplemented
}

func (s *GRPCCartService) RemoveItem(_ context.Context, _, _ string) (*domain.Cart, error) {
	return nil, ErrNotImplemented
}

func (s *GRPCCartService) ClearCart(_ context.Context, _ string) error {
	return ErrNotImplemented
}
