package domain

import (
	"errors"
	"time"
)

// Sentinel errors returned by the store and service layers.
var (
	ErrNotFound          = errors.New("warehouse: resource not found")
	ErrInsufficientStock = errors.New("warehouse: insufficient stock for requested operation")
)

// Warehouse represents a physical warehouse location.
type Warehouse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Location  string    `json:"location"`
	Address   string    `json:"address"`
	Capacity  int       `json:"capacity"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// MovementType classifies a stock movement direction.
type MovementType string

const (
	MovementInbound  MovementType = "inbound"
	MovementOutbound MovementType = "outbound"
)

// StockMovement records the transfer of product units into or out of a warehouse.
type StockMovement struct {
	ID           string       `json:"id"`
	WarehouseID  string       `json:"warehouseId"`
	ProductID    string       `json:"productId"`
	SKU          string       `json:"sku"`
	MovementType MovementType `json:"movementType"`
	Quantity     int          `json:"quantity"`
	ReferenceID  string       `json:"referenceId"`
	Notes        string       `json:"notes"`
	CreatedAt    time.Time    `json:"createdAt"`
}
