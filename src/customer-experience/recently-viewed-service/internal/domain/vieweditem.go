package domain

import "time"

// ViewedItem represents a product that a customer has viewed.
type ViewedItem struct {
	ProductID   string    `json:"productId"`
	ProductName string    `json:"productName"`
	ImageURL    string    `json:"imageUrl"`
	Price       float64   `json:"price"`
	ViewedAt    time.Time `json:"viewedAt"`
}

// RecentlyViewedList is the response payload for a customer's recently viewed products.
type RecentlyViewedList struct {
	CustomerID string       `json:"customerId"`
	Items      []ViewedItem `json:"items"`
	Total      int          `json:"total"`
}
