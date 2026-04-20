package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopos/b2b-credit-limit-service/internal/domain"
	"github.com/shopos/b2b-credit-limit-service/internal/store"
)

// Servicer exposes all business operations for B2B credit limits.
type Servicer interface {
	SetCreditLimit(orgID uuid.UUID, limit float64, currency string) (*domain.OrgCreditLimit, error)
	GetCreditLimit(id uuid.UUID) (*domain.OrgCreditLimit, error)
	GetByOrg(orgID uuid.UUID) (*domain.OrgCreditLimit, error)
	UtilizeCredit(orgID uuid.UUID, amount float64, reference string) (*domain.OrgCreditLimit, error)
	MakePayment(orgID uuid.UUID, amount float64, reference string) (*domain.OrgCreditLimit, error)
	AdjustLimit(orgID uuid.UUID, newLimit float64) (*domain.OrgCreditLimit, error)
	SuspendOrg(orgID uuid.UUID) error
	ReviewCredit(orgID uuid.UUID, newRiskScore int) (*domain.OrgCreditLimit, error)
	CheckAvailability(orgID uuid.UUID, amount float64) (*domain.AvailabilityCheck, error)
	GetCreditHistory(orgID uuid.UUID, limit int) ([]*domain.CreditTransaction, error)
}

// Service is the concrete implementation of Servicer.
type Service struct {
	store store.Storer
}

// New constructs a Service backed by the supplied Storer.
func New(s store.Storer) *Service {
	return &Service{store: s}
}

// SetCreditLimit creates or fully replaces the credit limit record for an org.
// If a record already exists it updates the limit; otherwise it creates a new one.
func (svc *Service) SetCreditLimit(orgID uuid.UUID, limit float64, currency string) (*domain.OrgCreditLimit, error) {
	if orgID == uuid.Nil {
		return nil, fmt.Errorf("org_id is required")
	}
	if limit < 0 {
		return nil, fmt.Errorf("credit limit must be non-negative")
	}
	if currency == "" {
		currency = "USD"
	}

	// If the org already has a record, update the limit instead of duplicating.
	existing, err := svc.store.GetByOrgID(orgID)
	if err == nil {
		// Record exists — adjust to new limit.
		if err := svc.store.UpdateCreditLimit(orgID, limit); err != nil {
			return nil, fmt.Errorf("update credit limit: %w", err)
		}
		return svc.store.GetByOrgID(orgID)
	}

	// Doesn't exist yet — create fresh.
	now := time.Now().UTC()
	cl := &domain.OrgCreditLimit{
		ID:              uuid.New(),
		OrgID:           orgID,
		CreditLimit:     limit,
		UsedCredit:      0,
		AvailableCredit: limit,
		Currency:        currency,
		Status:          domain.CreditLimitStatusActive,
		RiskScore:       50, // neutral default
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	_ = existing // suppress unused warning
	if err := svc.store.CreateLimit(cl); err != nil {
		return nil, fmt.Errorf("create credit limit: %w", err)
	}
	return cl, nil
}

// GetCreditLimit retrieves a credit limit record by its primary key.
func (svc *Service) GetCreditLimit(id uuid.UUID) (*domain.OrgCreditLimit, error) {
	return svc.store.GetLimit(id)
}

// GetByOrg retrieves the credit limit record for an organisation.
func (svc *Service) GetByOrg(orgID uuid.UUID) (*domain.OrgCreditLimit, error) {
	return svc.store.GetByOrgID(orgID)
}

// UtilizeCredit deducts the requested amount from available credit.
func (svc *Service) UtilizeCredit(orgID uuid.UUID, amount float64, reference string) (*domain.OrgCreditLimit, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("amount must be positive")
	}

	// Check account status before deducting.
	cl, err := svc.store.GetByOrgID(orgID)
	if err != nil {
		return nil, err
	}
	if cl.Status == domain.CreditLimitStatusSuspended {
		return nil, domain.ErrAccountSuspended
	}

	updated, err := svc.store.UpdateCredit(orgID, amount) // positive delta = utilization
	if err != nil {
		return nil, err
	}

	tx := &domain.CreditTransaction{
		ID:        uuid.New(),
		OrgID:     orgID,
		Type:      domain.TransactionTypeUtilization,
		Amount:    amount,
		Reference: reference,
		Balance:   updated.AvailableCredit,
		CreatedAt: time.Now().UTC(),
	}
	if err := svc.store.SaveTransaction(tx); err != nil {
		// Non-fatal: transaction audit failure shouldn't roll back the credit deduction.
		// In a production system this would be handled via outbox pattern.
		_ = err
	}
	return updated, nil
}

