package store

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/shopos/b2b-credit-limit-service/internal/domain"
)

// Storer defines all persistence operations for credit limits.
type Storer interface {
	CreateLimit(cl *domain.OrgCreditLimit) error
	GetLimit(id uuid.UUID) (*domain.OrgCreditLimit, error)
	GetByOrgID(orgID uuid.UUID) (*domain.OrgCreditLimit, error)
	// UpdateCredit atomically adjusts used_credit and available_credit.
	// delta > 0 increases used_credit (utilization); delta < 0 decreases it (payment).
	// Returns ErrInsufficientCredit if the resulting available_credit would go negative.
	UpdateCredit(orgID uuid.UUID, delta float64) (*domain.OrgCreditLimit, error)
	SaveTransaction(tx *domain.CreditTransaction) error
	ListTransactions(orgID uuid.UUID, limit int) ([]*domain.CreditTransaction, error)
	SuspendLimit(orgID uuid.UUID) error
	UpdateRiskScore(orgID uuid.UUID, score int) error
	UpdateCreditLimit(orgID uuid.UUID, newLimit float64) error
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

// CreateLimit inserts a new OrgCreditLimit row.
func (s *PostgresStore) CreateLimit(cl *domain.OrgCreditLimit) error {
	const qry = `
		INSERT INTO org_credit_limits
			(id, org_id, credit_limit, used_credit, available_credit,
			 currency, status, risk_score, last_reviewed_at, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`

	_, err := s.db.Exec(qry,
		cl.ID, cl.OrgID, cl.CreditLimit, cl.UsedCredit, cl.AvailableCredit,
		cl.Currency, string(cl.Status), cl.RiskScore,
		cl.LastReviewedAt, cl.CreatedAt, cl.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("store.CreateLimit: %w", err)
	}
	return nil
}

// GetLimit retrieves a credit limit record by its primary key.
func (s *PostgresStore) GetLimit(id uuid.UUID) (*domain.OrgCreditLimit, error) {
	const qry = `
		SELECT id, org_id, credit_limit, used_credit, available_credit,
		       currency, status, risk_score, last_reviewed_at, created_at, updated_at
		FROM org_credit_limits WHERE id = $1`
	return s.scanOne(s.db.QueryRow(qry, id))
}

// GetByOrgID retrieves the credit limit record for an organisation.
func (s *PostgresStore) GetByOrgID(orgID uuid.UUID) (*domain.OrgCreditLimit, error) {
	const qry = `
		SELECT id, org_id, credit_limit, used_credit, available_credit,
		       currency, status, risk_score, last_reviewed_at, created_at, updated_at
		FROM org_credit_limits WHERE org_id = $1`
	return s.scanOne(s.db.QueryRow(qry, orgID))
}

// UpdateCredit atomically adjusts the credit counters.
// delta > 0 = utilization; delta < 0 = payment/adjustment restoring credit.
func (s *PostgresStore) UpdateCredit(orgID uuid.UUID, delta float64) (*domain.OrgCreditLimit, error) {
	// We use a CTE to perform the check-and-update atomically.
	const qry = `
		UPDATE org_credit_limits
		SET used_credit      = used_credit + $1,
		    available_credit = available_credit - $1,
		    updated_at       = $2
		WHERE org_id = $3
		  AND (available_credit - $1) >= 0
		RETURNING id, org_id, credit_limit, used_credit, available_credit,
		          currency, status, risk_score, last_reviewed_at, created_at, updated_at`

	row := s.db.QueryRow(qry, delta, time.Now().UTC(), orgID)
	cl, err := s.scanOne(row)
	if errors.Is(err, domain.ErrNotFound) {
		// Either the org doesn't exist or the check failed — distinguish by lookup.
		if _, lookupErr := s.GetByOrgID(orgID); lookupErr != nil {
			return nil, lookupErr
		}
		return nil, domain.ErrInsufficientCredit
	}
	return cl, err
}

// SaveTransaction inserts a credit transaction record.
func (s *PostgresStore) SaveTransaction(tx *domain.CreditTransaction) error {
	const qry = `
		INSERT INTO credit_transactions
			(id, org_id, type, amount, reference, balance, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`

	_, err := s.db.Exec(qry,
		tx.ID, tx.OrgID, string(tx.Type), tx.Amount,
		tx.Reference, tx.Balance, tx.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("store.SaveTransaction: %w", err)
	}
	return nil
}

// ListTransactions returns the most recent transactions for an org.
func (s *PostgresStore) ListTransactions(orgID uuid.UUID, limit int) ([]*domain.CreditTransaction, error) {
	if limit <= 0 {
		limit = 50
	}
	const qry = `
		SELECT id, org_id, type, amount, reference, balance, created_at
		FROM credit_transactions
		WHERE org_id = $1
		ORDER BY created_at DESC
		LIMIT $2`

	rows, err := s.db.Query(qry, orgID, limit)
	if err != nil {
		return nil, fmt.Errorf("store.ListTransactions: %w", err)
	}
	defer rows.Close()

	var txs []*domain.CreditTransaction
	for rows.Next() {
		tx := &domain.CreditTransaction{}
		var txType string
		if err := rows.Scan(&tx.ID, &tx.OrgID, &txType, &tx.Amount,
			&tx.Reference, &tx.Balance, &tx.CreatedAt); err != nil {
			return nil, fmt.Errorf("store.ListTransactions scan: %w", err)
		}
		tx.Type = domain.TransactionType(txType)
		txs = append(txs, tx)
	}
	return txs, rows.Err()
}

// SuspendLimit sets the status to SUSPENDED.
func (s *PostgresStore) SuspendLimit(orgID uuid.UUID) error {
	const qry = `UPDATE org_credit_limits SET status = $1, updated_at = $2 WHERE org_id = $3`
	res, err := s.db.Exec(qry, string(domain.CreditLimitStatusSuspended), time.Now().UTC(), orgID)
	if err != nil {
		return fmt.Errorf("store.SuspendLimit: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// UpdateRiskScore sets the risk_score for an org's credit account.
func (s *PostgresStore) UpdateRiskScore(orgID uuid.UUID, score int) error {
	const qry = `UPDATE org_credit_limits SET risk_score = $1, updated_at = $2 WHERE org_id = $3`
	res, err := s.db.Exec(qry, score, time.Now().UTC(), orgID)
	if err != nil {
		return fmt.Errorf("store.UpdateRiskScore: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// UpdateCreditLimit changes the credit_limit and recalculates available_credit.
func (s *PostgresStore) UpdateCreditLimit(orgID uuid.UUID, newLimit float64) error {
	const qry = `
		UPDATE org_credit_limits
		SET credit_limit     = $1,
		    available_credit = $1 - used_credit,
		    updated_at       = $2
		WHERE org_id = $3`
	res, err := s.db.Exec(qry, newLimit, time.Now().UTC(), orgID)
	if err != nil {
		return fmt.Errorf("store.UpdateCreditLimit: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// --- scan helper ---

func (s *PostgresStore) scanOne(row *sql.Row) (*domain.OrgCreditLimit, error) {
	cl := &domain.OrgCreditLimit{}
	var st string
	err := row.Scan(
		&cl.ID, &cl.OrgID, &cl.CreditLimit, &cl.UsedCredit, &cl.AvailableCredit,
		&cl.Currency, &st, &cl.RiskScore,
		&cl.LastReviewedAt, &cl.CreatedAt, &cl.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("store.scanOne: %w", err)
	}
	cl.Status = domain.CreditLimitStatus(st)
	return cl, nil
}
