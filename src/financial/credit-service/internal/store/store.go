package store

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/shopos/credit-service/internal/domain"
)

// Storer defines all persistence operations for credit accounts and transactions.
type Storer interface {
	CreateAccount(acc *domain.CreditAccount) error
	GetAccount(id uuid.UUID) (*domain.CreditAccount, error)
	GetByCustomerID(customerID uuid.UUID) (*domain.CreditAccount, error)
	UpdateCredit(accountID uuid.UUID, availableCredit, usedCredit float64) error
	SaveTransaction(tx *domain.CreditTransaction) error
	ListTransactions(accountID uuid.UUID, limit int) ([]domain.CreditTransaction, error)
	SuspendAccount(id uuid.UUID) error
	CloseAccount(id uuid.UUID) error
}

// PostgresStore implements Storer using a PostgreSQL database.
type PostgresStore struct {
	db *sql.DB
}

// New opens a PostgreSQL connection pool and returns a PostgresStore.
func New(databaseURL string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("store: opening db: %w", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)	return &PostgresStore{db: db}, nil
}

// Close releases all database resources.
func (s *PostgresStore) Close() error {
	return s.db.Close()
}

// CreateAccount inserts a new CreditAccount row.
func (s *PostgresStore) CreateAccount(acc *domain.CreditAccount) error {
	const q = `
		INSERT INTO credit_accounts
			(id, customer_id, credit_limit, available_credit, used_credit,
			 currency, status, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`

	_, err := s.db.Exec(q,
		acc.ID, acc.CustomerID, acc.CreditLimit, acc.AvailableCredit,
		acc.UsedCredit, acc.Currency, acc.Status, acc.CreatedAt, acc.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("store: CreateAccount: %w", err)
	}
	return nil
}

// GetAccount fetches a CreditAccount by primary key.
func (s *PostgresStore) GetAccount(id uuid.UUID) (*domain.CreditAccount, error) {
	const q = `
		SELECT id, customer_id, credit_limit, available_credit, used_credit,
		       currency, status, created_at, updated_at
		FROM credit_accounts WHERE id = $1`

	acc := &domain.CreditAccount{}
	err := s.db.QueryRow(q, id).Scan(
		&acc.ID, &acc.CustomerID, &acc.CreditLimit, &acc.AvailableCredit,
		&acc.UsedCredit, &acc.Currency, &acc.Status, &acc.CreatedAt, &acc.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("store: GetAccount: %w", err)
	}
	return acc, nil
}

// GetByCustomerID fetches the CreditAccount for a given customer.
func (s *PostgresStore) GetByCustomerID(customerID uuid.UUID) (*domain.CreditAccount, error) {
	const q = `
		SELECT id, customer_id, credit_limit, available_credit, used_credit,
		       currency, status, created_at, updated_at
		FROM credit_accounts WHERE customer_id = $1`

	acc := &domain.CreditAccount{}
	err := s.db.QueryRow(q, customerID).Scan(
		&acc.ID, &acc.CustomerID, &acc.CreditLimit, &acc.AvailableCredit,
		&acc.UsedCredit, &acc.Currency, &acc.Status, &acc.CreatedAt, &acc.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("store: GetByCustomerID: %w", err)
	}
	return acc, nil
}

// UpdateCredit modifies the available_credit and used_credit columns atomically.
func (s *PostgresStore) UpdateCredit(accountID uuid.UUID, availableCredit, usedCredit float64) error {
	const q = `
		UPDATE credit_accounts
		SET available_credit = $1, used_credit = $2, updated_at = $3
		WHERE id = $4`

	res, err := s.db.Exec(q, availableCredit, usedCredit, time.Now().UTC(), accountID)
	if err != nil {
		return fmt.Errorf("store: UpdateCredit: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// SaveTransaction inserts a CreditTransaction record.
func (s *PostgresStore) SaveTransaction(tx *domain.CreditTransaction) error {
	const q = `
		INSERT INTO credit_transactions
			(id, account_id, type, amount, reference, description, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`

	_, err := s.db.Exec(q,
		tx.ID, tx.AccountID, tx.Type, tx.Amount, tx.Reference,
		tx.Description, tx.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("store: SaveTransaction: %w", err)
	}
	return nil
}

// ListTransactions returns up to limit transactions for an account ordered
// newest-first.
func (s *PostgresStore) ListTransactions(accountID uuid.UUID, limit int) ([]domain.CreditTransaction, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	const q = `
		SELECT id, account_id, type, amount, reference, description, created_at
		FROM credit_transactions
		WHERE account_id = $1
		ORDER BY created_at DESC
		LIMIT $2`

	rows, err := s.db.Query(q, accountID, limit)
	if err != nil {
		return nil, fmt.Errorf("store: ListTransactions: %w", err)
	}
	defer rows.Close()

	var txs []domain.CreditTransaction
	for rows.Next() {
		var tx domain.CreditTransaction
		if err := rows.Scan(
			&tx.ID, &tx.AccountID, &tx.Type, &tx.Amount,
			&tx.Reference, &tx.Description, &tx.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("store: ListTransactions scan: %w", err)
		}
		txs = append(txs, tx)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: ListTransactions rows: %w", err)
	}
	return txs, nil
}

// SuspendAccount sets an account's status to suspended.
func (s *PostgresStore) SuspendAccount(id uuid.UUID) error {
	return s.setStatus(id, domain.StatusSuspended)
}

// CloseAccount sets an account's status to closed.
func (s *PostgresStore) CloseAccount(id uuid.UUID) error {
	return s.setStatus(id, domain.StatusClosed)
}

func (s *PostgresStore) setStatus(id uuid.UUID, status domain.AccountStatus) error {
	const q = `UPDATE credit_accounts SET status = $1, updated_at = $2 WHERE id = $3`
	res, err := s.db.Exec(q, status, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("store: setStatus: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}
