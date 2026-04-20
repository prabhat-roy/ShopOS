package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Sentinel errors.
var (
	ErrNotFound      = errors.New("budget: record not found")
	ErrBudgetExceeded = errors.New("budget: spending would exceed budget")
)

// BudgetPeriod classifies the time horizon of a Budget.
type BudgetPeriod string

const (
	PeriodMonthly   BudgetPeriod = "MONTHLY"
	PeriodQuarterly BudgetPeriod = "QUARTERLY"
	PeriodAnnual    BudgetPeriod = "ANNUAL"
)

// BudgetStatus represents the lifecycle state of a Budget.
type BudgetStatus string

const (
	StatusDraft  BudgetStatus = "DRAFT"
	StatusActive BudgetStatus = "ACTIVE"
	StatusClosed BudgetStatus = "CLOSED"
)

// Budget represents a departmental spending plan for a given fiscal period.
// remainingAmount = totalAmount - spentAmount at all times.
// allocatedAmount is the sum of all BudgetAllocation.allocatedAmount rows.
type Budget struct {
	ID               uuid.UUID    `json:"id"`
	Department       string       `json:"department"`
	Name             string       `json:"name"`
	Period           BudgetPeriod `json:"period"`
	FiscalYear       int          `json:"fiscal_year"`
	StartDate        time.Time    `json:"start_date"`
	EndDate          time.Time    `json:"end_date"`
	TotalAmount      float64      `json:"total_amount"`
	AllocatedAmount  float64      `json:"allocated_amount"`
	SpentAmount      float64      `json:"spent_amount"`
	RemainingAmount  float64      `json:"remaining_amount"`
	Currency         string       `json:"currency"`
	Status           BudgetStatus `json:"status"`
	CreatedAt        time.Time    `json:"created_at"`
	UpdatedAt        time.Time    `json:"updated_at"`
}

// BudgetAllocation is a named category slice of a Budget's total amount.
type BudgetAllocation struct {
	ID              uuid.UUID `json:"id"`
	BudgetID        uuid.UUID `json:"budget_id"`
	Category        string    `json:"category"`
	AllocatedAmount float64   `json:"allocated_amount"`
	SpentAmount     float64   `json:"spent_amount"`
	Notes           string    `json:"notes"`
}

// SpendingRecord documents an individual expense against a Budget (and
// optionally a specific BudgetAllocation).
type SpendingRecord struct {
	ID           uuid.UUID  `json:"id"`
	BudgetID     uuid.UUID  `json:"budget_id"`
	AllocationID *uuid.UUID `json:"allocation_id,omitempty"`
	Category     string     `json:"category"`
	Description  string     `json:"description"`
	Amount       float64    `json:"amount"`
	Reference    string     `json:"reference"`
	CreatedAt    time.Time  `json:"created_at"`
}

// BudgetSummary is a read-only view returned by GetBudgetSummary.
type BudgetSummary struct {
	Budget      *Budget            `json:"budget"`
	Allocations []BudgetAllocation `json:"allocations"`
	Utilization float64            `json:"utilization_pct"` // spent / total * 100
}
