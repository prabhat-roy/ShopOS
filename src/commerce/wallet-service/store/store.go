package store

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/shopos/wallet-service/domain"
)

// Store handles all Postgres interactions for the wallet service.
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

// GetWallet returns the wallet for a customer.
// Returns domain.ErrNotFound when no wallet exists.
func (s *Store) GetWallet(customerID string) (*domain.Wallet, error) {
	const q = `
		SELECT id, customer_id, balance, currency, created_at, updated_at
		FROM wallets
		WHERE customer_id = $1`

	row := s.db.QueryRow(q, customerID)
	var w domain.Wallet
	err := row.Scan(&w.ID, &w.CustomerID, &w.Balance, &w.Currency, &w.CreatedAt, &w.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get wallet: %w", err)
	}
	return &w, nil
}

// CreateWallet inserts a new wallet with zero balance.
func (s *Store) CreateWallet(customerID, currency string) (*domain.Wallet, error) {
	const q = `
		INSERT INTO wallets (id, customer_id, balance, currency, created_at, updated_at)
		VALUES ($1, $2, 0, $3, NOW(), NOW())
		RETURNING id, customer_id, balance, currency, created_at, updated_at`

	id := uuid.NewString()
	row := s.db.QueryRow(q, id, customerID, currency)
	var w domain.Wallet
	if err := row.Scan(&w.ID, &w.CustomerID, &w.Balance, &w.Currency, &w.CreatedAt, &w.UpdatedAt); err != nil {
		return nil, fmt.Errorf("create wallet: %w", err)
	}
	return &w, nil
}

// Credit adds funds to a wallet and records the transaction atomically.
func (s *Store) Credit(walletID string, amount float64, reference, description string) (*domain.WalletTransaction, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	const updateQ = `
		UPDATE wallets
		SET balance = balance + $1, updated_at = NOW()
		WHERE id = $2
		RETURNING id`

	var id string
	if err := tx.QueryRow(updateQ, amount, walletID).Scan(&id); err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("credit wallet: %w", err)
	}

	txn, err := insertWalletTransaction(tx, walletID, domain.TxCredit, amount, reference, description)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit credit: %w", err)
	}
	return txn, nil
}

// Debit removes funds from a wallet atomically, rejecting the operation when balance is insufficient.
func (s *Store) Debit(walletID string, amount float64, reference, description string) (*domain.WalletTransaction, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	// Lock row and check balance.
	const selectQ = `SELECT balance FROM wallets WHERE id = $1 FOR UPDATE`
	var current float64
	if err := tx.QueryRow(selectQ, walletID).Scan(&current); err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("lock wallet: %w", err)
	}
	if current < amount {
		return nil, domain.ErrInsufficientFunds
	}

	const updateQ = `
		UPDATE wallets
		SET balance = balance - $1, updated_at = NOW()
		WHERE id = $2`
	if _, err := tx.Exec(updateQ, amount, walletID); err != nil {
		return nil, fmt.Errorf("debit wallet: %w", err)
	}

	txn, err := insertWalletTransaction(tx, walletID, domain.TxDebit, amount, reference, description)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit debit: %w", err)
	}
	return txn, nil
}

// GetTransactions returns the most recent transactions for a wallet (descending order).
func (s *Store) GetTransactions(walletID string, limit int) ([]domain.WalletTransaction, error) {
	const q = `
		SELECT id, wallet_id, type, amount, reference, description, created_at
		FROM wallet_transactions
		WHERE wallet_id = $1
		ORDER BY created_at DESC
		LIMIT $2`

	rows, err := s.db.Query(q, walletID, limit)
	if err != nil {
		return nil, fmt.Errorf("query transactions: %w", err)
	}
	defer rows.Close()

	var txns []domain.WalletTransaction
	for rows.Next() {
		var t domain.WalletTransaction
		if err := rows.Scan(&t.ID, &t.WalletID, &t.Type, &t.Amount,
			&t.Reference, &t.Description, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan transaction: %w", err)
		}
		txns = append(txns, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return txns, nil
}

// insertWalletTransaction writes a wallet_transaction row inside an existing DB transaction.
func insertWalletTransaction(tx *sql.Tx, walletID string, txType domain.TxType,
	amount float64, reference, description string) (*domain.WalletTransaction, error) {

	const q = `
		INSERT INTO wallet_transactions
			(id, wallet_id, type, amount, reference, description, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, wallet_id, type, amount, reference, description, created_at`

	id := uuid.NewString()
	now := time.Now().UTC()
	row := tx.QueryRow(q, id, walletID, string(txType), amount, reference, description, now)

	var t domain.WalletTransaction
	if err := row.Scan(&t.ID, &t.WalletID, &t.Type, &t.Amount,
		&t.Reference, &t.Description, &t.CreatedAt); err != nil {
		return nil, fmt.Errorf("insert wallet transaction: %w", err)
	}
	return &t, nil
}
