package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// CreditLimitStatus represents the lifecycle state of an org's credit account.
type CreditLimitStatus string

const (
	CreditLimitStatusActive      CreditLimitStatus = "ACTIVE"
	CreditLimitStatusSuspended   CreditLimitStatus = "SUSPENDED"
	CreditLimitStatusUnderReview CreditLimitStatus = "UNDER_REVIEW"
)

// TransactionType categorises credit movements.
type TransactionType string

const (
	TransactionTypeUtilization TransactionType = "utilization"
	TransactionTypePayment     TransactionType = "payment"
	TransactionTypeAdjustment  TransactionType = "adjustment"
)

// Sentinel errors.
var (
	ErrNotFound           = errors.New("credit limit not found")
	ErrInsufficientCredit = errors.New("insufficient available credit")
	ErrAccountSuspended   = errors.New("credit account is suspended")
)

// OrgCreditLimit is the aggregate root for an organisation's credit account.
type OrgCreditLimit struct {
	ID             uuid.UUID         `json:"id"`
	OrgID          uuid.UUID         `json:"org_id"`
	CreditLimit    float64           `json:"credit_limit"`
	UsedCredit     float64           `json:"used_credit"`
	AvailableCredit float64          `json:"available_credit"`
	Currency       string            `json:"currency"`
	Status         CreditLimitStatus `json:"status"`
	RiskScore      int               `json:"risk_score"` // 0–100
	LastReviewedAt *time.Time        `json:"last_reviewed_at,omitempty"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

// CreditTransaction records a single debit or credit movement.
type CreditTransaction struct {
	ID        uuid.UUID       `json:"id"`
	OrgID     uuid.UUID       `json:"org_id"`
	Type      TransactionType `json:"type"`
	Amount    float64         `json:"amount"`
	Reference string          `json:"reference"`
	Balance   float64         `json:"balance"` // available credit after this transaction
	CreatedAt time.Time       `json:"created_at"`
}

// AvailabilityCheck is the response for a credit check query.
type AvailabilityCheck struct {
	Available       bool    `json:"available"`
	AvailableAmount float64 `json:"available_amount"`
	RequestedAmount float64 `json:"requested_amount"`
}
