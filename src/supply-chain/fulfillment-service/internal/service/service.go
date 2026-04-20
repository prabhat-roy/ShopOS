package service

import (
	"fmt"

	"github.com/shopos/fulfillment-service/internal/domain"
	"github.com/shopos/fulfillment-service/internal/store"
)

// Servicer defines the business logic contract for the fulfillment domain.
type Servicer interface {
	CreateFulfillment(f *domain.FulfillmentOrder) (*domain.FulfillmentOrder, error)
	GetFulfillment(id string) (*domain.FulfillmentOrder, error)
	ListFulfillments(orderID string) ([]*domain.FulfillmentOrder, error)
	GetByOrderID(orderID string) (*domain.FulfillmentOrder, error)
	StartPicking(id string) error
	StartPacking(id string) error
	MarkReadyToShip(id string) error
	Ship(id, trackingNumber, carrier string) error
	Deliver(id string) error
	Cancel(id string) error
}

// Service implements Servicer using a Storer for persistence.
type Service struct {
	store store.Storer
}

// New constructs a Service wired to the provided Storer.
func New(s store.Storer) *Service {
	return &Service{store: s}
}

// CreateFulfillment validates the input and persists a new fulfillment order.
func (svc *Service) CreateFulfillment(f *domain.FulfillmentOrder) (*domain.FulfillmentOrder, error) {
	if f.OrderID == "" {
		return nil, fmt.Errorf("service: orderId is required")
	}
	if f.WarehouseID == "" {
		return nil, fmt.Errorf("service: warehouseId is required")
	}
	if len(f.Items) == 0 {
		return nil, fmt.Errorf("service: at least one item is required")
	}
	for i, item := range f.Items {
		if item.ProductID == "" {
			return nil, fmt.Errorf("service: item[%d] productId is required", i)
		}
		if item.SKU == "" {
			return nil, fmt.Errorf("service: item[%d] sku is required", i)
		}
		if item.Quantity <= 0 {
			return nil, fmt.Errorf("service: item[%d] quantity must be positive", i)
		}
	}
	f.Status = domain.StatusPending
	if err := svc.store.Create(f); err != nil {
		return nil, err
	}
	return f, nil
}

// GetFulfillment retrieves a fulfillment order by ID.
func (svc *Service) GetFulfillment(id string) (*domain.FulfillmentOrder, error) {
	if id == "" {
		return nil, fmt.Errorf("service: id is required")
	}
	return svc.store.Get(id)
}

// ListFulfillments returns fulfillment orders. When orderID is non-empty, only
// orders for that commerce order are returned.
func (svc *Service) ListFulfillments(orderID string) ([]*domain.FulfillmentOrder, error) {
	return svc.store.List(orderID)
}

// GetByOrderID returns the most recent fulfillment for a given commerce order.
func (svc *Service) GetByOrderID(orderID string) (*domain.FulfillmentOrder, error) {
	if orderID == "" {
		return nil, fmt.Errorf("service: orderId is required")
	}
	return svc.store.GetByOrderID(orderID)
}

// StartPicking transitions a fulfillment order from PENDING to PICKING.
func (svc *Service) StartPicking(id string) error {
	return svc.transition(id, domain.StatusPicking)
}

// StartPacking transitions a fulfillment order from PICKING to PACKING.
func (svc *Service) StartPacking(id string) error {
	return svc.transition(id, domain.StatusPacking)
}

// MarkReadyToShip transitions a fulfillment order from PACKING to READY_TO_SHIP.
func (svc *Service) MarkReadyToShip(id string) error {
	return svc.transition(id, domain.StatusReadyToShip)
}

// Ship transitions a fulfillment order to SHIPPED and records tracking info.
func (svc *Service) Ship(id, trackingNumber, carrier string) error {
	if trackingNumber == "" {
		return fmt.Errorf("service: trackingNumber is required to ship")
	}
	if carrier == "" {
		return fmt.Errorf("service: carrier is required to ship")
	}
	f, err := svc.store.Get(id)
	if err != nil {
		return err
	}
	if !domain.CanTransition(f.Status, domain.StatusShipped) {
		return fmt.Errorf("%w: cannot move from %s to %s", domain.ErrInvalidTransition, f.Status, domain.StatusShipped)
	}
	if err := svc.store.UpdateTracking(id, trackingNumber, carrier); err != nil {
		return err
	}
	return svc.store.UpdateStatus(id, domain.StatusShipped)
}

// Deliver transitions a fulfillment order from SHIPPED to DELIVERED.
func (svc *Service) Deliver(id string) error {
	return svc.transition(id, domain.StatusDelivered)
}

// Cancel transitions a fulfillment order to CANCELLED from any cancellable state.
func (svc *Service) Cancel(id string) error {
	return svc.transition(id, domain.StatusCancelled)
}

// transition fetches the current state, validates the transition, and persists.
func (svc *Service) transition(id string, next domain.FulfillmentStatus) error {
	if id == "" {
		return fmt.Errorf("service: id is required")
	}
	f, err := svc.store.Get(id)
	if err != nil {
		return err
	}
	if !domain.CanTransition(f.Status, next) {
		return fmt.Errorf("%w: cannot move from %s to %s", domain.ErrInvalidTransition, f.Status, next)
	}
	return svc.store.UpdateStatus(id, next)
}
