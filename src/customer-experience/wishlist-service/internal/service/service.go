package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopos/wishlist-service/internal/domain"
	"github.com/shopos/wishlist-service/internal/store"
)

// Servicer defines the business logic contract for wishlist operations.
type Servicer interface {
	AddToWishlist(ctx context.Context, customerID uuid.UUID, req *domain.AddItemRequest) (*domain.WishlistItem, error)
	GetWishlistItem(ctx context.Context, customerID uuid.UUID, productID string) (*domain.WishlistItem, error)
	RemoveFromWishlist(ctx context.Context, customerID uuid.UUID, productID string) error
	GetWishlist(ctx context.Context, customerID uuid.UUID, limit, offset int) (*domain.WishlistPage, error)
	ClearWishlist(ctx context.Context, customerID uuid.UUID) error
	CheckWishlist(ctx context.Context, customerID uuid.UUID, productID string) (bool, error)
}

// WishlistService implements Servicer.
type WishlistService struct {
	store store.Storer
}

// New returns a new WishlistService backed by the provided Storer.
func New(s store.Storer) *WishlistService {
	return &WishlistService{store: s}
}

// AddToWishlist adds a product to the customer's wishlist.
func (svc *WishlistService) AddToWishlist(ctx context.Context, customerID uuid.UUID, req *domain.AddItemRequest) (*domain.WishlistItem, error) {
	if req.ProductID == "" {
		return nil, fmt.Errorf("productId is required")
	}
	if req.ProductName == "" {
		return nil, fmt.Errorf("productName is required")
	}

	item := &domain.WishlistItem{
		ID:          uuid.New(),
		CustomerID:  customerID,
		ProductID:   req.ProductID,
		ProductName: req.ProductName,
		Price:       req.Price,
		ImageURL:    req.ImageURL,
		AddedAt:     time.Now().UTC(),
	}

	if err := svc.store.AddItem(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

// GetWishlistItem retrieves a single item from the customer's wishlist.
func (svc *WishlistService) GetWishlistItem(ctx context.Context, customerID uuid.UUID, productID string) (*domain.WishlistItem, error) {
	return svc.store.GetItem(ctx, customerID, productID)
}

// RemoveFromWishlist removes a product from the customer's wishlist.
func (svc *WishlistService) RemoveFromWishlist(ctx context.Context, customerID uuid.UUID, productID string) error {
	return svc.store.RemoveItem(ctx, customerID, productID)
}

// GetWishlist returns a paginated list of wishlist items for a customer.
func (svc *WishlistService) GetWishlist(ctx context.Context, customerID uuid.UUID, limit, offset int) (*domain.WishlistPage, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	items, total, err := svc.store.ListItems(ctx, customerID, limit, offset)
	if err != nil {
		return nil, err
	}

	return &domain.WishlistPage{
		CustomerID: customerID,
		Items:      items,
		Total:      total,
		Limit:      limit,
		Offset:     offset,
	}, nil
}

// ClearWishlist removes all items from the customer's wishlist.
func (svc *WishlistService) ClearWishlist(ctx context.Context, customerID uuid.UUID) error {
	return svc.store.ClearWishlist(ctx, customerID)
}

// CheckWishlist returns whether a product is in the customer's wishlist.
func (svc *WishlistService) CheckWishlist(ctx context.Context, customerID uuid.UUID, productID string) (bool, error) {
	return svc.store.IsInWishlist(ctx, customerID, productID)
}
