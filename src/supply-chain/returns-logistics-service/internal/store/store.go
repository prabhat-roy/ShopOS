package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/lib/pq"

	"github.com/shopos/returns-logistics-service/internal/domain"
)

// Storer defines the persistence operations for return authorisations.
type Storer interface {
	Create(ra *domain.ReturnAuth) error
	Get(id string) (*domain.ReturnAuth, error)
	List(customerID string) ([]*domain.ReturnAuth, error)
	UpdateStatus(id string, status domain.ReturnAuthStatus) error
	IssueLabel(id, label, trackingNumber string) error
	SetInspectionNotes(id, notes, warehouseID string) error
	SetRejectionReason(id, reason string) error
}

// PostgresStore is the PostgreSQL-backed implementation of Storer.
type PostgresStore struct {
	db *sql.DB
}

// New opens a connection to PostgreSQL and verifies it with a ping.
func New(dataSourceName string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)
	return &PostgresStore{db: db}, nil
}

// Close closes the underlying database connection pool.
func (s *PostgresStore) Close() error {
	return s.db.Close()
}

// Create inserts a new return authorisation record.
func (s *PostgresStore) Create(ra *domain.ReturnAuth) error {
	itemsJSON, err := json.Marshal(ra.Items)
	if err != nil {
		return fmt.Errorf("marshal items: %w", err)
	}

	query := `
		INSERT INTO return_authorizations
			(id, order_id, customer_id, items, reason, status, return_label,
			 tracking_number, warehouse_id, inspection_notes, rejection_reason,
			 created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`

	_, err = s.db.Exec(query,
		ra.ID, ra.OrderID, ra.CustomerID, string(itemsJSON),
		ra.Reason, string(ra.Status),
		nullableString(ra.ReturnLabel), nullableString(ra.TrackingNumber),
		nullableString(ra.WarehouseID), nullableString(ra.InspectionNotes),
		nullableString(ra.RejectionReason),
		ra.CreatedAt, ra.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert return_authorization: %w", err)
	}
	return nil
}

// Get retrieves a return authorisation by its ID.
func (s *PostgresStore) Get(id string) (*domain.ReturnAuth, error) {
	query := `
		SELECT id, order_id, customer_id, items, reason, status, return_label,
		       tracking_number, warehouse_id, inspection_notes, rejection_reason,
		       created_at, updated_at
		FROM return_authorizations
		WHERE id = $1`

	row := s.db.QueryRow(query, id)
	ra, err := scanReturnAuth(row)
	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get return_authorization: %w", err)
	}
	return ra, nil
}

// List retrieves all return authorisations, optionally filtered by customerID.
func (s *PostgresStore) List(customerID string) ([]*domain.ReturnAuth, error) {
	var (
		rows *sql.Rows
		err  error
	)
	if customerID != "" {
		rows, err = s.db.Query(`
			SELECT id, order_id, customer_id, items, reason, status, return_label,
			       tracking_number, warehouse_id, inspection_notes, rejection_reason,
			       created_at, updated_at
			FROM return_authorizations
			WHERE customer_id = $1
			ORDER BY created_at DESC`, customerID)
	} else {
		rows, err = s.db.Query(`
			SELECT id, order_id, customer_id, items, reason, status, return_label,
			       tracking_number, warehouse_id, inspection_notes, rejection_reason,
			       created_at, updated_at
			FROM return_authorizations
			ORDER BY created_at DESC`)
	}
	if err != nil {
		return nil, fmt.Errorf("list return_authorizations: %w", err)
	}
	defer rows.Close()

	var results []*domain.ReturnAuth
	for rows.Next() {
		ra, err := scanReturnAuthRow(rows)
		if err != nil {
			return nil, fmt.Errorf("scan return_authorization: %w", err)
		}
		results = append(results, ra)
	}
	return results, rows.Err()
}

// UpdateStatus updates only the status and updated_at fields.
func (s *PostgresStore) UpdateStatus(id string, status domain.ReturnAuthStatus) error {
	res, err := s.db.Exec(`
		UPDATE return_authorizations SET status=$1, updated_at=$2 WHERE id=$3`,
		string(status), time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("update status: %w", err)
	}
	return requireOneRow(res, id)
}

// IssueLabel stores the return label URL and tracking number for an authorisation.
func (s *PostgresStore) IssueLabel(id, label, trackingNumber string) error {
	res, err := s.db.Exec(`
		UPDATE return_authorizations
		SET return_label=$1, tracking_number=$2, status=$3, updated_at=$4
		WHERE id=$5`,
		label, trackingNumber, string(domain.StatusLabelIssued), time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("issue label: %w", err)
	}
	return requireOneRow(res, id)
}

// SetInspectionNotes stores inspection notes and the receiving warehouse ID.
func (s *PostgresStore) SetInspectionNotes(id, notes, warehouseID string) error {
	res, err := s.db.Exec(`
		UPDATE return_authorizations
		SET inspection_notes=$1, warehouse_id=$2, updated_at=$3
		WHERE id=$4`,
		notes, warehouseID, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("set inspection notes: %w", err)
	}
	return requireOneRow(res, id)
}

// SetRejectionReason stores the reason for a rejection.
func (s *PostgresStore) SetRejectionReason(id, reason string) error {
	res, err := s.db.Exec(`
		UPDATE return_authorizations SET rejection_reason=$1, updated_at=$2 WHERE id=$3`,
		reason, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("set rejection reason: %w", err)
	}
	return requireOneRow(res, id)
}

// ─── scan helpers ─────────────────────────────────────────────────────────────

type rowScanner interface {
	Scan(dest ...interface{}) error
}

func scanReturnAuth(row *sql.Row) (*domain.ReturnAuth, error) {
	return scanRA(row)
}

func scanReturnAuthRow(row *sql.Rows) (*domain.ReturnAuth, error) {
	return scanRA(row)
}

func scanRA(row rowScanner) (*domain.ReturnAuth, error) {
	var (
		ra              domain.ReturnAuth
		itemsJSON       string
		returnLabel     sql.NullString
		trackingNumber  sql.NullString
		warehouseID     sql.NullString
		inspectionNotes sql.NullString
		rejectionReason sql.NullString
		statusStr       string
	)

	err := row.Scan(
		&ra.ID, &ra.OrderID, &ra.CustomerID, &itemsJSON,
		&ra.Reason, &statusStr,
		&returnLabel, &trackingNumber, &warehouseID,
		&inspectionNotes, &rejectionReason,
		&ra.CreatedAt, &ra.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	ra.Status = domain.ReturnAuthStatus(statusStr)
	ra.ReturnLabel = returnLabel.String
	ra.TrackingNumber = trackingNumber.String
	ra.WarehouseID = warehouseID.String
	ra.InspectionNotes = inspectionNotes.String
	ra.RejectionReason = rejectionReason.String

	if err := json.Unmarshal([]byte(itemsJSON), &ra.Items); err != nil {
		return nil, fmt.Errorf("unmarshal items: %w", err)
	}
	return &ra, nil
}

// ─── misc helpers ─────────────────────────────────────────────────────────────

func nullableString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func requireOneRow(res sql.Result, id string) error {
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("%w: id=%s", domain.ErrNotFound, id)
	}
	return nil
}
