package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// ErrNotFound is returned when a requested expense does not exist.
var ErrNotFound = errors.New("expense not found")

// ErrInvalidTransition is returned when a status change is not permitted.
var ErrInvalidTransition = errors.New("invalid status transition")

// ExpenseStatus represents the lifecycle state of an expense.
type ExpenseStatus string

const (
	StatusDraft       ExpenseStatus = "DRAFT"
	StatusSubmitted   ExpenseStatus = "SUBMITTED"
	StatusApproved    ExpenseStatus = "APPROVED"
	StatusRejected    ExpenseStatus = "REJECTED"
	StatusReimbursed  ExpenseStatus = "REIMBURSED"
)

// ExpenseCategory classifies the nature of an expense.
type ExpenseCategory string

const (
	CategoryTravel    ExpenseCategory = "TRAVEL"
	CategoryMeals     ExpenseCategory = "MEALS"
	CategorySoftware  ExpenseCategory = "SOFTWARE"
	CategoryHardware  ExpenseCategory = "HARDWARE"
	CategoryOffice    ExpenseCategory = "OFFICE"
	CategoryMarketing ExpenseCategory = "MARKETING"
	CategoryOther     ExpenseCategory = "OTHER"
)

// Expense is the core domain entity representing an employee expense claim.
type Expense struct {
	ID           uuid.UUID       `json:"id"`
	EmployeeID   uuid.UUID       `json:"employeeId"`
	Category     ExpenseCategory `json:"category"`
	Amount       float64         `json:"amount"`
	Currency     string          `json:"currency"`
	Description  string          `json:"description"`
	ReceiptURL   string          `json:"receiptUrl"`
	Status       ExpenseStatus   `json:"status"`
	ApprovedBy   *uuid.UUID      `json:"approvedBy,omitempty"`
	ApprovedAt   *time.Time      `json:"approvedAt,omitempty"`
	ReimbursedAt *time.Time      `json:"reimbursedAt,omitempty"`
	Notes        string          `json:"notes,omitempty"`
	CreatedAt    time.Time       `json:"createdAt"`
	UpdatedAt    time.Time       `json:"updatedAt"`
}

// ListFilter holds optional filter parameters for listing expenses.
type ListFilter struct {
	EmployeeID string
	Status     string
	Category   string
}

// ValidCategory returns true when c is a recognised ExpenseCategory.
func ValidCategory(c ExpenseCategory) bool {
	switch c {
	case CategoryTravel, CategoryMeals, CategorySoftware, CategoryHardware,
		CategoryOffice, CategoryMarketing, CategoryOther:
		return true
	}
	return false
}

// ValidStatus returns true when s is a recognised ExpenseStatus.
func ValidStatus(s ExpenseStatus) bool {
	switch s {
	case StatusDraft, StatusSubmitted, StatusApproved, StatusRejected, StatusReimbursed:
		return true
	}
	return false
}
