package store

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/shopos/loyalty-service/domain"
)

// Store handles all Postgres interactions for the loyalty service.
type Store struct {
	db *sql.DB
}

// New opens a Postgres connection and verifies it with a ping.
func New(databaseURL string) (*Store, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	return &Store{db: db}, nil
}

// Close closes the underlying database connection pool.
func (s *Store) Close() error {
	return s.db.Close()
}

// GetAccount returns the loyalty account for a customer.
// Returns domain.ErrNotFound when no account exists.
func (s *Store) GetAccount(customerID string) (*domain.LoyaltyAccount, error) {
	const q = `
		SELECT customer_id, points, tier_name, created_at, updated_at
		FROM loyalty_accounts
		WHERE customer_id = $1`

	row := s.db.QueryRow(q, customerID)
	var a domain.LoyaltyAccount
	err := row.Scan(&a.CustomerID, &a.Points, &a.TierName, &a.CreatedAt, &a.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get account: %w", err)
	}
	return &a, nil
}

// CreateAccount inserts a new loyalty account with zero points.
func (s *Store) CreateAccount(customerID string) (*domain.LoyaltyAccount, error) {
	const q = `
		INSERT INTO loyalty_accounts (customer_id, points, tier_name, created_at, updated_at)
		VALUES ($1, 0, 'Bronze', NOW(), NOW())
		RETURNING customer_id, points, tier_name, created_at, updated_at`

	row := s.db.QueryRow(q, customerID)
	var a domain.LoyaltyAccount
	if err := row.Scan(&a.CustomerID, &a.Points, &a.TierName, &a.CreatedAt, &a.UpdatedAt); err != nil {
		return nil, fmt.Errorf("create account: %w", err)
	}
	return &a, nil
}

// EarnPoints adds points to the account and records the transaction. All within a single transaction.
func (s *Store) EarnPoints(customerID string, points int64, orderID, description string) (*domain.PointTransaction, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	// Lock and update the account row.
	const updateQ = `
		UPDATE loyalty_accounts
		SET points = points + $1, updated_at = NOW()
		WHERE customer_id = $2
		RETURNING points`

	var newBalance int64
	if err := tx.QueryRow(updateQ, points, customerID).Scan(&newBalance); err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("update points: %w", err)
	}

	txn, err := insertTransaction(tx, customerID, domain.TxEarn, points, newBalance, orderID, description)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit earn: %w", err)
	}
	return txn, nil
}

// RedeemPoints validates balance and deducts points atomically.
// Returns domain.ErrInsufficientPoints when the account lacks sufficient points.
func (s *Store) RedeemPoints(customerID string, points int64, orderID string) (*domain.PointTransaction, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	// Read current balance with a row lock.
	const selectQ = `SELECT points FROM loyalty_accounts WHERE customer_id = $1 FOR UPDATE`
	var current int64
	if err := tx.QueryRow(selectQ, customerID).Scan(&current); err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("lock account: %w", err)
	}
	if current < points {
		return nil, domain.ErrInsufficientPoints
	}

	const updateQ = `
		UPDATE loyalty_accounts
		SET points = points - $1, updated_at = NOW()
		WHERE customer_id = $2
		RETURNING points`

	var newBalance int64
	if err := tx.QueryRow(updateQ, points, customerID).Scan(&newBalance); err != nil {
		return nil, fmt.Errorf("deduct points: %w", err)
	}

	desc := fmt.Sprintf("Redeemed %d points", points)
	txn, err := insertTransaction(tx, customerID, domain.TxRedeem, points, newBalance, orderID, desc)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit redeem: %w", err)
	}
	return txn, nil
}

// GetTransactions returns the most recent transactions for a customer (descending order).
func (s *Store) GetTransactions(customerID string, limit int) ([]domain.PointTransaction, error) {
	const q = `
		SELECT id, customer_id, type, points, balance, order_id, description, created_at
		FROM point_transactions
		WHERE customer_id = $1
		ORDER BY created_at DESC
		LIMIT $2`

	rows, err := s.db.Query(q, customerID, limit)
	if err != nil {
		return nil, fmt.Errorf("query transactions: %w", err)
	}
	defer rows.Close()

	var txns []domain.PointTransaction
	for rows.Next() {
		var t domain.PointTransaction
		if err := rows.Scan(&t.ID, &t.CustomerID, &t.Type, &t.Points, &t.Balance,
			&t.OrderID, &t.Description, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan transaction: %w", err)
		}
		txns = append(txns, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return txns, nil
}

// insertTransaction writes a point_transaction row inside an existing DB transaction.
func insertTransaction(tx *sql.Tx, customerID string, txType domain.TransactionType,
	points, balance int64, orderID, description string) (*domain.PointTransaction, error) {

	const q = `
		INSERT INTO point_transactions
			(id, customer_id, type, points, balance, order_id, description, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, customer_id, type, points, balance, order_id, description, created_at`

	id := uuid.NewString()
	now := time.Now().UTC()
	row := tx.QueryRow(q, id, customerID, string(txType), points, balance, orderID, description, now)

	var t domain.PointTransaction
	if err := row.Scan(&t.ID, &t.CustomerID, &t.Type, &t.Points, &t.Balance,
		&t.OrderID, &t.Description, &t.CreatedAt); err != nil {
		return nil, fmt.Errorf("insert transaction: %w", err)
	}
	return &t, nil
}