// MakePayment restores available credit (negative delta).
func (svc *Service) MakePayment(orgID uuid.UUID, amount float64, reference string) (*domain.OrgCreditLimit, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("payment amount must be positive")
	}

	// Negative delta reduces used_credit and increases available_credit.
	updated, err := svc.store.UpdateCredit(orgID, -amount)
	if err != nil {
		return nil, err
	}

	tx := &domain.CreditTransaction{
		ID:        uuid.New(),
		OrgID:     orgID,
		Type:      domain.TransactionTypePayment,
		Amount:    amount,
		Reference: reference,
		Balance:   updated.AvailableCredit,
		CreatedAt: time.Now().UTC(),
	}
	_ = svc.store.SaveTransaction(tx)
	return updated, nil
}

// AdjustLimit changes the credit_limit ceiling and recomputes available credit.
func (svc *Service) AdjustLimit(orgID uuid.UUID, newLimit float64) (*domain.OrgCreditLimit, error) {
	if newLimit < 0 {
		return nil, fmt.Errorf("new limit must be non-negative")
	}
	if err := svc.store.UpdateCreditLimit(orgID, newLimit); err != nil {
		return nil, err
	}
	cl, err := svc.store.GetByOrgID(orgID)
	if err != nil {
		return nil, err
	}
	tx := &domain.CreditTransaction{
		ID:        uuid.New(),
		OrgID:     orgID,
		Type:      domain.TransactionTypeAdjustment,
		Amount:    newLimit,
		Reference: "limit-adjustment",
		Balance:   cl.AvailableCredit,
		CreatedAt: time.Now().UTC(),
	}
	_ = svc.store.SaveTransaction(tx)
	return cl, nil
}

// SuspendOrg sets the org's credit account status to SUSPENDED.
func (svc *Service) SuspendOrg(orgID uuid.UUID) error {
	return svc.store.SuspendLimit(orgID)
}

// ReviewCredit updates the risk score and stamps last_reviewed_at.
func (svc *Service) ReviewCredit(orgID uuid.UUID, newRiskScore int) (*domain.OrgCreditLimit, error) {
	if newRiskScore < 0 || newRiskScore > 100 {
		return nil, fmt.Errorf("risk_score must be between 0 and 100")
	}
	if err := svc.store.UpdateRiskScore(orgID, newRiskScore); err != nil {
		return nil, err
	}
	return svc.store.GetByOrgID(orgID)
}

// CheckAvailability returns whether the org has sufficient available credit.
func (svc *Service) CheckAvailability(orgID uuid.UUID, amount float64) (*domain.AvailabilityCheck, error) {
	cl, err := svc.store.GetByOrgID(orgID)
	if err != nil {
		return nil, err
	}
	check := &domain.AvailabilityCheck{
		RequestedAmount: amount,
		AvailableAmount: cl.AvailableCredit,
		Available:       cl.Status == domain.CreditLimitStatusActive && cl.AvailableCredit >= amount,
	}
	return check, nil
}

// GetCreditHistory returns the transaction history for an org.
func (svc *Service) GetCreditHistory(orgID uuid.UUID, limit int) ([]*domain.CreditTransaction, error) {
	return svc.store.ListTransactions(orgID, limit)
}
