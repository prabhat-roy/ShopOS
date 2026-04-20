package store

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/shopos/quote-rfq-service/internal/domain"
)

// Storer defines all persistence operations for quotes.
type Storer interface {
	Create(q *domain.Quote) error
	Get(id uuid.UUID) (*domain.Quote, error)
	List(orgID *uuid.UUID, status *domain.QuoteStatus) ([]*domain.Quote, error)
	UpdateStatus(id uuid.UUID, status domain.QuoteStatus) error
	SetQuotedPrices(id uuid.UUID, items domain.QuoteItems, totalAmount float64, validUntil *time.Time) error
	UpdateNotes(id uuid.UUID, notes string) error
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

// Create inserts a new Quote row.
func (s *PostgresStore) Create(q *domain.Quote) error {
	const qry = `
		INSERT INTO quotes
			(id, org_id, title, description, items, requested_delivery,
			 status, total_amount, currency, valid_until, notes, created_by, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`

	_, err := s.db.Exec(qry,
		q.ID, q.OrgID, q.Title, q.Description, q.Items,
		q.RequestedDelivery, string(q.Status), q.TotalAmount,
		q.Currency, q.ValidUntil, q.Notes, q.CreatedBy,
		q.CreatedAt, q.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("store.Create: %w", err)
	}
	return nil
}

// Get retrieves a Quote by its primary key.
func (s *PostgresStore) Get(id uuid.UUID) (*domain.Quote, error) {
	const qry = `
		SELECT id, org_id, title, description, items, requested_delivery,
		       status, total_amount, currency, valid_until, notes, created_by, created_at, updated_at
		FROM quotes WHERE id = $1`

	q := &domain.Quote{}
	var status string
	err := s.db.QueryRow(qry, id).Scan(
		&q.ID, &q.OrgID, &q.Title, &q.Description, &q.Items,
		&q.RequestedDelivery, &status, &q.TotalAmount,
		&q.Currency, &q.ValidUntil, &q.Notes, &q.CreatedBy,
		&q.CreatedAt, &q.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("store.Get: %w", err)
	}
	q.Status = domain.QuoteStatus(status)
	return q, nil
}

// List returns quotes filtered by optional orgID and status.
func (s *PostgresStore) List(orgID *uuid.UUID, status *domain.QuoteStatus) ([]*domain.Quote, error) {
	qry := `
		SELECT id, org_id, title, description, items, requested_delivery,
		       status, total_amount, currency, valid_until, notes, created_by, created_at, updated_at
		FROM quotes`

	var conditions []string
	var args []interface{}
	idx := 1

	if orgID != nil {
		conditions = append(conditions, fmt.Sprintf("org_id = $%d", idx))
		args = append(args, *orgID)
		idx++
	}
	if status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", idx))
		args = append(args, string(*status))
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

	var quotes []*domain.Quote
	for rows.Next() {
		q := &domain.Quote{}
		var st string
		if err := rows.Scan(
			&q.ID, &q.OrgID, &q.Title, &q.Description, &q.Items,
			&q.RequestedDelivery, &st, &q.TotalAmount,
			&q.Currency, &q.ValidUntil, &q.Notes, &q.CreatedBy,
			&q.CreatedAt, &q.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("store.List scan: %w", err)
		}
		q.Status = domain.QuoteStatus(st)
		quotes = append(quotes, q)
	}
	return quotes, rows.Err()
}

// UpdateStatus sets only the status column (and updated_at) for a quote.
func (s *PostgresStore) UpdateStatus(id uuid.UUID, status domain.QuoteStatus) error {
	const qry = `UPDATE quotes SET status = $1, updated_at = $2 WHERE id = $3`
	res, err := s.db.Exec(qry, string(status), time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("store.UpdateStatus: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// SetQuotedPrices updates items, total amount, valid until, and status to QUOTED.
func (s *PostgresStore) SetQuotedPrices(id uuid.UUID, items domain.QuoteItems, totalAmount float64, validUntil *time.Time) error {
	const qry = `
		UPDATE quotes
		SET items = $1, total_amount = $2, valid_until = $3,
		    status = $4, updated_at = $5
		WHERE id = $6`

	res, err := s.db.Exec(qry, items, totalAmount, validUntil,
		string(domain.QuoteStatusQuoted), time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("store.SetQuotedPrices: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// UpdateNotes sets the notes field on a quote.
func (s *PostgresStore) UpdateNotes(id uuid.UUID, notes string) error {
	const qry = `UPDATE quotes SET notes = $1, updated_at = $2 WHERE id = $3`
	res, err := s.db.Exec(qry, notes, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("store.UpdateNotes: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}
