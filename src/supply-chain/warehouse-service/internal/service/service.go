package service

import (
	"fmt"

	"github.com/shopos/warehouse-service/internal/domain"
	"github.com/shopos/warehouse-service/internal/store"
)

// Servicer defines the business logic contract for the warehouse domain.
type Servicer interface {
	CreateWarehouse(w *domain.Warehouse) (*domain.Warehouse, error)
	GetWarehouse(id string) (*domain.Warehouse, error)
	ListWarehouses(activeOnly bool) ([]*domain.Warehouse, error)
	UpdateWarehouse(id string, update *domain.Warehouse) (*domain.Warehouse, error)
	ReceiveStock(m *domain.StockMovement) (*domain.StockMovement, error)
	ShipStock(m *domain.StockMovement) (*domain.StockMovement, error)
	GetStockLevel(warehouseID, productID string) (int, error)
	ListMovements(warehouseID string, limit int) ([]*domain.StockMovement, error)
}

// Service implements Servicer using a Storer for persistence.
type Service struct {
	store store.Storer
}

// New constructs a Service wired to the provided Storer.
func New(s store.Storer) *Service {
	return &Service{store: s}
}

// CreateWarehouse validates the input and persists a new warehouse.
func (svc *Service) CreateWarehouse(w *domain.Warehouse) (*domain.Warehouse, error) {
	if w.Name == "" {
		return nil, fmt.Errorf("service: warehouse name is required")
	}
	if w.Capacity < 0 {
		return nil, fmt.Errorf("service: warehouse capacity must be non-negative")
	}
	// Default active to true for new warehouses unless explicitly set.
	if err := svc.store.CreateWarehouse(w); err != nil {
		return nil, err
	}
	return w, nil
}

// GetWarehouse retrieves a warehouse by ID.
func (svc *Service) GetWarehouse(id string) (*domain.Warehouse, error) {
	if id == "" {
		return nil, fmt.Errorf("service: warehouse id is required")
	}
	return svc.store.GetWarehouse(id)
}

// ListWarehouses returns all warehouses, optionally limited to active ones.
func (svc *Service) ListWarehouses(activeOnly bool) ([]*domain.Warehouse, error) {
	return svc.store.ListWarehouses(activeOnly)
}

// UpdateWarehouse fetches an existing warehouse, applies the supplied update
// fields, and persists the result.
func (svc *Service) UpdateWarehouse(id string, update *domain.Warehouse) (*domain.Warehouse, error) {
	if id == "" {
		return nil, fmt.Errorf("service: warehouse id is required")
	}
	existing, err := svc.store.GetWarehouse(id)
	if err != nil {
		return nil, err
	}
	if update.Name != "" {
		existing.Name = update.Name
	}
	if update.Location != "" {
		existing.Location = update.Location
	}
	if update.Address != "" {
		existing.Address = update.Address
	}
	if update.Capacity > 0 {
		existing.Capacity = update.Capacity
	}
	// Allow explicit false to deactivate.
	existing.Active = update.Active

	if err := svc.store.UpdateWarehouse(existing); err != nil {
		return nil, err
	}
	return existing, nil
}

// ReceiveStock records an inbound stock movement for a warehouse.
func (svc *Service) ReceiveStock(m *domain.StockMovement) (*domain.StockMovement, error) {
	if err := svc.validateMovement(m); err != nil {
		return nil, err
	}
	m.MovementType = domain.MovementInbound
	if _, err := svc.store.GetWarehouse(m.WarehouseID); err != nil {
		return nil, err
	}
	if err := svc.store.RecordMovement(m); err != nil {
		return nil, err
	}
	return m, nil
}

// ShipStock records an outbound stock movement. The store layer checks that
// sufficient stock exists; ErrInsufficientStock is returned if not.
func (svc *Service) ShipStock(m *domain.StockMovement) (*domain.StockMovement, error) {
	if err := svc.validateMovement(m); err != nil {
		return nil, err
	}
	m.MovementType = domain.MovementOutbound
	if _, err := svc.store.GetWarehouse(m.WarehouseID); err != nil {
		return nil, err
	}
	if err := svc.store.RecordMovement(m); err != nil {
		return nil, err
	}
	return m, nil
}

// GetStockLevel returns the current on-hand quantity for a product in a warehouse.
func (svc *Service) GetStockLevel(warehouseID, productID string) (int, error) {
	if warehouseID == "" || productID == "" {
		return 0, fmt.Errorf("service: warehouseId and productId are required")
	}
	if _, err := svc.store.GetWarehouse(warehouseID); err != nil {
		return 0, err
	}
	return svc.store.GetStock(warehouseID, productID)
}

// ListMovements returns recent stock movements for a warehouse.
func (svc *Service) ListMovements(warehouseID string, limit int) ([]*domain.StockMovement, error) {
	if warehouseID == "" {
		return nil, fmt.Errorf("service: warehouseId is required")
	}
	if _, err := svc.store.GetWarehouse(warehouseID); err != nil {
		return nil, err
	}
	return svc.store.ListMovements(warehouseID, limit)
}

// validateMovement enforces basic movement preconditions.
func (svc *Service) validateMovement(m *domain.StockMovement) error {
	if m.WarehouseID == "" {
		return fmt.Errorf("service: warehouseId is required")
	}
	if m.ProductID == "" {
		return fmt.Errorf("service: productId is required")
	}
	if m.SKU == "" {
		return fmt.Errorf("service: sku is required")
	}
	if m.Quantity <= 0 {
		return fmt.Errorf("service: quantity must be positive")
	}
	return nil
}
