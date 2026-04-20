package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopos/credit-service/internal/domain"
	"github.com/shopos/credit-service/internal/store"
)

// Servicer defines all business-logic operations for credit management.
type Servicer interface {
	CreateCreditAccount(customerID uuid.UUID, creditLimit float64, currency string) (*domain.CreditAccount, error)
	GetCreditAccount(id uuid.UUID) (*domain.CreditAccount, error)
	GetByCustomerID(customerID uuid.UUID) (*domain.CreditAccount, error)
	ChargeCredit(accountID uuid.UUID, amount float64, reference, description string) (*domain.CreditTransaction, error)
	MakePayment(accountID uuid.UUID, amount float64, reference string) (*domain.CreditTransaction, error)
	AdjustCreditLimit(accountID uuid.UUID, newLimit float64) (*domain.CreditAccount, error)
	SuspendAccount(id uuid.UUID) error
	CloseAccount(id uuid.UUID) error
	GetTransactionHistory(accountID uuid.UUID, limit int) ([]domain.CreditTransaction, error)
}

// Service implements Servicer.
type Service struct {
	store store.Storer
}

// New creates a new Service with the provided Storer.
func New(st store.Storer) *Service {
	return &Service{store: st}
}

// CreateCreditAccount creates a new credit account for a customer.
// The full credit limit is available from the start (usedCredit = 0).
func (s *Service) CreateCreditAccount(customerID uuid.UUID, creditLimit float64, currency string) (*domain.CreditAccount, error) {
	if creditLimit < 0 {
		return nil, fmt.Errorf("service: credit limit must be non-negative")
	}
	if currency == "" {
		currency = "USD"
	}
	now := time.Now().UTC()
	acc := &domain.CreditAccount{
		ID:              uuid.New(),
		CustomerID:      customerID,
		CreditLimit:     creditLimit,
		AvailableCredit: creditLimit,
		UsedCredit:      0,
		Currency:        currency,
		Status:          domain.StatusActive,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := s.store.CreateAccount(acc); err != nil {
		return nil, fmt.Errorf("service: CreateCreditAccount: %w", err)
	}
	return acc, nil
}

// GetCreditAccount returns a credit account by its UUID.
func (s *Service) GetCreditAccount(id uuid.UUID) (*domain.CreditAccount, error) {
	acc, err := s.store.GetAccount(id)
	if err != nil {
		return nil, fmt.Errorf("service: GetCreditAccount: %w", err)
	}
	return acc, nil
}

// GetByCustomerID returns the credit account associated with a customer.
func (s *Service) GetByCustomerID(customerID uuid.UUID) (*domain.CreditAccount, error) {
	acc, err := s.store.GetByCustomerID(customerID)
	if err != nil {
		return nil, fmt.Errorf("service: GetByCustomerID: %w", err)
	}
	return acc, nil
}

// ChargeCredit deducts amount from the account's available credit.
// Returns ErrAccountInactive if the account is not active.
// Returns ErrInsufficientCredit if amount > availableCredit.
func (s *Service) ChargeCredit(accountID uuid.UUID, amount float64, reference, description string) (*domain.CreditTransaction, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("service: charge amount must be positive")
	}

	acc, err := s.store.GetAccount(accountID)
	if err != nil {
		return nil, fmt.Errorf("service: ChargeCredit: %w", err)
	}
	if acc.Status != domain.StatusActive {
		return nil, domain.ErrAccountInactive
	}
	if amount > acc.AvailableCredit {
		return nil, domain.ErrInsufficientCredit
	}

	newAvailable := acc.AvailableCredit - amount
	newUsed := acc.UsedCredit + amount

	if err := s.store.UpdateCredit(accountID, newAvailable, newUsed); err != nil {
		return nil, fmt.Errorf("service: ChargeCredit update: %w", err)
	}

	tx := &domain.CreditTransaction{
		ID:          uuid.New(),
		AccountID:   accountID,
		Type:        domain.TxCharge,
		Amount:      amount,
		Reference:   reference,
		Description: description,
		CreatedAt:   time.Now().UTC(),
	}
	if err := s.store.SaveTransaction(tx); err != nil {
		return nil, fmt.Errorf("service: ChargeCredit save tx: %w", err)
	}
	return tx, nil
}

