package store

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopos/subscription-billing-service/domain"
)

// Storer defines the persistence interface used by the service layer.
type Storer interface {
	Create(sub *domain.Subscription) error
	Get(id string) (*domain.Subscription, error)
	List(customerID string) ([]*domain.Subscription, error)
	UpdateStatus(id string, status domain.SubStatus, cancelledAt *time.Time) error
	UpdateNextBilling(id string, next time.Time) error
	SaveBillingRecord(rec *domain.BillingRecord) error
	ListBillingRecords(subscriptionID string) ([]*domain.BillingRecord, error)
}

// PostgresStore implements Storer backed by PostgreSQL.
type PostgresStore struct {
	db *sql.DB
}

// NewPostgresStore constructs a store from an open *sql.DB.
func NewPostgresStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

// Create inserts a new subscription row.
func (s *PostgresStore) Create(sub *domain.Subscription) error {
	query := `
		INSERT INTO subscriptions
		  (id, customer_id, plan_id, product_id, status, cycle, price, currency,
		   trial_ends_at, next_billing_at, started_at, cancelled_at, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`

	_, err := s.db.Exec(query,
		sub.ID, sub.CustomerID, sub.PlanID, sub.ProductID,
		string(sub.Status), string(sub.Cycle), sub.Price, sub.Currency,
		sub.TrialEndsAt, sub.NextBillingAt, sub.StartedAt, sub.CancelledAt, sub.CreatedAt,
	)
	return err
}

// Get retrieves a single subscription by ID.
func (s *PostgresStore) Get(id string) (*domain.Subscription, error) {
	query := `
		SELECT id, customer_id, plan_id, product_id, status, cycle, price, currency,
		       trial_ends_at, next_billing_at, started_at, cancelled_at, created_at
		FROM subscriptions
		WHERE id = $1`

	row := s.db.QueryRow(query, id)
	return scanSubscription(row)
}

// List returns all subscriptions for a customer.
func (s *PostgresStore) List(customerID string) ([]*domain.Subscription, error) {
	query := `
		SELECT id, customer_id, plan_id, product_id, status, cycle, price, currency,
		       trial_ends_at, next_billing_at, started_at, cancelled_at, created_at
		FROM subscriptions
		WHERE customer_id = $1
		ORDER BY created_at DESC`

	rows, err := s.db.Query(query, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []*domain.Subscription
	for rows.Next() {
		sub, err := scanSubscriptionRow(rows)
		if err != nil {
			return nil, err
		}
		subs = append(subs, sub)
	}
	return subs, rows.Err()
}

// UpdateStatus updates the status (and optionally cancelled_at) of a subscription.
func (s *PostgresStore) UpdateStatus(id string, status domain.SubStatus, cancelledAt *time.Time) error {
	query := `UPDATE subscriptions SET status=$1, cancelled_at=$2 WHERE id=$3`
	res, err := s.db.Exec(query, string(status), cancelledAt, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// UpdateNextBilling advances the next_billing_at timestamp.
func (s *PostgresStore) UpdateNextBilling(id string, next time.Time) error {
	query := `UPDATE subscriptions SET next_billing_at=$1 WHERE id=$2`
	res, err := s.db.Exec(query, next, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// SaveBillingRecord inserts a billing attempt record.
func (s *PostgresStore) SaveBillingRecord(rec *domain.BillingRecord) error {
	query := `
		INSERT INTO billing_records
		  (id, subscription_id, amount, currency, status, created_at)
		VALUES ($1,$2,$3,$4,$5,$6)`

	_, err := s.db.Exec(query,
		rec.ID, rec.SubscriptionID, rec.Amount, rec.Currency, rec.Status, rec.CreatedAt,
	)
	return err
}

// ListBillingRecords returns all billing records for a subscription, newest first.
func (s *PostgresStore) ListBillingRecords(subscriptionID string) ([]*domain.BillingRecord, error) {
	query := `
		SELECT id, subscription_id, amount, currency, status, created_at
		FROM billing_records
		WHERE subscription_id = $1
		ORDER BY created_at DESC`

	rows, err := s.db.Query(query, subscriptionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*domain.BillingRecord
	for rows.Next() {
		rec := &domain.BillingRecord{}
		if err := rows.Scan(&rec.ID, &rec.SubscriptionID, &rec.Amount, &rec.Currency, &rec.Status, &rec.CreatedAt); err != nil {
			return nil, err
		}
		records = append(records, rec)
	}
	return records, rows.Err()
}

// ---- scan helpers ----

type rowScanner interface {
	Scan(dest ...any) error
}

func scanSubscription(row rowScanner) (*domain.Subscription, error) {
	sub := &domain.Subscription{}
	var status, cycle string
	err := row.Scan(
		&sub.ID, &sub.CustomerID, &sub.PlanID, &sub.ProductID,
		&status, &cycle, &sub.Price, &sub.Currency,
		&sub.TrialEndsAt, &sub.NextBillingAt, &sub.StartedAt, &sub.CancelledAt, &sub.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	sub.Status = domain.SubStatus(status)
	sub.Cycle = domain.BillingCycle(cycle)
	return sub, nil
}

func scanSubscriptionRow(rows *sql.Rows) (*domain.Subscription, error) {
	sub := &domain.Subscription{}
	var status, cycle string
	err := rows.Scan(
		&sub.ID, &sub.CustomerID, &sub.PlanID, &sub.ProductID,
		&status, &cycle, &sub.Price, &sub.Currency,
		&sub.TrialEndsAt, &sub.NextBillingAt, &sub.StartedAt, &sub.CancelledAt, &sub.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan subscription row: %w", err)
	}
	sub.Status = domain.SubStatus(status)
	sub.Cycle = domain.BillingCycle(cycle)
	return sub, nil
}

// NewID returns a new random UUID string.
func NewID() string {
	return uuid.New().String()
}
