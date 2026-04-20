package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopos/budget-service/internal/domain"
	"github.com/shopos/budget-service/internal/store"
)

// Servicer defines all business-logic operations for budget management.
type Servicer interface {
	CreateBudget(department, name string, period domain.BudgetPeriod, fiscalYear int, startDate, endDate time.Time, totalAmount float64, currency string) (*domain.Budget, error)
	GetBudget(id uuid.UUID) (*domain.Budget, error)
	ListBudgets(department string, status domain.BudgetStatus, fiscalYear int) ([]domain.Budget, error)
	ActivateBudget(id uuid.UUID) error
	CloseBudget(id uuid.UUID) error
	CreateAllocation(budgetID uuid.UUID, category string, amount float64, notes string) (*domain.BudgetAllocation, error)
	ListAllocations(budgetID uuid.UUID) ([]domain.BudgetAllocation, error)
	RecordSpending(budgetID uuid.UUID, allocationID *uuid.UUID, category, description string, amount float64, reference string) (*domain.SpendingRecord, error)
	ListSpending(budgetID uuid.UUID, limit int) ([]domain.SpendingRecord, error)
	GetBudgetSummary(id uuid.UUID) (*domain.BudgetSummary, error)
}

// Service implements Servicer.
type Service struct {
	store store.Storer
}

// New creates a new Service with the provided Storer.
func New(st store.Storer) *Service {
	return &Service{store: st}
}

// CreateBudget creates a new budget in DRAFT status.
func (s *Service) CreateBudget(department, name string, period domain.BudgetPeriod, fiscalYear int, startDate, endDate time.Time, totalAmount float64, currency string) (*domain.Budget, error) {
	if totalAmount < 0 {
		return nil, fmt.Errorf("service: total_amount must be non-negative")
	}
	if endDate.Before(startDate) {
		return nil, fmt.Errorf("service: end_date must be after start_date")
	}
	if currency == "" {
		currency = "USD"
	}
	now := time.Now().UTC()
	b := &domain.Budget{
		ID:              uuid.New(),
		Department:      department,
		Name:            name,
		Period:          period,
		FiscalYear:      fiscalYear,
		StartDate:       startDate.UTC(),
		EndDate:         endDate.UTC(),
		TotalAmount:     totalAmount,
		AllocatedAmount: 0,
		SpentAmount:     0,
		RemainingAmount: totalAmount,
		Currency:        currency,
		Status:          domain.StatusDraft,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := s.store.CreateBudget(b); err != nil {
		return nil, fmt.Errorf("service: CreateBudget: %w", err)
	}
	return b, nil
}

// GetBudget returns a budget by its UUID.
func (s *Service) GetBudget(id uuid.UUID) (*domain.Budget, error) {
	b, err := s.store.GetBudget(id)
	if err != nil {
		return nil, fmt.Errorf("service: GetBudget: %w", err)
	}
	return b, nil
}

// ListBudgets returns budgets filtered by the supplied criteria.
// Zero-value fields are treated as "no filter".
func (s *Service) ListBudgets(department string, status domain.BudgetStatus, fiscalYear int) ([]domain.Budget, error) {
	budgets, err := s.store.ListBudgets(store.ListBudgetsFilter{
		Department: department,
		Status:     status,
		FiscalYear: fiscalYear,
	})
	if err != nil {
		return nil, fmt.Errorf("service: ListBudgets: %w", err)
	}
	if budgets == nil {
		budgets = []domain.Budget{}
	}
	return budgets, nil
}

// ActivateBudget transitions a budget from DRAFT to ACTIVE.
func (s *Service) ActivateBudget(id uuid.UUID) error {
	b, err := s.store.GetBudget(id)
	if err != nil {
		return fmt.Errorf("service: ActivateBudget: %w", err)
	}
	if b.Status != domain.StatusDraft {
		return fmt.Errorf("service: ActivateBudget: budget must be in DRAFT status (current: %s)", b.Status)
	}
	if err := s.store.UpdateBudgetStatus(id, domain.StatusActive); err != nil {
		return fmt.Errorf("service: ActivateBudget update: %w", err)
	}
	return nil
}

// CloseBudget transitions a budget from ACTIVE to CLOSED.
func (s *Service) CloseBudget(id uuid.UUID) error {
	b, err := s.store.GetBudget(id)
	if err != nil {
		return fmt.Errorf("service: CloseBudget: %w", err)
	}
	if b.Status != domain.StatusActive {
		return fmt.Errorf("service: CloseBudget: budget must be in ACTIVE status (current: %s)", b.Status)
	}
	if err := s.store.UpdateBudgetStatus(id, domain.StatusClosed); err != nil {
		return fmt.Errorf("service: CloseBudget update: %w", err)
	}
	return nil
}

// CreateAllocation adds a named category allocation to a budget.
// The sum of all allocations cannot exceed the budget's total amount.
func (s *Service) CreateAllocation(budgetID uuid.UUID, category string, amount float64, notes string) (*domain.BudgetAllocation, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("service: allocation amount must be positive")
	}

	b, err := s.store.GetBudget(budgetID)
	if err != nil {
		return nil, fmt.Errorf("service: CreateAllocation: %w", err)
	}
	if b.Status == domain.StatusClosed {
		return nil, fmt.Errorf("service: CreateAllocation: cannot allocate to a closed budget")
	}

	newAllocated := b.AllocatedAmount + amount
	if newAllocated > b.TotalAmount {
		return nil, fmt.Errorf("service: CreateAllocation: allocation would exceed total budget (%.2f > %.2f)", newAllocated, b.TotalAmount)
	}

	alloc := &domain.BudgetAllocation{
		ID:              uuid.New(),
		BudgetID:        budgetID,
		Category:        category,
		AllocatedAmount: amount,
		SpentAmount:     0,
		Notes:           notes,
	}
	if err := s.store.CreateAllocation(alloc); err != nil {
		return nil, fmt.Errorf("service: CreateAllocation save: %w", err)
	}

	if err := s.store.UpdateBudgetAllocated(budgetID, newAllocated); err != nil {
		return nil, fmt.Errorf("service: CreateAllocation update allocated: %w", err)
	}
	return alloc, nil
}