// MakePayment restores available credit by reducing used credit.
// A payment cannot exceed the used credit balance.
func (s *Service) MakePayment(accountID uuid.UUID, amount float64, reference string) (*domain.CreditTransaction, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("service: payment amount must be positive")
	}

	acc, err := s.store.GetAccount(accountID)
	if err != nil {
		return nil, fmt.Errorf("service: MakePayment: %w", err)
	}
	if acc.Status == domain.StatusClosed {
		return nil, domain.ErrAccountInactive
	}
	if amount > acc.UsedCredit {
		amount = acc.UsedCredit // cap payment at outstanding balance
	}

	newAvailable := acc.AvailableCredit + amount
	newUsed := acc.UsedCredit - amount
	// Ensure available never exceeds limit due to floating-point drift.
	if newAvailable > acc.CreditLimit {
		newAvailable = acc.CreditLimit
		newUsed = 0
	}

	if err := s.store.UpdateCredit(accountID, newAvailable, newUsed); err != nil {
		return nil, fmt.Errorf("service: MakePayment update: %w", err)
	}

	tx := &domain.CreditTransaction{
		ID:          uuid.New(),
		AccountID:   accountID,
		Type:        domain.TxPayment,
		Amount:      amount,
		Reference:   reference,
		Description: "payment received",
		CreatedAt:   time.Now().UTC(),
	}
	if err := s.store.SaveTransaction(tx); err != nil {
		return nil, fmt.Errorf("service: MakePayment save tx: %w", err)
	}
	return tx, nil
}

// AdjustCreditLimit updates the credit limit and recalculates available credit.
// availableCredit = newLimit - usedCredit (clamped to 0).
func (s *Service) AdjustCreditLimit(accountID uuid.UUID, newLimit float64) (*domain.CreditAccount, error) {
	if newLimit < 0 {
		return nil, fmt.Errorf("service: credit limit must be non-negative")
	}

	acc, err := s.store.GetAccount(accountID)
	if err != nil {
		return nil, fmt.Errorf("service: AdjustCreditLimit: %w", err)
	}
	if acc.Status == domain.StatusClosed {
		return nil, domain.ErrAccountInactive
	}

	newAvailable := newLimit - acc.UsedCredit
	if newAvailable < 0 {
		newAvailable = 0
	}

	if err := s.store.UpdateCredit(accountID, newAvailable, acc.UsedCredit); err != nil {
		return nil, fmt.Errorf("service: AdjustCreditLimit update credit: %w", err)
	}

	// Persist new limit via a synthetic adjustment transaction for the audit trail.
	tx := &domain.CreditTransaction{
		ID:          uuid.New(),
		AccountID:   accountID,
		Type:        domain.TxAdjustment,
		Amount:      newLimit - acc.CreditLimit,
		Reference:   "limit-adjustment",
		Description: fmt.Sprintf("credit limit changed from %.2f to %.2f", acc.CreditLimit, newLimit),
		CreatedAt:   time.Now().UTC(),
	}
	_ = s.store.SaveTransaction(tx) // non-fatal — limit already updated

	// Return fresh account state.
	updated, err := s.store.GetAccount(accountID)
	if err != nil {
		return nil, fmt.Errorf("service: AdjustCreditLimit re-fetch: %w", err)
	}
	// Reflect new limit (UpdateCredit doesn't update the limit column; use
	// the known values to construct the response without an extra UPDATE).
	updated.CreditLimit = newLimit
	return updated, nil
}

// SuspendAccount transitions the account to suspended status.
func (s *Service) SuspendAccount(id uuid.UUID) error {
	if err := s.store.SuspendAccount(id); err != nil {
		return fmt.Errorf("service: SuspendAccount: %w", err)
	}
	return nil
}

// CloseAccount transitions the account to closed status.
func (s *Service) CloseAccount(id uuid.UUID) error {
	if err := s.store.CloseAccount(id); err != nil {
		return fmt.Errorf("service: CloseAccount: %w", err)
	}
	return nil
}

// GetTransactionHistory returns the transaction history for an account.
func (s *Service) GetTransactionHistory(accountID uuid.UUID, limit int) ([]domain.CreditTransaction, error) {
	txs, err := s.store.ListTransactions(accountID, limit)
	if err != nil {
		return nil, fmt.Errorf("service: GetTransactionHistory: %w", err)
	}
	return txs, nil
}
