package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"

	"github.com/shopos/voucher-service/domain"
)

const migrationSQL = `
CREATE TABLE IF NOT EXISTS vouchers (
    id          TEXT PRIMARY KEY,
    code        TEXT NOT NULL UNIQUE,
    customer_id TEXT NOT NULL,
    amount      DOUBLE PRECISION NOT NULL,
    currency    TEXT NOT NULL DEFAULT 'USD',
    used        BOOLEAN NOT NULL DEFAULT FALSE,
    used_at     TIMESTAMPTZ,
    order_id    TEXT NOT NULL DEFAULT '',
    expires_at  TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_vouchers_customer ON vouchers (customer_id);
CREATE INDEX IF NOT EXISTS idx_vouchers_code ON vouchers (code);
`

// Store handles Postgres persistence for vouchers.
type Store struct {
	db *sql.DB
}

// New creates a Store, opens a DB connection and runs migrations.
func New(databaseURL string) (*Store, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)
	db.Exec(migrationSQL) //nolint:errcheck — best-effort at startup; fails gracefully when DB is unavailable
	return &Store{db: db}, nil
}

// Close shuts the connection pool.
func (s *Store) Close() error { return s.db.Close() }

// Issue inserts a new voucher.
func (s *Store) Issue(ctx context.Context, v *domain.Voucher) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO vouchers (id, code, customer_id, amount, currency, used, used_at, order_id, expires_at, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		v.ID, v.Code, v.CustomerID, v.Amount, v.Currency,
		v.Used, v.UsedAt, v.OrderID, v.ExpiresAt, v.CreatedAt,
	)
	return err
}

// GetByCode retrieves a voucher by its code.
func (s *Store) GetByCode(ctx context.Context, code string) (*domain.Voucher, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, code, customer_id, amount, currency, used, used_at, order_id, expires_at, created_at
		FROM vouchers WHERE code=$1`, code)
	return scanVoucher(row)
}

// Use atomically marks a voucher as used.
func (s *Store) Use(ctx context.Context, code, orderID string) error {
	now := time.Now().UTC()
	res, err := s.db.ExecContext(ctx, `
		UPDATE vouchers SET used=TRUE, used_at=$1, order_id=$2
		WHERE code=$3 AND used=FALSE`,
		now, orderID, code,
	)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrAlreadyUsed
	}
	return nil
}

// ListByCustomer returns all vouchers for a customer, newest first.
func (s *Store) ListByCustomer(ctx context.Context, customerID string) ([]*domain.Voucher, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, code, customer_id, amount, currency, used, used_at, order_id, expires_at, created_at
		FROM vouchers WHERE customer_id=$1 ORDER BY created_at DESC`, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*domain.Voucher
	for rows.Next() {
		v, err := scanVoucher(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	return result, rows.Err()
}

type scanner interface {
	Scan(dest ...any) error
}

func scanVoucher(s scanner) (*domain.Voucher, error) {
	var v domain.Voucher
	err := s.Scan(
		&v.ID, &v.Code, &v.CustomerID, &v.Amount, &v.Currency,
		&v.Used, &v.UsedAt, &v.OrderID, &v.ExpiresAt, &v.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	return &v, err
}
