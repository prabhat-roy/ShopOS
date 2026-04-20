package store

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/shopos/budget-service/internal/domain"
)

// ListBudgetsFilter carries optional filter criteria for ListBudgets.
type ListBudgetsFilter struct {
	Department string
	Status     domain.BudgetStatus
	FiscalYear int // 0 means no filter
}

// Storer defines all persistence operations for the budget domain.
type Storer interface {
	CreateBudget(b *domain.Budget) error
	GetBudget(id uuid.UUID) (*domain.Budget, error)
	ListBudgets(f ListBudgetsFilter) ([]domain.Budget, error)
	UpdateBudgetSpending(id uuid.UUID, spentAmount, remainingAmount float64) error
	UpdateBudgetStatus(id uuid.UUID, status domain.BudgetStatus) error
	UpdateBudgetAllocated(id uuid.UUID, allocatedAmount float64) error

	CreateAllocation(a *domain.BudgetAllocation) error
	GetAllocation(id uuid.UUID) (*domain.BudgetAllocation, error)
	ListAllocations(budgetID uuid.UUID) ([]domain.BudgetAllocation, error)
	UpdateAllocationSpending(id uuid.UUID, spentAmount float64) error

	RecordSpending(s *domain.SpendingRecord) error
	ListSpending(budgetID uuid.UUID, limit int) ([]domain.SpendingRecord, error)
}

// PostgresStore implements Storer on top of PostgreSQL.
type PostgresStore struct {
	db *sql.DB
}

// New opens a connection pool and returns a PostgresStore.
func New(databaseURL string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("store: opening db: %w", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)	return &PostgresStore{db: db}, nil
}

// Close releases the connection pool.
func (s *PostgresStore) Close() error { return s.db.Close() }

// ---- Budget -----------------------------------------------------------------

func (s *PostgresStore) CreateBudget(b *domain.Budget) error {
	const q = `
		INSERT INTO budgets
			(id, department, name, period, fiscal_year, start_date, end_date,
			 total_amount, allocated_amount, spent_amount, remaining_amount,
			 currency, status, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`

	_, err := s.db.Exec(q,
		b.ID, b.Department, b.Name, b.Period, b.FiscalYear,
		b.StartDate, b.EndDate, b.TotalAmount, b.AllocatedAmount,
		b.SpentAmount, b.RemainingAmount, b.Currency, b.Status,
		b.CreatedAt, b.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("store: CreateBudget: %w", err)
	}
	return nil
}

func (s *PostgresStore) GetBudget(id uuid.UUID) (*domain.Budget, error) {
	const q = `
		SELECT id, department, name, period, fiscal_year, start_date, end_date,
		       total_amount, allocated_amount, spent_amount, remaining_amount,
		       currency, status, created_at, updated_at
		FROM budgets WHERE id = $1`

	b := &domain.Budget{}
	err := s.db.QueryRow(q, id).Scan(
		&b.ID, &b.Department, &b.Name, &b.Period, &b.FiscalYear,
		&b.StartDate, &b.EndDate, &b.TotalAmount, &b.AllocatedAmount,
		&b.SpentAmount, &b.RemainingAmount, &b.Currency, &b.Status,
		&b.CreatedAt, &b.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("store: GetBudget: %w", err)
	}
	return b, nil
}

func (s *PostgresStore) ListBudgets(f ListBudgetsFilter) ([]domain.Budget, error) {
	q := `
		SELECT id, department, name, period, fiscal_year, start_date, end_date,
		       total_amount, allocated_amount, spent_amount, remaining_amount,
		       currency, status, created_at, updated_at
		FROM budgets WHERE 1=1`

	args := []any{}
	n := 1
	if f.Department != "" {
		q += fmt.Sprintf(" AND department = $%d", n)
		args = append(args, f.Department)
		n++
	}
	if f.Status != "" {
		q += fmt.Sprintf(" AND status = $%d", n)
		args = append(args, f.Status)
		n++
	}
	if f.FiscalYear > 0 {
		q += fmt.Sprintf(" AND fiscal_year = $%d", n)
		args = append(args, f.FiscalYear)
		n++
	}
	q += " ORDER BY created_at DESC"

	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil, fmt.Errorf("store: ListBudgets: %w", err)
	}
	defer rows.Close()

	var budgets []domain.Budget
	for rows.Next() {
		var b domain.Budget
		if err := rows.Scan(
			&b.ID, &b.Department, &b.Name, &b.Period, &b.FiscalYear,
			&b.StartDate, &b.EndDate, &b.TotalAmount, &b.AllocatedAmount,
			&b.SpentAmount, &b.RemainingAmount, &b.Currency, &b.Status,
			&b.CreatedAt, &b.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("store: ListBudgets scan: %w", err)
		}
		budgets = append(budgets, b)
	}
	return budgets, rows.Err()
}

