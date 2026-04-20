package domain

import (
	"errors"
	"time"
)

// TxType classifies a wallet movement.
type TxType string

const (
	TxCredit TxType = "CREDIT"
	TxDebit  TxType = "DEBIT"
)

// Wallet holds the current balance for a customer.
type Wallet struct {
	ID         string
	CustomerID string
	Balance    float64
	Currency   string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// WalletTransaction is an immutable ledger entry for a wallet movement.
type WalletTransaction struct {
	ID          string
	WalletID    string
	Type        TxType
	Amount      float64
	Reference   string
	Description string
	CreatedAt   time.Time
}

var (
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrNotFound          = errors.New("not found")
)
