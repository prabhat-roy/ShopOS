package store

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"github.com/shopos/tax-reporting-service/internal/domain"
)

// Storer defines the persistence contract for tax records.
type Storer interface {
	SaveRecord(r *domain.TaxRecord) error
	GetRecord(id string) (*domain.TaxRecord, error)
	ListRecords(f domain.ListFilter) ([]*domain.TaxRecord, error)
	GetSummary(jurisdiction, period string) ([]*domain.TaxSummary, error)
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

// SaveRecord inserts a new TaxRecord into the database.
func (s *PostgresStore) SaveRecord(r *domain.TaxRecord) error {
	const q = `
		INSERT INTO tax_records
			(id, order_id, customer_id, jurisdiction, tax_type,
			 taxable_amount, tax_rate, tax_amount, currency, transaction_date, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`

	_, err := s.db.Exec(q,
		r.ID, r.OrderID, r.CustomerID, r.Jurisdiction, string(r.TaxType),
		r.TaxableAmount, r.TaxRate, r.TaxAmount, r.Currency,
		r.TransactionDate, r.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("store: save record: %w", err)
	}
	return nil
}

// GetRecord retrieves a single TaxRecord by its ID.
func (s *PostgresStore) GetRecord(id string) (*domain.TaxRecord, error) {
	const q = `
		SELECT id, order_id, customer_id, jurisdiction, tax_type,
		       taxable_amount, tax_rate, tax_amount, currency, transaction_date, created_at
		FROM tax_records
		WHERE id = $1`

	row := s.db.QueryRow(q, id)
	r := &domain.TaxRecord{}
	var taxType string
	err := row.Scan(
		&r.ID, &r.OrderID, &r.CustomerID, &r.Jurisdiction, &taxType,
		&r.TaxableAmount, &r.TaxRate, &r.TaxAmount, &r.Currency,
		&r.TransactionDate, &r.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("store: get record: %w", err)
	}
	r.TaxType = domain.TaxType(taxType)
	return r, nil
}

// ListRecords returns tax records filtered by optional jurisdiction, taxType and date range.
func (s *PostgresStore) ListRecords(f domain.ListFilter) ([]*domain.TaxRecord, error) {
	args := []interface{}{}
	conds := []string{}
	idx := 1

	if f.Jurisdiction != "" {
		conds = append(conds, fmt.Sprintf("jurisdiction = $%d", idx))
		args = append(args, f.Jurisdiction)
		idx++
	}
	if f.TaxType != "" {
		conds = append(conds, fmt.Sprintf("tax_type = $%d", idx))
		args = append(args, f.TaxType)
		idx++
	}
	if f.StartDate != "" {
		conds = append(conds, fmt.Sprintf("transaction_date >= $%d", idx))
		args = append(args, f.StartDate)
		idx++
	}
	if f.EndDate != "" {
		conds = append(conds, fmt.Sprintf("transaction_date <= $%d", idx))
		args = append(args, f.EndDate)
		idx++
	}

	where := ""
	if len(conds) > 0 {
		where = "WHERE " + strings.Join(conds, " AND ")
	}

	limit := 100
	if f.Limit > 0 && f.Limit <= 1000 {
		limit = f.Limit
	}

	q := fmt.Sprintf(`
		SELECT id, order_id, customer_id, jurisdiction, tax_type,
		       taxable_amount, tax_rate, tax_amount, currency, transaction_date, created_at
		FROM tax_records
		%s
		ORDER BY transaction_date DESC
		LIMIT %d`, where, limit)

	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil, fmt.Errorf("store: list records: %w", err)
	}
	defer rows.Close()

	var records []*domain.TaxRecord
	for rows.Next() {
		r := &domain.TaxRecord{}
		var taxType string
		if err := rows.Scan(
			&r.ID, &r.OrderID, &r.CustomerID, &r.Jurisdiction, &taxType,
			&r.TaxableAmount, &r.TaxRate, &r.TaxAmount, &r.Currency,
			&r.TransactionDate, &r.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("store: scan record: %w", err)
		}
		r.TaxType = domain.TaxType(taxType)
		records = append(records, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: rows error: %w", err)
	}
	return records, nil
}

// GetSummary aggregates tax records by jurisdiction and tax_type for the given YYYY-MM period.
// If jurisdiction is empty, all jurisdictions are included.
func (s *PostgresStore) GetSummary(jurisdiction, period string) ([]*domain.TaxSummary, error) {
	args := []interface{}{period}
	jurisdictionClause := ""
	if jurisdiction != "" {
		jurisdictionClause = "AND jurisdiction = $2"
		args = append(args, jurisdiction)
	}

	q := fmt.Sprintf(`
		SELECT
			jurisdiction,
			tax_type,
			to_char(transaction_date, 'YYYY-MM') AS period,
			SUM(taxable_amount)                  AS total_taxable,
			SUM(tax_amount)                      AS total_tax,
			COUNT(*)                             AS transaction_count
		FROM tax_records
		WHERE to_char(transaction_date, 'YYYY-MM') = $1
		%s
		GROUP BY jurisdiction, tax_type, to_char(transaction_date, 'YYYY-MM')
		ORDER BY jurisdiction, tax_type`, jurisdictionClause)

	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil, fmt.Errorf("store: get summary: %w", err)
	}
	defer rows.Close()

	var summaries []*domain.TaxSummary
	for rows.Next() {
		s := &domain.TaxSummary{}
		var taxType string
		if err := rows.Scan(
			&s.Jurisdiction, &taxType, &s.Period,
			&s.TotalTaxable, &s.TotalTax, &s.TransactionCount,
		); err != nil {
			return nil, fmt.Errorf("store: scan summary: %w", err)
		}
		s.TaxType = domain.TaxType(taxType)
		summaries = append(summaries, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: summary rows error: %w", err)
	}
	return summaries, nil
}
