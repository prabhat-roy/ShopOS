package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Sentinel errors returned by service operations.
var (
	ErrNotFound           = errors.New("credit: record not found")
	ErrInsufficientCredit = errors.New("credit: insufficient available credit")
	ErrAccountInactive    = errors.New("credit: account is not active")
)

// AccountStatus represents the lifecycle state of a CreditAccount.
type AccountStatus string

const (
	StatusActive    AccountStatus = "active"
	StatusSuspended AccountStatus = "suspended"
	StatusClosed    AccountStatus = "closed"
)

// TransactionType classifies a CreditTransaction.
type TransactionType string

const (
	TxCharge     TransactionType = "charge"
	TxPayment    TransactionType = "payment"
	TxAdjustment TransactionType = "adjustment"
)

// CreditAccount holds the credit line details for a single customer.
// availableCredit + usedCredit == creditLimit at all times.
type CreditAccount struct {
	ID               uuid.UUID     `json:"id"`
	CustomerID       uuid.UUID     `json:"customer_id"`
	CreditLimit      float64       `json:"credit_limit"`
	AvailableCredit  float64       `json:"available_credit"`
	UsedCredit       float64       `json:"used_credit"`
	Currency         string        `json:"currency"`
	Status           AccountStatus `json:"status"`
	CreatedAt        time.Time     `json:"created_at"`
	UpdatedAt        time.Time     `json:"updated_at"`
}

// CreditTransaction records a single debit, credit, or adjustment event
// against a CreditAccount.
type CreditTransaction struct {
	ID          uuid.UUID       `json:"id"`
	AccountID   uuid.UUID       `json:"account_id"`
	Type        TransactionType `json:"type"`
	Amount      float64         `json:"amount"`
	Reference   string          `json:"reference"`
	Description string          `json:"description"`
	CreatedAt   time.Time       `json:"created_at"`
}