func (s *PostgresStore) UpdateBudgetSpending(id uuid.UUID, spentAmount, remainingAmount float64) error {
	const q = `UPDATE budgets SET spent_amount=$1, remaining_amount=$2, updated_at=$3 WHERE id=$4`
	res, err := s.db.Exec(q, spentAmount, remainingAmount, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("store: UpdateBudgetSpending: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (s *PostgresStore) UpdateBudgetStatus(id uuid.UUID, status domain.BudgetStatus) error {
	const q = `UPDATE budgets SET status=$1, updated_at=$2 WHERE id=$3`
	res, err := s.db.Exec(q, status, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("store: UpdateBudgetStatus: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (s *PostgresStore) UpdateBudgetAllocated(id uuid.UUID, allocatedAmount float64) error {
	const q = `UPDATE budgets SET allocated_amount=$1, updated_at=$2 WHERE id=$3`
	res, err := s.db.Exec(q, allocatedAmount, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("store: UpdateBudgetAllocated: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// ---- BudgetAllocation -------------------------------------------------------

func (s *PostgresStore) CreateAllocation(a *domain.BudgetAllocation) error {
	const q = `
		INSERT INTO budget_allocations
			(id, budget_id, category, allocated_amount, spent_amount, notes)
		VALUES ($1,$2,$3,$4,$5,$6)`
	_, err := s.db.Exec(q, a.ID, a.BudgetID, a.Category, a.AllocatedAmount, a.SpentAmount, a.Notes)
	if err != nil {
		return fmt.Errorf("store: CreateAllocation: %w", err)
	}
	return nil
}

func (s *PostgresStore) GetAllocation(id uuid.UUID) (*domain.BudgetAllocation, error) {
	const q = `
		SELECT id, budget_id, category, allocated_amount, spent_amount, notes
		FROM budget_allocations WHERE id = $1`
	a := &domain.BudgetAllocation{}
	err := s.db.QueryRow(q, id).Scan(&a.ID, &a.BudgetID, &a.Category, &a.AllocatedAmount, &a.SpentAmount, &a.Notes)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("store: GetAllocation: %w", err)
	}
	return a, nil
}

func (s *PostgresStore) ListAllocations(budgetID uuid.UUID) ([]domain.BudgetAllocation, error) {
	const q = `
		SELECT id, budget_id, category, allocated_amount, spent_amount, notes
		FROM budget_allocations WHERE budget_id = $1 ORDER BY category`

	rows, err := s.db.Query(q, budgetID)
	if err != nil {
		return nil, fmt.Errorf("store: ListAllocations: %w", err)
	}
	defer rows.Close()

	var allocs []domain.BudgetAllocation
	for rows.Next() {
		var a domain.BudgetAllocation
		if err := rows.Scan(&a.ID, &a.BudgetID, &a.Category, &a.AllocatedAmount, &a.SpentAmount, &a.Notes); err != nil {
			return nil, fmt.Errorf("store: ListAllocations scan: %w", err)
		}
		allocs = append(allocs, a)
	}
	return allocs, rows.Err()
}

func (s *PostgresStore) UpdateAllocationSpending(id uuid.UUID, spentAmount float64) error {
	const q = `UPDATE budget_allocations SET spent_amount=$1 WHERE id=$2`
	res, err := s.db.Exec(q, spentAmount, id)
	if err != nil {
		return fmt.Errorf("store: UpdateAllocationSpending: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// ---- SpendingRecord ---------------------------------------------------------

func (s *PostgresStore) RecordSpending(sr *domain.SpendingRecord) error {
	const q = `
		INSERT INTO spending_records
			(id, budget_id, allocation_id, category, description, amount, reference, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`

	var allocID interface{}
	if sr.AllocationID != nil {
		allocID = *sr.AllocationID
	}
	_, err := s.db.Exec(q, sr.ID, sr.BudgetID, allocID, sr.Category, sr.Description, sr.Amount, sr.Reference, sr.CreatedAt)
	if err != nil {
		return fmt.Errorf("store: RecordSpending: %w", err)
	}
	return nil
}

func (s *PostgresStore) ListSpending(budgetID uuid.UUID, limit int) ([]domain.SpendingRecord, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	const q = `
		SELECT id, budget_id, allocation_id, category, description, amount, reference, created_at
		FROM spending_records
		WHERE budget_id = $1
		ORDER BY created_at DESC
		LIMIT $2`

	rows, err := s.db.Query(q, budgetID, limit)
	if err != nil {
		return nil, fmt.Errorf("store: ListSpending: %w", err)
	}
	defer rows.Close()

	var records []domain.SpendingRecord
	for rows.Next() {
		var sr domain.SpendingRecord
		var allocID sql.NullString
		if err := rows.Scan(
			&sr.ID, &sr.BudgetID, &allocID,
			&sr.Category, &sr.Description, &sr.Amount, &sr.Reference, &sr.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("store: ListSpending scan: %w", err)
		}
		if allocID.Valid {
			id, err := uuid.Parse(allocID.String)
			if err == nil {
				sr.AllocationID = &id
			}
		}
		records = append(records, sr)
	}
	return records, rows.Err()
}
