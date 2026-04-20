package store

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/shopos/warehouse-service/internal/domain"
)

// Storer defines the data-access contract for warehouse persistence.
type Storer interface {
	CreateWarehouse(w *domain.Warehouse) error
	GetWarehouse(id string) (*domain.Warehouse, error)
	ListWarehouses(activeOnly bool) ([]*domain.Warehouse, error)
	UpdateWarehouse(w *domain.Warehouse) error
	GetStock(warehouseID, productID string) (int, error)
	RecordMovement(m *domain.StockMovement) error
	ListMovements(warehouseID string, limit int) ([]*domain.StockMovement, error)
}

// PostgresStore implements Storer against a PostgreSQL database.
type PostgresStore struct {
	db *sql.DB
}

// New opens a PostgreSQL connection and returns a ready-to-use PostgresStore.
func New(databaseURL string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("store: open db: %w", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	return &PostgresStore{db: db}, nil
}

// Close releases the underlying database connection pool.
func (s *PostgresStore) Close() error {
	return s.db.Close()
}

// CreateWarehouse inserts a new warehouse row.
func (s *PostgresStore) CreateWarehouse(w *domain.Warehouse) error {
	if w.ID == "" {
		w.ID = uuid.NewString()
	}
	now := time.Now().UTC()
	w.CreatedAt = now
	w.UpdatedAt = now

	const q = `
		INSERT INTO warehouses (id, name, location, address, capacity, active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := s.db.Exec(q, w.ID, w.Name, w.Location, w.Address, w.Capacity, w.Active, w.CreatedAt, w.UpdatedAt)
	if err != nil {
		return fmt.Errorf("store: create warehouse: %w", err)
	}
	return nil
}

// GetWarehouse retrieves a single warehouse by primary key.
func (s *PostgresStore) GetWarehouse(id string) (*domain.Warehouse, error) {
	const q = `
		SELECT id, name, location, address, capacity, active, created_at, updated_at
		FROM warehouses WHERE id = $1`
	row := s.db.QueryRow(q, id)
	w := &domain.Warehouse{}
	err := row.Scan(&w.ID, &w.Name, &w.Location, &w.Address, &w.Capacity, &w.Active, &w.CreatedAt, &w.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("store: get warehouse: %w", err)
	}
	return w, nil
}

// ListWarehouses returns all warehouses, optionally filtering to active only.
func (s *PostgresStore) ListWarehouses(activeOnly bool) ([]*domain.Warehouse, error) {
	q := `SELECT id, name, location, address, capacity, active, created_at, updated_at FROM warehouses`
	if activeOnly {
		q += ` WHERE active = TRUE`
	}
	q += ` ORDER BY created_at DESC`

	rows, err := s.db.Query(q)
	if err != nil {
		return nil, fmt.Errorf("store: list warehouses: %w", err)
	}
	defer rows.Close()

	var result []*domain.Warehouse
	for rows.Next() {
		w := &domain.Warehouse{}
		if err := rows.Scan(&w.ID, &w.Name, &w.Location, &w.Address, &w.Capacity, &w.Active, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, fmt.Errorf("store: scan warehouse: %w", err)
		}
		result = append(result, w)
	}
	return result, rows.Err()
}

// UpdateWarehouse writes mutable fields back to the database.
func (s *PostgresStore) UpdateWarehouse(w *domain.Warehouse) error {
	w.UpdatedAt = time.Now().UTC()
	const q = `
		UPDATE warehouses
		SET name=$1, location=$2, address=$3, capacity=$4, active=$5, updated_at=$6
		WHERE id=$7`
	res, err := s.db.Exec(q, w.Name, w.Location, w.Address, w.Capacity, w.Active, w.UpdatedAt, w.ID)
	if err != nil {
		return fmt.Errorf("store: update warehouse: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// GetStock returns the current on-hand quantity for a product in a warehouse.
// It sums all movements, adding inbound and subtracting outbound quantities.
func (s *PostgresStore) GetStock(warehouseID, productID string) (int, error) {
	const q = `
		SELECT COALESCE(
			SUM(CASE WHEN movement_type = 'inbound' THEN quantity ELSE -quantity END),
			0
		)
		FROM stock_movements
		WHERE warehouse_id = $1 AND product_id = $2`
	var stock int
	if err := s.db.QueryRow(q, warehouseID, productID).Scan(&stock); err != nil {
		return 0, fmt.Errorf("store: get stock: %w", err)
	}
	return stock, nil
}

// RecordMovement persists a stock movement. For outbound movements it verifies
// sufficient stock exists before inserting, returning ErrInsufficientStock if not.
func (s *PostgresStore) RecordMovement(m *domain.StockMovement) error {
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	m.CreatedAt = time.Now().UTC()

	// For outbound, check current stock level first.
	if m.MovementType == domain.MovementOutbound {
		current, err := s.GetStock(m.WarehouseID, m.ProductID)
		if err != nil {
			return err
		}
		if current < m.Quantity {
			return domain.ErrInsufficientStock
		}
	}

	const q = `
		INSERT INTO stock_movements (id, warehouse_id, product_id, sku, movement_type, quantity, reference_id, notes, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := s.db.Exec(q, m.ID, m.WarehouseID, m.ProductID, m.SKU, m.MovementType, m.Quantity, m.ReferenceID, m.Notes, m.CreatedAt)
	if err != nil {
		return fmt.Errorf("store: record movement: %w", err)
	}
	return nil
}

// ListMovements returns the most recent movements for a warehouse, up to limit rows.
func (s *PostgresStore) ListMovements(warehouseID string, limit int) ([]*domain.StockMovement, error) {
	if limit <= 0 {
		limit = 50
	}
	const q = `
		SELECT id, warehouse_id, product_id, sku, movement_type, quantity, reference_id, notes, created_at
		FROM stock_movements
		WHERE warehouse_id = $1
		ORDER BY created_at DESC
		LIMIT $2`
	rows, err := s.db.Query(q, warehouseID, limit)
	if err != nil {
		return nil, fmt.Errorf("store: list movements: %w", err)
	}
	defer rows.Close()

	var result []*domain.StockMovement
	for rows.Next() {
		m := &domain.StockMovement{}
		if err := rows.Scan(&m.ID, &m.WarehouseID, &m.ProductID, &m.SKU, &m.MovementType, &m.Quantity, &m.ReferenceID, &m.Notes, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("store: scan movement: %w", err)
		}
		result = append(result, m)
	}
	return result, rows.Err()
}
