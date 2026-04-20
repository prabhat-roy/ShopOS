package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/shopos/inventory-service/domain"
)

// Store handles all Postgres persistence for the inventory service.
type Store struct {
	db *sql.DB
}

// New opens a Postgres connection and returns a ready Store.
func New(dsn string) (*Store, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("store: open: %w", err)
	}
	return &Store{db: db}, nil
}

// Close releases the underlying database connection pool.
func (s *Store) Close() error {
	return s.db.Close()
}

// GetStock returns the stock level for a specific product+warehouse combination.
func (s *Store) GetStock(ctx context.Context, productID, warehouseID string) (*domain.StockLevel, error) {
	const q = `
		SELECT id, product_id, sku, warehouse_id, available, reserved, reorder_point, updated_at
		FROM stock_levels
		WHERE product_id = $1 AND warehouse_id = $2`

	row := s.db.QueryRowContext(ctx, q, productID, warehouseID)
	sl, err := scanStockLevel(row)
	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("store: GetStock: %w", err)
	}
	return sl, nil
}

// ListStock returns all warehouse stock levels for the given product.
func (s *Store) ListStock(ctx context.Context, productID string) ([]*domain.StockLevel, error) {
	const q = `
		SELECT id, product_id, sku, warehouse_id, available, reserved, reorder_point, updated_at
		FROM stock_levels
		WHERE product_id = $1
		ORDER BY warehouse_id`

	rows, err := s.db.QueryContext(ctx, q, productID)
	if err != nil {
		return nil, fmt.Errorf("store: ListStock: %w", err)
	}
	defer rows.Close()

	var levels []*domain.StockLevel
	for rows.Next() {
		sl, err := scanStockLevelRows(rows)
		if err != nil {
			return nil, fmt.Errorf("store: ListStock scan: %w", err)
		}
		levels = append(levels, sl)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: ListStock rows: %w", err)
	}
	return levels, nil
}

// UpsertStock creates or updates a stock level record for a product+warehouse.
func (s *Store) UpsertStock(ctx context.Context, productID, sku, warehouseID string, available, reorder int) (*domain.StockLevel, error) {
	const q = `
		INSERT INTO stock_levels (id, product_id, sku, warehouse_id, available, reorder_point, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (product_id, warehouse_id)
		DO UPDATE SET
			sku           = EXCLUDED.sku,
			available     = EXCLUDED.available,
			reorder_point = EXCLUDED.reorder_point,
			updated_at    = EXCLUDED.updated_at
		RETURNING id, product_id, sku, warehouse_id, available, reserved, reorder_point, updated_at`

	id := uuid.New().String()
	now := time.Now().UTC()
	row := s.db.QueryRowContext(ctx, q, id, productID, sku, warehouseID, available, reorder, now)
	sl, err := scanStockLevel(row)
	if err != nil {
		return nil, fmt.Errorf("store: UpsertStock: %w", err)
	}
	return sl, nil
}

