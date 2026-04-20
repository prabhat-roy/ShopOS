package store

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/shopos/gift-card-service/domain"
)

// Store handles all Postgres interactions for the gift-card service.
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

// Issue inserts a new gift card into the database and returns it.
func (s *Store) Issue(card *domain.GiftCard) (*domain.GiftCard, error) {
	const q = `
		INSERT INTO gift_cards
			(id, code, initial_balance, current_balance, currency, issued_to, active, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
		RETURNING id, code, initial_balance, current_balance, currency, issued_to, active, expires_at, created_at, updated_at`

	row := s.db.QueryRow(q,
		card.ID, card.Code, card.InitialBalance, card.CurrentBalance,
		card.Currency, card.IssuedTo, card.Active, card.ExpiresAt,
	)
	return scanCard(row)
}

// GetByCode retrieves a gift card by its redemption code.
// Returns domain.ErrNotFound when no card matches.
func (s *Store) GetByCode(code string) (*domain.GiftCard, error) {
	const q = `
		SELECT id, code, initial_balance, current_balance, currency, issued_to, active, expires_at, created_at, updated_at
		FROM gift_cards
		WHERE code = $1`
	row := s.db.QueryRow(q, code)
	card, err := scanCard(row)
	if err == domain.ErrNotFound {
		return nil, domain.ErrNotFound
	}
	return card, err
}

// GetByID retrieves a gift card by its primary key.
func (s *Store) GetByID(id string) (*domain.GiftCard, error) {
	const q = `
		SELECT id, code, initial_balance, current_balance, currency, issued_to, active, expires_at, created_at, updated_at
		FROM gift_cards
		WHERE id = $1`
	row := s.db.QueryRow(q, id)
	card, err := scanCard(row)
	if err == domain.ErrNotFound {
		return nil, domain.ErrNotFound
	}
	return card, err
}

// Redeem deducts amount from the card's balance and writes a redemption record atomically.
// Returns domain.ErrNotFound, domain.ErrCardInactive, domain.ErrCardExpired, or domain.ErrInsufficientBalance on failure.
func (s *Store) Redeem(code, orderID string, amount float64) (*domain.RedemptionRecord, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	// Lock card row.
	const selectQ = `
		SELECT id, current_balance, active, expires_at
		FROM gift_cards
		WHERE code = $1
		FOR UPDATE`

	var (
		cardID  string
		balance float64
		active  bool
		expires *time.Time
	)
	err = tx.QueryRow(selectQ, code).Scan(&cardID, &balance, &active, &expires)
	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("lock card: %w", err)
	}
	if !active {
		return nil, domain.ErrCardInactive
	}
	if expires != nil && time.Now().UTC().After(*expires) {
		return nil, domain.ErrCardExpired
	}
	if balance < amount {
		return nil, domain.ErrInsufficientBalance
	}

	const updateQ = `UPDATE gift_cards SET current_balance = current_balance - $1, updated_at = NOW() WHERE id = $2`
	if _, err := tx.Exec(updateQ, amount, cardID); err != nil {
		return nil, fmt.Errorf("update balance: %w", err)
	}

	rec, err := insertRedemption(tx, cardID, orderID, amount)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit redeem: %w", err)
	}
	return rec, nil
}

// ListRedemptions returns all redemption records for a card.
func (s *Store) ListRedemptions(cardID string) ([]domain.RedemptionRecord, error) {
	const q = `
		SELECT id, card_id, order_id, amount, created_at
		FROM redemption_records
		WHERE card_id = $1
		ORDER BY created_at DESC`

	rows, err := s.db.Query(q, cardID)
	if err != nil {
		return nil, fmt.Errorf("query redemptions: %w", err)
	}
	defer rows.Close()

	var records []domain.RedemptionRecord
	for rows.Next() {
		var r domain.RedemptionRecord
		if err := rows.Scan(&r.ID, &r.CardID, &r.OrderID, &r.Amount, &r.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan redemption: %w", err)
		}
		records = append(records, r)
	}
	return records, rows.Err()
}

// Deactivate marks a gift card as inactive.
func (s *Store) Deactivate(id string) error {
	const q = `UPDATE gift_cards SET active = false, updated_at = NOW() WHERE id = $1`
	res, err := s.db.Exec(q, id)
	if err != nil {
		return fmt.Errorf("deactivate: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// scanCard scans a single gift card row from a *sql.Row.
func scanCard(row *sql.Row) (*domain.GiftCard, error) {
	var c domain.GiftCard
	err := row.Scan(
		&c.ID, &c.Code, &c.InitialBalance, &c.CurrentBalance,
		&c.Currency, &c.IssuedTo, &c.Active, &c.ExpiresAt,
		&c.CreatedAt, &c.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan card: %w", err)
	}
	return &c, nil
}

// insertRedemption writes a redemption_record row inside an existing DB transaction.
func insertRedemption(tx *sql.Tx, cardID, orderID string, amount float64) (*domain.RedemptionRecord, error) {
	const q = `
		INSERT INTO redemption_records (id, card_id, order_id, amount, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, card_id, order_id, amount, created_at`

	id := uuid.NewString()
	now := time.Now().UTC()
	row := tx.QueryRow(q, id, cardID, orderID, amount, now)

	var r domain.RedemptionRecord
	if err := row.Scan(&r.ID, &r.CardID, &r.OrderID, &r.Amount, &r.CreatedAt); err != nil {
		return nil, fmt.Errorf("insert redemption: %w", err)
	}
	return &r, nil
}
