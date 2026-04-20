package domain

import (
	"errors"
	"time"
)

// TransactionType classifies a point movement.
type TransactionType string

const (
	TxEarn     TransactionType = "EARN"
	TxRedeem   TransactionType = "REDEEM"
	TxExpire   TransactionType = "EXPIRE"
	TxAdjust   TransactionType = "ADJUST"
)

// LoyaltyAccount holds the current state for a customer's loyalty balance.
type LoyaltyAccount struct {
	CustomerID string
	Points     int64
	TierName   string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// PointTransaction is an immutable ledger entry for a point movement.
type PointTransaction struct {
	ID          string
	CustomerID  string
	Type        TransactionType
	Points      int64
	Balance     int64
	OrderID     string
	Description string
	CreatedAt   time.Time
}

var (
	ErrInsufficientPoints = errors.New("insufficient points")
	ErrNotFound           = errors.New("not found")
)
