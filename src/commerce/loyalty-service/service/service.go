package service

import (
	"fmt"

	"github.com/shopos/loyalty-service/domain"
)

// Storer is the persistence contract required by Service.
type Storer interface {
	GetAccount(customerID string) (*domain.LoyaltyAccount, error)
	CreateAccount(customerID string) (*domain.LoyaltyAccount, error)
	EarnPoints(customerID string, points int64, orderID, description string) (*domain.PointTransaction, error)
	RedeemPoints(customerID string, points int64, orderID string) (*domain.PointTransaction, error)
	GetTransactions(customerID string, limit int) ([]domain.PointTransaction, error)
}

// Service implements the loyalty business logic.
type Service struct {
	store Storer
}

// New creates a new Service with the provided Storer.
func New(store Storer) *Service {
	return &Service{store: store}
}

// getOrCreate returns the account for the customer, creating one if it does not yet exist.
func (s *Service) getOrCreate(customerID string) (*domain.LoyaltyAccount, error) {
	acc, err := s.store.GetAccount(customerID)
	if err == domain.ErrNotFound {
		return s.store.CreateAccount(customerID)
	}
	return acc, err
}

// EarnPoints awards 1 point per dollar (rounded down) based on the provided dollar amount.
// Pass the dollar amount as the `dollarAmount` argument; points will be calculated internally.
func (s *Service) EarnPoints(customerID string, dollarAmount int64, orderID, description string) (*domain.PointTransaction, error) {
	if _, err := s.getOrCreate(customerID); err != nil {
		return nil, fmt.Errorf("ensure account: %w", err)
	}
	// 1 point per dollar spent.
	points := dollarAmount
	if points <= 0 {
		return nil, fmt.Errorf("points must be positive")
	}
	if description == "" {
		description = fmt.Sprintf("Earned %d points for order %s", points, orderID)
	}
	return s.store.EarnPoints(customerID, points, orderID, description)
}

// RedeemPoints redeems the exact number of points specified. 100 points = $1.
// Returns the transaction and the equivalent dollar value.
func (s *Service) RedeemPoints(customerID string, points int64, orderID string) (*domain.PointTransaction, float64, error) {
	if points <= 0 {
		return nil, 0, fmt.Errorf("points must be positive")
	}
	txn, err := s.store.RedeemPoints(customerID, points, orderID)
	if err != nil {
		return nil, 0, err
	}
	// 100 points = $1.00
	dollarValue := float64(points) / 100.0
	return txn, dollarValue, nil
}

// GetAccount returns the loyalty account for a customer.
func (s *Service) GetAccount(customerID string) (*domain.LoyaltyAccount, error) {
	return s.getOrCreate(customerID)
}

// GetTransactions returns recent transactions for a customer.
func (s *Service) GetTransactions(customerID string, limit int) ([]domain.PointTransaction, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	return s.store.GetTransactions(customerID, limit)
}

// GetBalance returns the current points balance for a customer.
func (s *Service) GetBalance(customerID string) (int64, error) {
	acc, err := s.getOrCreate(customerID)
	if err != nil {
		return 0, err
	}
	return acc.Points, nil
}