// ListAllocations returns all allocations for a budget.
func (s *Service) ListAllocations(budgetID uuid.UUID) ([]domain.BudgetAllocation, error) {
	allocs, err := s.store.ListAllocations(budgetID)
	if err != nil {
		return nil, fmt.Errorf("service: ListAllocations: %w", err)
	}
	if allocs == nil {
		allocs = []domain.BudgetAllocation{}
	}
	return allocs, nil
}

// RecordSpending records an expense against a budget and optional allocation.
// Returns ErrBudgetExceeded if the expense would exceed the remaining budget
// or the allocation's remaining capacity.
func (s *Service) RecordSpending(budgetID uuid.UUID, allocationID *uuid.UUID, category, description string, amount float64, reference string) (*domain.SpendingRecord, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("service: spending amount must be positive")
	}

	b, err := s.store.GetBudget(budgetID)
	if err != nil {
		return nil, fmt.Errorf("service: RecordSpending: %w", err)
	}
	if b.Status != domain.StatusActive {
		return nil, fmt.Errorf("service: RecordSpending: budget is not active (status: %s)", b.Status)
	}
	if amount > b.RemainingAmount {
		return nil, domain.ErrBudgetExceeded
	}

	// Validate against allocation if provided.
	if allocationID != nil {
		alloc, err := s.store.GetAllocation(*allocationID)
		if err != nil {
			return nil, fmt.Errorf("service: RecordSpending: allocation not found: %w", err)
		}
		if alloc.BudgetID != budgetID {
			return nil, fmt.Errorf("service: RecordSpending: allocation does not belong to this budget")
		}
		allocRemaining := alloc.AllocatedAmount - alloc.SpentAmount
		if amount > allocRemaining {
			return nil, fmt.Errorf("service: RecordSpending: amount %.2f exceeds allocation remaining %.2f", amount, allocRemaining)
		}
		newAllocSpent := alloc.SpentAmount + amount
		if err := s.store.UpdateAllocationSpending(*allocationID, newAllocSpent); err != nil {
			return nil, fmt.Errorf("service: RecordSpending update alloc: %w", err)
		}
	}

	newSpent := b.SpentAmount + amount
	newRemaining := b.RemainingAmount - amount
	if err := s.store.UpdateBudgetSpending(budgetID, newSpent, newRemaining); err != nil {
		return nil, fmt.Errorf("service: RecordSpending update budget: %w", err)
	}

	sr := &domain.SpendingRecord{
		ID:           uuid.New(),
		BudgetID:     budgetID,
		AllocationID: allocationID,
		Category:     category,
		Description:  description,
		Amount:       amount,
		Reference:    reference,
		CreatedAt:    time.Now().UTC(),
	}
	if err := s.store.RecordSpending(sr); err != nil {
		return nil, fmt.Errorf("service: RecordSpending save: %w", err)
	}
	return sr, nil
}

// ListSpending returns up to limit spending records for a budget.
func (s *Service) ListSpending(budgetID uuid.UUID, limit int) ([]domain.SpendingRecord, error) {
	records, err := s.store.ListSpending(budgetID, limit)
	if err != nil {
		return nil, fmt.Errorf("service: ListSpending: %w", err)
	}
	if records == nil {
		records = []domain.SpendingRecord{}
	}
	return records, nil
}

// GetBudgetSummary returns the budget, its allocations, and utilization %.
func (s *Service) GetBudgetSummary(id uuid.UUID) (*domain.BudgetSummary, error) {
	b, err := s.store.GetBudget(id)
	if err != nil {
		return nil, fmt.Errorf("service: GetBudgetSummary: %w", err)
	}
	allocs, err := s.store.ListAllocations(id)
	if err != nil {
		return nil, fmt.Errorf("service: GetBudgetSummary allocations: %w", err)
	}
	if allocs == nil {
		allocs = []domain.BudgetAllocation{}
	}

	var utilization float64
	if b.TotalAmount > 0 {
		utilization = (b.SpentAmount / b.TotalAmount) * 100
	}

	return &domain.BudgetSummary{
		Budget:      b,
		Allocations: allocs,
		Utilization: utilization,
	}, nil
}
