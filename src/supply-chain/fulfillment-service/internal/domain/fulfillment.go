package domain

import (
	"errors"
	"time"
)

// Sentinel errors returned by the service layer.
var (
	ErrNotFound         = errors.New("fulfillment: resource not found")
	ErrInvalidTransition = errors.New("fulfillment: invalid status transition")
)

// FulfillmentStatus represents the lifecycle state of a fulfillment order.
type FulfillmentStatus string

const (
	StatusPending      FulfillmentStatus = "PENDING"
	StatusPicking      FulfillmentStatus = "PICKING"
	StatusPacking      FulfillmentStatus = "PACKING"
	StatusReadyToShip  FulfillmentStatus = "READY_TO_SHIP"
	StatusShipped      FulfillmentStatus = "SHIPPED"
	StatusDelivered    FulfillmentStatus = "DELIVERED"
	StatusCancelled    FulfillmentStatus = "CANCELLED"
)

// validTransitions defines which status moves are legal from a given state.
var validTransitions = map[FulfillmentStatus][]FulfillmentStatus{
	StatusPending:     {StatusPicking, StatusCancelled},
	StatusPicking:     {StatusPacking, StatusCancelled},
	StatusPacking:     {StatusReadyToShip, StatusCancelled},
	StatusReadyToShip: {StatusShipped, StatusCancelled},
	StatusShipped:     {StatusDelivered},
	StatusDelivered:   {},
	StatusCancelled:   {},
}

// CanTransition returns true when moving from current to next is a legal transition.
func CanTransition(current, next FulfillmentStatus) bool {
	allowed, ok := validTransitions[current]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == next {
			return true
		}
	}
	return false
}

// FulfillmentItem is a single line in a fulfillment order.
type FulfillmentItem struct {
	ID            string `json:"id"`
	FulfillmentID string `json:"fulfillmentId"`
	ProductID     string `json:"productId"`
	SKU           string `json:"sku"`
	Quantity      int    `json:"quantity"`
	PickedQty     int    `json:"pickedQty"`
}

// FulfillmentOrder represents the top-level fulfillment record for an order.
type FulfillmentOrder struct {
	ID              string            `json:"id"`
	OrderID         string            `json:"orderId"`
	WarehouseID     string            `json:"warehouseId"`
	Status          FulfillmentStatus `json:"status"`
	TrackingNumber  string            `json:"trackingNumber"`
	Carrier         string            `json:"carrier"`
	Items           []FulfillmentItem `json:"items"`
	ShippingAddress string            `json:"shippingAddress"`
	Notes           string            `json:"notes"`
	CreatedAt       time.Time         `json:"createdAt"`
	UpdatedAt       time.Time         `json:"updatedAt"`
}
