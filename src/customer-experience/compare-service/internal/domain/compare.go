package domain

import (
	"errors"
	"time"
)

// ErrListFull is returned when a customer's compare list already contains the maximum number of items.
var ErrListFull = errors.New("compare list is full: maximum 4 items allowed")

// ErrItemNotFound is returned when a product is not present in the compare list.
var ErrItemNotFound = errors.New("item not found in compare list")

// CompareItem represents a single product in a comparison list.
type CompareItem struct {
	ProductID   string            `json:"productId"`
	ProductName string            `json:"productName"`
	Price       float64           `json:"price"`
	ImageURL    string            `json:"imageUrl"`
	Attributes  map[string]string `json:"attributes"`
}

// CompareList represents the full comparison list for a customer.
type CompareList struct {
	CustomerID string        `json:"customerId"`
	Items      []CompareItem `json:"items"`
	UpdatedAt  time.Time     `json:"updatedAt"`
}
