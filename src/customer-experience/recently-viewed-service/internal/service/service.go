package service

import (
	"context"
	"fmt"
	"time"

	"github.com/shopos/recently-viewed-service/internal/domain"
	"github.com/shopos/recently-viewed-service/internal/store"
)

// Servicer defines the business logic contract for recently-viewed operations.
type Servicer interface {
	RecordView(ctx context.Context, customerID string, item domain.ViewedItem) error
	GetRecentlyViewed(ctx context.Context, customerID string, limit int) (*domain.RecentlyViewedList, error)
	ClearHistory(ctx context.Context, customerID string) error
	GetCount(ctx context.Context, customerID string) (int, error)
}

// RecentlyViewedService implements Servicer.
type RecentlyViewedService struct {
	store    store.Storer
	maxItems int
	itemTTL  time.Duration
}

// New returns a new RecentlyViewedService.
func New(s store.Storer, maxItems int, itemTTL time.Duration) *RecentlyViewedService {
	return &RecentlyViewedService{
		store:    s,
		maxItems: maxItems,
		itemTTL:  itemTTL,
	}
}

// RecordView records that a customer viewed a product.
// If ViewedAt is zero it is set to the current time.
func (svc *RecentlyViewedService) RecordView(ctx context.Context, customerID string, item domain.ViewedItem) error {
	if customerID == "" {
		return fmt.Errorf("customerID is required")
	}
	if item.ProductID == "" {
		return fmt.Errorf("productId is required")
	}
	if item.ViewedAt.IsZero() {
		item.ViewedAt = time.Now().UTC()
	}
	return svc.store.RecordView(ctx, customerID, item, svc.maxItems, svc.itemTTL)
}

// GetRecentlyViewed returns the most recently viewed products for a customer.
func (svc *RecentlyViewedService) GetRecentlyViewed(ctx context.Context, customerID string, limit int) (*domain.RecentlyViewedList, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > svc.maxItems {
		limit = svc.maxItems
	}

	items, err := svc.store.GetRecent(ctx, customerID, limit)
	if err != nil {
		return nil, err
	}

	total, err := svc.store.GetCount(ctx, customerID)
	if err != nil {
		return nil, err
	}

	if items == nil {
		items = []domain.ViewedItem{}
	}

	return &domain.RecentlyViewedList{
		CustomerID: customerID,
		Items:      items,
		Total:      total,
	}, nil
}

// ClearHistory removes all recently-viewed history for a customer.
func (svc *RecentlyViewedService) ClearHistory(ctx context.Context, customerID string) error {
	return svc.store.ClearHistory(ctx, customerID)
}

// GetCount returns the number of items in the customer's recently-viewed list.
func (svc *RecentlyViewedService) GetCount(ctx context.Context, customerID string) (int, error) {
	return svc.store.GetCount(ctx, customerID)
}
