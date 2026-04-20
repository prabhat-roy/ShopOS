package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/shopos/wishlist-service/internal/domain"
)

// Storer defines the persistence contract for wishlist operations.
type Storer interface {
	AddItem(ctx context.Context, item *domain.WishlistItem) error
	GetItem(ctx context.Context, customerID uuid.UUID, productID string) (*domain.WishlistItem, error)
	RemoveItem(ctx context.Context, customerID uuid.UUID, productID string) error
	ListItems(ctx context.Context, customerID uuid.UUID, limit, offset int) ([]*domain.WishlistItem, int, error)
	ClearWishlist(ctx context.Context, customerID uuid.UUID) error
	IsInWishlist(ctx context.Context, customerID uuid.UUID, productID string) (bool, error)
}

// PostgresStore implements Storer using PostgreSQL.
type PostgresStore struct {
	db *sql.DB
}

// New creates a new PostgresStore with the given *sql.DB connection.
func New(db *sql.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

// AddItem inserts a new wishlist item. Returns ErrAlreadyExists if the
// (customer_id, product_id) pair is already present.
func (s *PostgresStore) AddItem(ctx context.Context, item *domain.WishlistItem) error {
	query := `
		INSERT INTO wishlist_items (id, customer_id, product_id, product_name, price, image_url, added_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (customer_id, product_id) DO NOTHING`

	result, err := s.db.ExecContext(ctx, query,
		item.ID,
		item.CustomerID,
		item.ProductID,
		item.ProductName,
		item.Price,
		item.ImageURL,
		item.AddedAt,
	)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return domain.ErrAlreadyExists
		}
		return fmt.Errorf("store.AddItem: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("store.AddItem rows affected: %w", err)
	}
	if rows == 0 {
		return domain.ErrAlreadyExists
	}
	return nil
}

// GetItem retrieves a single wishlist item by customer and product.
func (s *PostgresStore) GetItem(ctx context.Context, customerID uuid.UUID, productID string) (*domain.WishlistItem, error) {
	query := `
		SELECT id, customer_id, product_id, product_name, price, image_url, added_at
		FROM wishlist_items
		WHERE customer_id = $1 AND product_id = $2`

	row := s.db.QueryRowContext(ctx, query, customerID, productID)
	return scanItem(row)
}

// RemoveItem deletes a wishlist item by customer and product.
func (s *PostgresStore) RemoveItem(ctx context.Context, customerID uuid.UUID, productID string) error {
	query := `DELETE FROM wishlist_items WHERE customer_id = $1 AND product_id = $2`
	result, err := s.db.ExecContext(ctx, query, customerID, productID)
	if err != nil {
		return fmt.Errorf("store.RemoveItem: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("store.RemoveItem rows affected: %w", err)
	}
	if rows == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// ListItems returns a paginated list of wishlist items for a customer along with the total count.
func (s *PostgresStore) ListItems(ctx context.Context, customerID uuid.UUID, limit, offset int) ([]*domain.WishlistItem, int, error) {
	countQuery := `SELECT COUNT(*) FROM wishlist_items WHERE customer_id = $1`
	var total int
	if err := s.db.QueryRowContext(ctx, countQuery, customerID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("store.ListItems count: %w", err)
	}

	query := `
		SELECT id, customer_id, product_id, product_name, price, image_url, added_at
		FROM wishlist_items
		WHERE customer_id = $1
		ORDER BY added_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := s.db.QueryContext(ctx, query, customerID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("store.ListItems query: %w", err)
	}
	defer rows.Close()

	var items []*domain.WishlistItem
	for rows.Next() {
		item, err := scanRow(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("store.ListItems scan: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("store.ListItems rows: %w", err)
	}

	if items == nil {
		items = []*domain.WishlistItem{}
	}
	return items, total, nil
}

// ClearWishlist removes all wishlist items for a customer.
func (s *PostgresStore) ClearWishlist(ctx context.Context, customerID uuid.UUID) error {
	query := `DELETE FROM wishlist_items WHERE customer_id = $1`
	if _, err := s.db.ExecContext(ctx, query, customerID); err != nil {
		return fmt.Errorf("store.ClearWishlist: %w", err)
	}
	return nil
}

// IsInWishlist returns true if the product is in the customer's wishlist.
func (s *PostgresStore) IsInWishlist(ctx context.Context, customerID uuid.UUID, productID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM wishlist_items WHERE customer_id = $1 AND product_id = $2)`
	var exists bool
	if err := s.db.QueryRowContext(ctx, query, customerID, productID).Scan(&exists); err != nil {
		return false, fmt.Errorf("store.IsInWishlist: %w", err)
	}
	return exists, nil
}

// scanItem scans a single row from a *sql.Row.
func scanItem(row *sql.Row) (*domain.WishlistItem, error) {
	var item domain.WishlistItem
	err := row.Scan(
		&item.ID,
		&item.CustomerID,
		&item.ProductID,
		&item.ProductName,
		&item.Price,
		&item.ImageURL,
		&item.AddedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scanItem: %w", err)
	}
	item.AddedAt = item.AddedAt.UTC()
	return &item, nil
}

// scanRow scans a single row from *sql.Rows.
func scanRow(rows *sql.Rows) (*domain.WishlistItem, error) {
	var item domain.WishlistItem
	var addedAt time.Time
	err := rows.Scan(
		&item.ID,
		&item.CustomerID,
		&item.ProductID,
		&item.ProductName,
		&item.Price,
		&item.ImageURL,
		&addedAt,
	)
	if err != nil {
		return nil, err
	}
	item.AddedAt = addedAt.UTC()
	return &item, nil
}
