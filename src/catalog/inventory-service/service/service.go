package service

import (
	"context"

	"github.com/shopos/inventory-service/domain"
)

// Storer is the persistence contract required by the service layer.
type Storer interface {
	GetStock(ctx context.Context, productID, warehouseID string) (*domain.StockLevel, error)
	ListStock(ctx context.Context, productID string) ([]*domain.StockLevel, error)
	UpsertStock(ctx context.Context, productID, sku, warehouseID string, available, reorder int) (*domain.StockLevel, error)
	Reserve(ctx context.Context, orderID, productID string, qty int) (*domain.Reservation, error)
	Release(ctx context.Context, reservationID string) error
	Commit(ctx context.Context, reservationID string) error
	GetReservation(ctx context.Context, id string) (*domain.Reservation, error)
}

// Service implements inventory business logic.
type Service struct {
	store Storer
}

// New returns a Service backed by the given Storer.
func New(s Storer) *Service {
	return &Service{store: s}
}

// GetStock proxies to the store.
func (s *Service) GetStock(ctx context.Context, productID, warehouseID string) (*domain.StockLevel, error) {
	return s.store.GetStock(ctx, productID, warehouseID)
}

// ListStock proxies to the store.
func (s *Service) ListStock(ctx context.Context, productID string) ([]*domain.StockLevel, error) {
	return s.store.ListStock(ctx, productID)
}

// UpsertStock proxies to the store.
func (s *Service) UpsertStock(ctx context.Context, productID, sku, warehouseID string, available, reorder int) (*domain.StockLevel, error) {
	return s.store.UpsertStock(ctx, productID, sku, warehouseID, available, reorder)
}

// Reserve validates and delegates reservation creation to the store.
func (s *Service) Reserve(ctx context.Context, orderID, productID string, qty int) (*domain.Reservation, error) {
	if qty <= 0 {
		return nil, domain.ErrInsufficientStock
	}
	return s.store.Reserve(ctx, orderID, productID, qty)
}

// Release delegates reservation release to the store.
func (s *Service) Release(ctx context.Context, reservationID string) error {
	return s.store.Release(ctx, reservationID)
}

// Commit delegates reservation commit to the store.
func (s *Service) Commit(ctx context.Context, reservationID string) error {
	return s.store.Commit(ctx, reservationID)
}

// GetReservation proxies to the store.
func (s *Service) GetReservation(ctx context.Context, id string) (*domain.Reservation, error) {
	return s.store.GetReservation(ctx, id)
}
