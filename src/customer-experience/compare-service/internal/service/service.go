package service

import (
	"context"
	"fmt"
	"time"

	"github.com/shopos/compare-service/internal/domain"
	"github.com/shopos/compare-service/internal/store"
)

// Servicer defines the business logic contract for compare list operations.
type Servicer interface {
	AddItem(ctx context.Context, customerID string, item domain.CompareItem) (*domain.CompareList, error)
	RemoveItem(ctx context.Context, customerID string, productID string) (*domain.CompareList, error)
	GetCompareList(ctx context.Context, customerID string) (*domain.CompareList, error)
	ClearList(ctx context.Context, customerID string) error
}

// CompareService implements Servicer.
type CompareService struct {
	store    store.Storer
	ttl      time.Duration
	maxItems int
}

// New returns a new CompareService.
func New(s store.Storer, ttl time.Duration, maxItems int) *CompareService {
	return &CompareService{
		store:    s,
		ttl:      ttl,
		maxItems: maxItems,
	}
}

// AddItem adds a product to the customer's compare list.
// Returns ErrListFull if the list already contains maxItems products.
func (svc *CompareService) AddItem(ctx context.Context, customerID string, item domain.CompareItem) (*domain.CompareList, error) {
	if item.ProductID == "" {
		return nil, fmt.Errorf("productId is required")
	}

	list, err := svc.store.GetList(ctx, customerID)
	if err != nil {
		return nil, err
	}

	// Idempotency: if the product is already in the list, return the list as-is.
	for _, existing := range list.Items {
		if existing.ProductID == item.ProductID {
			return list, nil
		}
	}

	if len(list.Items) >= svc.maxItems {
		return nil, domain.ErrListFull
	}

	list.Items = append(list.Items, item)
	list.UpdatedAt = time.Now().UTC()

	if err := svc.store.SaveList(ctx, customerID, list, svc.ttl); err != nil {
		return nil, err
	}
	return list, nil
}

// RemoveItem removes a product from the customer's compare list.
// Returns ErrItemNotFound if the product is not in the list.
func (svc *CompareService) RemoveItem(ctx context.Context, customerID string, productID string) (*domain.CompareList, error) {
	list, err := svc.store.GetList(ctx, customerID)
	if err != nil {
		return nil, err
	}

	found := false
	filtered := make([]domain.CompareItem, 0, len(list.Items))
	for _, item := range list.Items {
		if item.ProductID == productID {
			found = true
			continue
		}
		filtered = append(filtered, item)
	}

	if !found {
		return nil, domain.ErrItemNotFound
	}

	list.Items = filtered
	list.UpdatedAt = time.Now().UTC()

	if err := svc.store.SaveList(ctx, customerID, list, svc.ttl); err != nil {
		return nil, err
	}
	return list, nil
}

// GetCompareList returns the full compare list for a customer.
func (svc *CompareService) GetCompareList(ctx context.Context, customerID string) (*domain.CompareList, error) {
	return svc.store.GetList(ctx, customerID)
}

// ClearList removes all items from the customer's compare list.
func (svc *CompareService) ClearList(ctx context.Context, customerID string) error {
	return svc.store.ClearList(ctx, customerID)
}
