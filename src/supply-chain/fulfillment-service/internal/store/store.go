package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/shopos/fulfillment-service/internal/domain"
)

// Storer defines the data-access contract for fulfillment persistence.
type Storer interface {
	Create(f *domain.FulfillmentOrder) error
	Get(id string) (*domain.FulfillmentOrder, error)
	List(orderID string) ([]*domain.FulfillmentOrder, error)
	UpdateStatus(id string, status domain.FulfillmentStatus) error
	UpdateTracking(id, trackingNumber, carrier string) error
	GetByOrderID(orderID string) (*domain.FulfillmentOrder, error)
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

// Create inserts a new fulfillment order row. Items are serialized to JSONB.
func (s *PostgresStore) Create(f *domain.FulfillmentOrder) error {
	if f.ID == "" {
		f.ID = uuid.NewString()
	}
	now := time.Now().UTC()
	f.CreatedAt = now
	f.UpdatedAt = now
	if f.Status == "" {
		f.Status = domain.StatusPending
	}
	if f.Items == nil {
		f.Items = []domain.FulfillmentItem{}
	}

	// Assign IDs to items that don't have one yet.
	for i := range f.Items {
		if f.Items[i].ID == "" {
			f.Items[i].ID = uuid.NewString()
		}
		f.Items[i].FulfillmentID = f.ID
	}

	itemsJSON, err := json.Marshal(f.Items)
	if err != nil {
		return fmt.Errorf("store: marshal items: %w", err)
	}

	const q = `
		INSERT INTO fulfillments
			(id, order_id, warehouse_id, status, tracking_number, carrier, items, shipping_address, notes, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`
	_, err = s.db.Exec(q,
		f.ID, f.OrderID, f.WarehouseID, f.Status,
		f.TrackingNumber, f.Carrier, itemsJSON,
		f.ShippingAddress, f.Notes,
		f.CreatedAt, f.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("store: create fulfillment: %w", err)
	}
	return nil
}

// Get retrieves a single fulfillment order by primary key.
func (s *PostgresStore) Get(id string) (*domain.FulfillmentOrder, error) {
	const q = `
		SELECT id, order_id, warehouse_id, status, tracking_number, carrier, items,
		       shipping_address, notes, created_at, updated_at
		FROM fulfillments WHERE id = $1`
	row := s.db.QueryRow(q, id)
	return scanFulfillment(row)
}

// List returns fulfillment orders, optionally filtered by order_id.
func (s *PostgresStore) List(orderID string) ([]*domain.FulfillmentOrder, error) {
	var (
		rows *sql.Rows
		err  error
	)
	if orderID != "" {
		rows, err = s.db.Query(`
			SELECT id, order_id, warehouse_id, status, tracking_number, carrier, items,
			       shipping_address, notes, created_at, updated_at
			FROM fulfillments WHERE order_id = $1
			ORDER BY created_at DESC`, orderID)
	} else {
		rows, err = s.db.Query(`
			SELECT id, order_id, warehouse_id, status, tracking_number, carrier, items,
			       shipping_address, notes, created_at, updated_at
			FROM fulfillments ORDER BY created_at DESC`)
	}
	if err != nil {
		return nil, fmt.Errorf("store: list fulfillments: %w", err)
	}
	defer rows.Close()

	var result []*domain.FulfillmentOrder
	for rows.Next() {
		f, err := scanFulfillmentRow(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, f)
	}
	return result, rows.Err()
}

// UpdateStatus transitions the status of a fulfillment record.
func (s *PostgresStore) UpdateStatus(id string, status domain.FulfillmentStatus) error {
	const q = `UPDATE fulfillments SET status=$1, updated_at=$2 WHERE id=$3`
	res, err := s.db.Exec(q, status, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("store: update status: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// UpdateTracking sets the tracking number and carrier on a fulfillment record.
func (s *PostgresStore) UpdateTracking(id, trackingNumber, carrier string) error {
	const q = `UPDATE fulfillments SET tracking_number=$1, carrier=$2, updated_at=$3 WHERE id=$4`
	res, err := s.db.Exec(q, trackingNumber, carrier, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("store: update tracking: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// GetByOrderID retrieves the most recent fulfillment for a given order.
func (s *PostgresStore) GetByOrderID(orderID string) (*domain.FulfillmentOrder, error) {
	const q = `
		SELECT id, order_id, warehouse_id, status, tracking_number, carrier, items,
		       shipping_address, notes, created_at, updated_at
		FROM fulfillments WHERE order_id = $1
		ORDER BY created_at DESC LIMIT 1`
	row := s.db.QueryRow(q, orderID)
	return scanFulfillment(row)
}

// ---- scan helpers -----------------------------------------------------------

type scanner interface {
	Scan(dest ...any) error
}

func scanFulfillment(row *sql.Row) (*domain.FulfillmentOrder, error) {
	f := &domain.FulfillmentOrder{}
	var itemsJSON []byte
	err := row.Scan(
		&f.ID, &f.OrderID, &f.WarehouseID, &f.Status,
		&f.TrackingNumber, &f.Carrier, &itemsJSON,
		&f.ShippingAddress, &f.Notes,
		&f.CreatedAt, &f.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("store: scan fulfillment: %w", err)
	}
	if err := json.Unmarshal(itemsJSON, &f.Items); err != nil {
		return nil, fmt.Errorf("store: unmarshal items: %w", err)
	}
	return f, nil
}

func scanFulfillmentRow(rows *sql.Rows) (*domain.FulfillmentOrder, error) {
	f := &domain.FulfillmentOrder{}
	var itemsJSON []byte
	err := rows.Scan(
		&f.ID, &f.OrderID, &f.WarehouseID, &f.Status,
		&f.TrackingNumber, &f.Carrier, &itemsJSON,
		&f.ShippingAddress, &f.Notes,
		&f.CreatedAt, &f.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("store: scan fulfillment row: %w", err)
	}
	if err := json.Unmarshal(itemsJSON, &f.Items); err != nil {
		return nil, fmt.Errorf("store: unmarshal items: %w", err)
	}
	return f, nil
}
