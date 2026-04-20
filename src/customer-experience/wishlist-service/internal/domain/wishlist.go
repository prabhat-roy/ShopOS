package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// ErrNotFound is returned when a wishlist item does not exist.
var ErrNotFound = errors.New("wishlist item not found")

// ErrAlreadyExists is returned when a product is already in the customer's wishlist.
var ErrAlreadyExists = errors.New("item already exists in wishlist")

// WishlistItem represents a single product saved in a customer's wishlist.
type WishlistItem struct {
	ID          uuid.UUID `json:"id"`
	CustomerID  uuid.UUID `json:"customerId"`
	ProductID   string    `json:"productId"`
	ProductName string    `json:"productName"`
	Price       float64   `json:"price"`
	ImageURL    string    `json:"imageUrl"`
	AddedAt     time.Time `json:"addedAt"`
}

// AddItemRequest carries the payload for adding an item to a wishlist.
type AddItemRequest struct {
	ProductID   string  `json:"productId"`
	ProductName string  `json:"productName"`
	Price       float64 `json:"price"`
	ImageURL    string  `json:"imageUrl"`
}

// WishlistPage is a paginated response containing wishlist items.
type WishlistPage struct {
	CustomerID uuid.UUID      `json:"customerId"`
	Items      []*WishlistItem `json:"items"`
	Total      int            `json:"total"`
	Limit      int            `json:"limit"`
	Offset     int            `json:"offset"`
}
