package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopos/expense-management-service/internal/domain"
	"github.com/shopos/expense-management-service/internal/store"
)

// Servicer defines the business-logic contract for expense management.
type Servicer interface {
	CreateExpense(e *domain.Expense) (*domain.Expense, error)
	GetExpense(id uuid.UUID) (*domain.Expense, error)
	ListExpenses(f domain.ListFilter) ([]*domain.Expense, error)
	SubmitExpense(id uuid.UUID) (*domain.Expense, error)
	ApproveExpense(id uuid.UUID, approverID uuid.UUID) (*domain.Expense, error)
	RejectExpense(id uuid.UUID, reason string) (*domain.Expense, error)
	ReimburseExpense(id uuid.UUID) (*domain.Expense, error)
	DeleteExpense(id uuid.UUID) error
}

// Service is the concrete implementation of Servicer.
type Service struct {
	store store.Storer
}

// New creates a new Service backed by the given Storer.
func New(s store.Storer) *Service {
	return &Service{store: s}
}

// CreateExpense validates and persists a new expense in DRAFT status.
func (svc *Service) CreateExpense(e *domain.Expense) (*domain.Expense, error) {
	if e.EmployeeID == uuid.Nil {
		return nil, fmt.Errorf("employeeId is required")
	}
	if !domain.ValidCategory(e.Category) {
		return nil, fmt.Errorf("category must be one of TRAVEL, MEALS, SOFTWARE, HARDWARE, OFFICE, MARKETING, OTHER")
	}
	if e.Amount <= 0 {
		return nil, fmt.Errorf("amount must be greater than zero")
	}
	if e.Currency == "" {
		return nil, fmt.Errorf("currency is required")
	}
	if e.Description == "" {
		return nil, fmt.Errorf("description is required")
	}

	now := time.Now().UTC()
	e.ID = uuid.New()
	e.Status = domain.StatusDraft
	e.CreatedAt = now
	e.UpdatedAt = now

	if err := svc.store.Create(e); err != nil {
		return nil, fmt.Errorf("service: create expense: %w", err)
	}
	return e, nil
}

// GetExpense retrieves a single expense by its ID.
func (svc *Service) GetExpense(id uuid.UUID) (*domain.Expense, error) {
	return svc.store.Get(id)
}

// ListExpenses returns expenses filtered by optional criteria.
func (svc *Service) ListExpenses(f domain.ListFilter) ([]*domain.Expense, error) {
	return svc.store.List(f)
}

// SubmitExpense transitions a DRAFT expense to SUBMITTED.
func (svc *Service) SubmitExpense(id uuid.UUID) (*domain.Expense, error) {
	e, err := svc.store.Get(id)
	if err != nil {
		return nil, err
	}
	if e.Status != domain.StatusDraft {
		return nil, fmt.Errorf("%w: cannot submit expense in status %s", domain.ErrInvalidTransition, e.Status)
	}
	now := time.Now().UTC()
	if err := svc.store.UpdateStatus(id, domain.StatusSubmitted, now); err != nil {
		return nil, fmt.Errorf("service: submit expense: %w", err)
	}
	e.Status = domain.StatusSubmitted
	e.UpdatedAt = now
	return e, nil
}

// ApproveExpense transitions a SUBMITTED expense to APPROVED and records the approver.
func (svc *Service) ApproveExpense(id uuid.UUID, approverID uuid.UUID) (*domain.Expense, error) {
	if approverID == uuid.Nil {
		return nil, fmt.Errorf("approverId is required")
	}
	e, err := svc.store.Get(id)
	if err != nil {
		return nil, err
	}
	if e.Status != domain.StatusSubmitted {
		return nil, fmt.Errorf("%w: cannot approve expense in status %s", domain.ErrInvalidTransition, e.Status)
	}
	now := time.Now().UTC()
	if err := svc.store.UpdateStatus(id, domain.StatusApproved, now); err != nil {
		return nil, fmt.Errorf("service: approve expense (status): %w", err)
	}
	if err := svc.store.SetApprover(id, approverID, now, now); err != nil {
		return nil, fmt.Errorf("service: approve expense (approver): %w", err)
	}
	e.Status = domain.StatusApproved
	e.ApprovedBy = &approverID
	e.ApprovedAt = &now
	e.UpdatedAt = now
	return e, nil
}

// RejectExpense transitions a SUBMITTED expense to REJECTED and stores the reason in Notes.
func (svc *Service) RejectExpense(id uuid.UUID, reason string) (*domain.Expense, error) {
	e, err := svc.store.Get(id)
	if err != nil {
		return nil, err
	}
	if e.Status != domain.StatusSubmitted {
		return nil, fmt.Errorf("%w: cannot reject expense in status %s", domain.ErrInvalidTransition, e.Status)
	}
	now := time.Now().UTC()
	if err := svc.store.UpdateStatus(id, domain.StatusRejected, now); err != nil {
		return nil, fmt.Errorf("service: reject expense (status): %w", err)
	}
	if reason != "" {
		if err := svc.store.SetNotes(id, reason, now); err != nil {
			return nil, fmt.Errorf("service: reject expense (notes): %w", err)
		}
	}
	e.Status = domain.StatusRejected
	e.Notes = reason
	e.UpdatedAt = now
	return e, nil
}

// ReimburseExpense transitions an APPROVED expense to REIMBURSED.
func (svc *Service) ReimburseExpense(id uuid.UUID) (*domain.Expense, error) {
	e, err := svc.store.Get(id)
	if err != nil {
		return nil, err
	}
	if e.Status != domain.StatusApproved {
		return nil, fmt.Errorf("%w: cannot reimburse expense in status %s", domain.ErrInvalidTransition, e.Status)
	}
	now := time.Now().UTC()
	if err := svc.store.UpdateStatus(id, domain.StatusReimbursed, now); err != nil {
		return nil, fmt.Errorf("service: reimburse expense (status): %w", err)
	}
	if err := svc.store.SetReimbursed(id, now, now); err != nil {
		return nil, fmt.Errorf("service: reimburse expense (timestamp): %w", err)
	}
	e.Status = domain.StatusReimbursed
	e.ReimbursedAt = &now
	e.UpdatedAt = now
	return e, nil
}

// DeleteExpense permanently removes a DRAFT expense. Other statuses are rejected.
func (svc *Service) DeleteExpense(id uuid.UUID) error {
	e, err := svc.store.Get(id)
	if err != nil {
		return err
	}
	if e.Status != domain.StatusDraft {
		return fmt.Errorf("%w: only DRAFT expenses can be deleted, current status is %s", domain.ErrInvalidTransition, e.Status)
	}
	if err := svc.store.Delete(id); err != nil {
		return fmt.Errorf("service: delete expense: %w", err)
	}
	return nil
}
