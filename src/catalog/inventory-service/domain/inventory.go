package domain

import (
	"errors"
	"time"
)

// StockLevel represents current stock for a product in a specific warehouse.
type StockLevel struct {
	ID          string    `json:"id"`
	ProductID   string    `json:"product_id"`
	SKU         string    `json:"sku"`
	WarehouseID string    `json:"warehouse_id"`
	Available   int       `json:"available"`
	Reserved    int       `json:"reserved"`
	Reorder     int       `json:"reorder_point"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ReservationStatus represents the lifecycle state of a stock reservation.
type ReservationStatus string

const (
	ReservedStatus  ReservationStatus = "reserved"
	ReleasedStatus  ReservationStatus = "released"
	CommittedStatus ReservationStatus = "committed"
)

// Reservation tracks a quantity hold placed against a stock level for an order.
type Reservation struct {
	ID        string            `json:"id"`
	OrderID   string            `json:"order_id"`
	ProductID string            `json:"product_id"`
	Quantity  int               `json:"quantity"`
	Status    ReservationStatus `json:"status"`
	CreatedAt time.Time         `json:"created_at"`
}

var ErrNotFound = errors.New("not found")
var ErrInsufficientStock = errors.New("insufficient stock")
