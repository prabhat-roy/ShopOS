package store

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/shopos/expense-management-service/internal/domain"
)

// Storer defines the persistence contract for expenses.
type Storer interface {
	Create(e *domain.Expense) error
	Get(id uuid.UUID) (*domain.Expense, error)
	List(f domain.ListFilter) ([]*domain.Expense, error)
	UpdateStatus(id uuid.UUID, status domain.ExpenseStatus, updatedAt time.Time) error
	SetApprover(id uuid.UUID, approverID uuid.UUID, approvedAt time.Time, updatedAt time.Time) error
	SetReimbursed(id uuid.UUID, reimbursedAt time.Time, updatedAt time.Time) error
	SetNotes(id uuid.UUID, notes string, updatedAt time.Time) error
	Delete(id uuid.UUID) error
}

// PostgresStore is the PostgreSQL implementation of Storer.
type PostgresStore struct {
	db *sql.DB
}

// New opens the Postgres connection pool and verifies connectivity.
func New(dsn string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("store: open db: %w", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)	return &PostgresStore{db: db}, nil
}

// Close releases the connection pool.
func (s *PostgresStore) Close() error {
	return s.db.Close()
}

// Create inserts a new expense record.
func (s *PostgresStore) Create(e *domain.Expense) error {
	const q = `
		INSERT INTO expenses
			(id, employee_id, category, amount, currency, description, receipt_url,
			 status, approved_by, approved_at, reimbursed_at, notes, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`

	_, err := s.db.Exec(q,
		e.ID, e.EmployeeID, string(e.Category), e.Amount, e.Currency,
		e.Description, e.ReceiptURL, string(e.Status),
		nullUUID(e.ApprovedBy), nullTime(e.ApprovedAt), nullTime(e.ReimbursedAt),
		e.Notes, e.CreatedAt, e.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("store: create expense: %w", err)
	}
	return nil
}

// Get retrieves a single expense by its UUID.
func (s *PostgresStore) Get(id uuid.UUID) (*domain.Expense, error) {
	const q = `
		SELECT id, employee_id, category, amount, currency, description, receipt_url,
		       status, approved_by, approved_at, reimbursed_at, notes, created_at, updated_at
		FROM expenses
		WHERE id = $1`

	row := s.db.QueryRow(q, id)
	return scanExpense(row.Scan)
}

// List returns expenses filtered by optional employeeId, status and category.
func (s *PostgresStore) List(f domain.ListFilter) ([]*domain.Expense, error) {
	args := []interface{}{}
	conds := []string{}
	idx := 1

	if f.EmployeeID != "" {
		conds = append(conds, fmt.Sprintf("employee_id = $%d", idx))
		args = append(args, f.EmployeeID)
		idx++
	}
	if f.Status != "" {
		conds = append(conds, fmt.Sprintf("status = $%d", idx))
		args = append(args, f.Status)
		idx++
	}
	if f.Category != "" {
		conds = append(conds, fmt.Sprintf("category = $%d", idx))
		args = append(args, f.Category)
		idx++
	}

	where := ""
	if len(conds) > 0 {
		where = "WHERE " + strings.Join(conds, " AND ")
	}

	q := fmt.Sprintf(`
		SELECT id, employee_id, category, amount, currency, description, receipt_url,
		       status, approved_by, approved_at, reimbursed_at, notes, created_at, updated_at
		FROM expenses
		%s
		ORDER BY created_at DESC
		LIMIT 500`, where)

	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil, fmt.Errorf("store: list expenses: %w", err)
	}
	defer rows.Close()

	var expenses []*domain.Expense
	for rows.Next() {
		e, err := scanExpense(rows.Scan)
		if err != nil {
			return nil, err
		}
		expenses = append(expenses, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: rows error: %w", err)
	}
	return expenses, nil
}

// UpdateStatus changes the status of an expense.
func (s *PostgresStore) UpdateStatus(id uuid.UUID, status domain.ExpenseStatus, updatedAt time.Time) error {
	res, err := s.db.Exec(
		`UPDATE expenses SET status = $1, updated_at = $2 WHERE id = $3`,
		string(status), updatedAt, id,
	)
	if err != nil {
		return fmt.Errorf("store: update status: %w", err)
	}
	return checkRowsAffected(res)
}

// SetApprover records the approver and approval timestamp.
func (s *PostgresStore) SetApprover(id uuid.UUID, approverID uuid.UUID, approvedAt time.Time, updatedAt time.Time) error {
	res, err := s.db.Exec(
		`UPDATE expenses SET approved_by = $1, approved_at = $2, updated_at = $3 WHERE id = $4`,
		approverID, approvedAt, updatedAt, id,
	)
	if err != nil {
		return fmt.Errorf("store: set approver: %w", err)
	}
	return checkRowsAffected(res)
}

// SetReimbursed marks the expense as reimbursed.
func (s *PostgresStore) SetReimbursed(id uuid.UUID, reimbursedAt time.Time, updatedAt time.Time) error {
	res, err := s.db.Exec(
		`UPDATE expenses SET reimbursed_at = $1, updated_at = $2 WHERE id = $3`,
		reimbursedAt, updatedAt, id,
	)
	if err != nil {
		return fmt.Errorf("store: set reimbursed: %w", err)
	}
	return checkRowsAffected(res)
}

// SetNotes updates the notes field (used for rejection reasons).
func (s *PostgresStore) SetNotes(id uuid.UUID, notes string, updatedAt time.Time) error {
	res, err := s.db.Exec(
		`UPDATE expenses SET notes = $1, updated_at = $2 WHERE id = $3`,
		notes, updatedAt, id,
	)
	if err != nil {
		return fmt.Errorf("store: set notes: %w", err)
	}
	return checkRowsAffected(res)
}

// Delete removes a DRAFT expense permanently.
func (s *PostgresStore) Delete(id uuid.UUID) error {
	res, err := s.db.Exec(`DELETE FROM expenses WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("store: delete expense: %w", err)
	}
	return checkRowsAffected(res)
}

// ---- helpers ----

type scanFn func(dest ...interface{}) error

func scanExpense(scan scanFn) (*domain.Expense, error) {
	e := &domain.Expense{}
	var (
		category     string
		status       string
		approvedBy   sql.NullString
		approvedAt   sql.NullTime
		reimbursedAt sql.NullTime
	)
	err := scan(
		&e.ID, &e.EmployeeID, &category, &e.Amount, &e.Currency,
		&e.Description, &e.ReceiptURL, &status,
		&approvedBy, &approvedAt, &reimbursedAt,
		&e.Notes, &e.CreatedAt, &e.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("store: scan expense: %w", err)
	}
	e.Category = domain.ExpenseCategory(category)
	e.Status = domain.ExpenseStatus(status)
	if approvedBy.Valid {
		id, err := uuid.Parse(approvedBy.String)
		if err == nil {
			e.ApprovedBy = &id
		}
	}
	if approvedAt.Valid {
		t := approvedAt.Time
		e.ApprovedAt = &t
	}
	if reimbursedAt.Valid {
		t := reimbursedAt.Time
		e.ReimbursedAt = &t
	}
	return e, nil
}

func nullUUID(u *uuid.UUID) interface{} {
	if u == nil {
		return nil
	}
	return u.String()
}

func nullTime(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	return *t
}

func checkRowsAffected(res sql.Result) error {
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("store: rows affected: %w", err)
	}
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}