// Reserve deducts qty from available and adds to reserved atomically.
// Returns ErrInsufficientStock when available < qty.
func (s *Store) Reserve(ctx context.Context, orderID, productID string, qty int) (*domain.Reservation, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("store: Reserve: begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	// Lock the row and check available stock.
	const lockQ = `
		SELECT id, available
		FROM stock_levels
		WHERE product_id = $1
		ORDER BY available DESC
		LIMIT 1
		FOR UPDATE`

	var slID string
	var avail int
	if err := tx.QueryRowContext(ctx, lockQ, productID).Scan(&slID, &avail); err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("store: Reserve: lock row: %w", err)
	}

	if avail < qty {
		return nil, domain.ErrInsufficientStock
	}

	// Deduct available, increment reserved.
	const updateQ = `
		UPDATE stock_levels
		SET available  = available  - $1,
		    reserved   = reserved   + $1,
		    updated_at = NOW()
		WHERE id = $2`
	if _, err := tx.ExecContext(ctx, updateQ, qty, slID); err != nil {
		return nil, fmt.Errorf("store: Reserve: update stock: %w", err)
	}

	// Create the reservation record.
	res := &domain.Reservation{
		ID:        uuid.New().String(),
		OrderID:   orderID,
		ProductID: productID,
		Quantity:  qty,
		Status:    domain.ReservedStatus,
		CreatedAt: time.Now().UTC(),
	}
	const insQ = `
		INSERT INTO reservations (id, order_id, product_id, quantity, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`
	if _, err := tx.ExecContext(ctx, insQ, res.ID, res.OrderID, res.ProductID, res.Quantity, res.Status, res.CreatedAt); err != nil {
		return nil, fmt.Errorf("store: Reserve: insert reservation: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("store: Reserve: commit: %w", err)
	}
	return res, nil
}

// Release returns reserved qty back to available and marks the reservation as released.
func (s *Store) Release(ctx context.Context, reservationID string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("store: Release: begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	// Fetch and lock the reservation.
	const fetchQ = `
		SELECT product_id, quantity, status
		FROM reservations
		WHERE id = $1
		FOR UPDATE`

	var productID string
	var qty int
	var status string
	if err := tx.QueryRowContext(ctx, fetchQ, reservationID).Scan(&productID, &qty, &status); err == sql.ErrNoRows {
		return domain.ErrNotFound
	} else if err != nil {
		return fmt.Errorf("store: Release: fetch reservation: %w", err)
	}

	// Return qty to available.
	const updateStockQ = `
		UPDATE stock_levels
		SET reserved   = reserved   - $1,
		    available  = available  + $1,
		    updated_at = NOW()
		WHERE product_id = $2
		  AND reserved   >= $1`
	res, err := tx.ExecContext(ctx, updateStockQ, qty, productID)
	if err != nil {
		return fmt.Errorf("store: Release: update stock: %w", err)
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("store: Release: no stock row updated for product %s", productID)
	}

	// Mark reservation as released.
	const updateResQ = `UPDATE reservations SET status = $1 WHERE id = $2`
	if _, err := tx.ExecContext(ctx, updateResQ, string(domain.ReleasedStatus), reservationID); err != nil {
		return fmt.Errorf("store: Release: update reservation: %w", err)
	}

	return tx.Commit()
}

// Commit finalises a reservation — deducts from reserved (order shipped/fulfilled).
func (s *Store) Commit(ctx context.Context, reservationID string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("store: Commit: begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	const fetchQ = `
		SELECT product_id, quantity, status
		FROM reservations
		WHERE id = $1
		FOR UPDATE`

	var productID string
	var qty int
	var status string
	if err := tx.QueryRowContext(ctx, fetchQ, reservationID).Scan(&productID, &qty, &status); err == sql.ErrNoRows {
		return domain.ErrNotFound
	} else if err != nil {
		return fmt.Errorf("store: Commit: fetch reservation: %w", err)
	}

	// Deduct from reserved only (already out of available).
	const updateStockQ = `
		UPDATE stock_levels
		SET reserved   = reserved - $1,
		    updated_at = NOW()
		WHERE product_id = $2
		  AND reserved   >= $1`
	res, err := tx.ExecContext(ctx, updateStockQ, qty, productID)
	if err != nil {
		return fmt.Errorf("store: Commit: update stock: %w", err)
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("store: Commit: no stock row updated for product %s", productID)
	}

	const updateResQ = `UPDATE reservations SET status = $1 WHERE id = $2`
	if _, err := tx.ExecContext(ctx, updateResQ, string(domain.CommittedStatus), reservationID); err != nil {
		return fmt.Errorf("store: Commit: update reservation: %w", err)
	}

	return tx.Commit()
}

// GetReservation fetches a single reservation by ID.
func (s *Store) GetReservation(ctx context.Context, id string) (*domain.Reservation, error) {
	const q = `
		SELECT id, order_id, product_id, quantity, status, created_at
		FROM reservations
		WHERE id = $1`

	row := s.db.QueryRowContext(ctx, q, id)
	r := &domain.Reservation{}
	var status string
	err := row.Scan(&r.ID, &r.OrderID, &r.ProductID, &r.Quantity, &status, &r.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("store: GetReservation: %w", err)
	}
	r.Status = domain.ReservationStatus(status)
	return r, nil
}

// scanStockLevel scans a *sql.Row into a StockLevel.
func scanStockLevel(row *sql.Row) (*domain.StockLevel, error) {
	sl := &domain.StockLevel{}
	err := row.Scan(
		&sl.ID, &sl.ProductID, &sl.SKU, &sl.WarehouseID,
		&sl.Available, &sl.Reserved, &sl.Reorder, &sl.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return sl, nil
}

// scanStockLevelRows scans a *sql.Rows into a StockLevel.
func scanStockLevelRows(rows *sql.Rows) (*domain.StockLevel, error) {
	sl := &domain.StockLevel{}
	err := rows.Scan(
		&sl.ID, &sl.ProductID, &sl.SKU, &sl.WarehouseID,
		&sl.Available, &sl.Reserved, &sl.Reorder, &sl.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return sl, nil
}
