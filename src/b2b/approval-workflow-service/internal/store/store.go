package store

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/shopos/approval-workflow-service/internal/domain"
)

// Storer defines all persistence operations for approval workflows.
type Storer interface {
	Create(w *domain.ApprovalWorkflow) error
	Get(id uuid.UUID) (*domain.ApprovalWorkflow, error)
	GetByEntityID(entityID uuid.UUID) (*domain.ApprovalWorkflow, error)
	List(entityType *domain.EntityType, status *domain.WorkflowStatus, orgID *uuid.UUID) ([]*domain.ApprovalWorkflow, error)
	UpdateWorkflow(w *domain.ApprovalWorkflow) error
}

// PostgresStore is the PostgreSQL-backed implementation of Storer.
type PostgresStore struct {
	db *sql.DB
}

// New opens a PostgreSQL connection and returns a ready PostgresStore.
func New(dsn string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	return &PostgresStore{db: db}, nil
}

// Close closes the underlying database connection pool.
func (s *PostgresStore) Close() error { return s.db.Close() }

// Create inserts a new ApprovalWorkflow row.
func (s *PostgresStore) Create(w *domain.ApprovalWorkflow) error {
	const qry = `
		INSERT INTO approval_workflows
			(id, entity_id, entity_type, org_id, total_amount, status,
			 steps, current_step_index, created_by, completed_at, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`

	_, err := s.db.Exec(qry,
		w.ID, w.EntityID, string(w.EntityType), w.OrgID, w.TotalAmount,
		string(w.Status), w.Steps, w.CurrentStepIndex,
		w.CreatedBy, w.CompletedAt, w.CreatedAt, w.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("store.Create: %w", err)
	}
	return nil
}

// Get retrieves an ApprovalWorkflow by its primary key.
func (s *PostgresStore) Get(id uuid.UUID) (*domain.ApprovalWorkflow, error) {
	const qry = `
		SELECT id, entity_id, entity_type, org_id, total_amount, status,
		       steps, current_step_index, created_by, completed_at, created_at, updated_at
		FROM approval_workflows WHERE id = $1`

	return s.scanOne(s.db.QueryRow(qry, id))
}

// GetByEntityID retrieves the workflow associated with a given entity UUID.
func (s *PostgresStore) GetByEntityID(entityID uuid.UUID) (*domain.ApprovalWorkflow, error) {
	const qry = `
		SELECT id, entity_id, entity_type, org_id, total_amount, status,
		       steps, current_step_index, created_by, completed_at, created_at, updated_at
		FROM approval_workflows WHERE entity_id = $1
		ORDER BY created_at DESC LIMIT 1`

	return s.scanOne(s.db.QueryRow(qry, entityID))
}

// List returns workflows filtered by optional entityType, status, and orgID.
func (s *PostgresStore) List(entityType *domain.EntityType, status *domain.WorkflowStatus, orgID *uuid.UUID) ([]*domain.ApprovalWorkflow, error) {
	qry := `
		SELECT id, entity_id, entity_type, org_id, total_amount, status,
		       steps, current_step_index, created_by, completed_at, created_at, updated_at
		FROM approval_workflows`

	var conditions []string
	var args []interface{}
	idx := 1

	if entityType != nil {
		conditions = append(conditions, fmt.Sprintf("entity_type = $%d", idx))
		args = append(args, string(*entityType))
		idx++
	}
	if status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", idx))
		args = append(args, string(*status))
		idx++
	}
	if orgID != nil {
		conditions = append(conditions, fmt.Sprintf("org_id = $%d", idx))
		args = append(args, *orgID)
		idx++
	}
	if len(conditions) > 0 {
		qry += " WHERE " + strings.Join(conditions, " AND ")
	}
	qry += " ORDER BY created_at DESC"

	rows, err := s.db.Query(qry, args...)
	if err != nil {
		return nil, fmt.Errorf("store.List: %w", err)
	}
	defer rows.Close()

	var workflows []*domain.ApprovalWorkflow
	for rows.Next() {
		w, err := s.scanRow(rows)
		if err != nil {
			return nil, err
		}
		workflows = append(workflows, w)
	}
	return workflows, rows.Err()
}

// UpdateWorkflow persists status, steps, currentStepIndex, completedAt, and updatedAt.
func (s *PostgresStore) UpdateWorkflow(w *domain.ApprovalWorkflow) error {
	const qry = `
		UPDATE approval_workflows
		SET status = $1, steps = $2, current_step_index = $3,
		    completed_at = $4, updated_at = $5
		WHERE id = $6`

	res, err := s.db.Exec(qry,
		string(w.Status), w.Steps, w.CurrentStepIndex,
		w.CompletedAt, time.Now().UTC(), w.ID,
	)
	if err != nil {
		return fmt.Errorf("store.UpdateWorkflow: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// --- scan helpers ---

type rowScanner interface {
	Scan(dest ...interface{}) error
}

func (s *PostgresStore) scanOne(row *sql.Row) (*domain.ApprovalWorkflow, error) {
	w := &domain.ApprovalWorkflow{}
	var st, et string
	err := row.Scan(
		&w.ID, &w.EntityID, &et, &w.OrgID, &w.TotalAmount, &st,
		&w.Steps, &w.CurrentStepIndex, &w.CreatedBy, &w.CompletedAt,
		&w.CreatedAt, &w.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("store.scanOne: %w", err)
	}
	w.Status = domain.WorkflowStatus(st)
	w.EntityType = domain.EntityType(et)
	return w, nil
}

func (s *PostgresStore) scanRow(rows *sql.Rows) (*domain.ApprovalWorkflow, error) {
	w := &domain.ApprovalWorkflow{}
	var st, et string
	if err := rows.Scan(
		&w.ID, &w.EntityID, &et, &w.OrgID, &w.TotalAmount, &st,
		&w.Steps, &w.CurrentStepIndex, &w.CreatedBy, &w.CompletedAt,
		&w.CreatedAt, &w.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("store.scanRow: %w", err)
	}
	w.Status = domain.WorkflowStatus(st)
	w.EntityType = domain.EntityType(et)
	return w, nil
}
