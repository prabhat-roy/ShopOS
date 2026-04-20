package service

import (
	"fmt"

	"github.com/shopos/wallet-service/domain"
)

// Storer is the persistence contract required by Service.
type Storer interface {
	GetWallet(customerID string) (*domain.Wallet, error)
	CreateWallet(customerID, currency string) (*domain.Wallet, error)
	Credit(walletID string, amount float64, reference, description string) (*domain.WalletTransaction, error)
	Debit(walletID string, amount float64, reference, description string) (*domain.WalletTransaction, error)
	GetTransactions(walletID string, limit int) ([]domain.WalletTransaction, error)
}

// Service implements the wallet business logic.
type Service struct {
	store Storer
}

// New creates a new Service with the provided Storer.
func New(store Storer) *Service {
	return &Service{store: store}
}

// GetOrCreateWallet returns the wallet for a customer, creating one with the default currency when absent.
func (s *Service) GetOrCreateWallet(customerID string) (*domain.Wallet, error) {
	w, err := s.store.GetWallet(customerID)
	if err == domain.ErrNotFound {
		return s.store.CreateWallet(customerID, "USD")
	}
	return w, err
}

// Credit adds money to the customer's wallet.
func (s *Service) Credit(customerID string, amount float64, reference, description string) (*domain.WalletTransaction, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("amount must be positive")
	}
	w, err := s.GetOrCreateWallet(customerID)
	if err != nil {
		return nil, fmt.Errorf("resolve wallet: %w", err)
	}
	return s.store.Credit(w.ID, amount, reference, description)
}

// Debit removes money from the customer's wallet; returns domain.ErrInsufficientFunds when balance is too low.
func (s *Service) Debit(customerID string, amount float64, reference, description string) (*domain.WalletTransaction, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("amount must be positive")
	}
	w, err := s.store.GetWallet(customerID)
	if err != nil {
		return nil, err
	}
	return s.store.Debit(w.ID, amount, reference, description)
}

// GetBalance returns the current balance for a customer's wallet.
func (s *Service) GetBalance(customerID string) (float64, string, error) {
	w, err := s.GetOrCreateWallet(customerID)
	if err != nil {
		return 0, "", err
	}
	return w.Balance, w.Currency, nil
}

// GetTransactions returns recent transactions for a customer's wallet.
func (s *Service) GetTransactions(customerID string, limit int) ([]domain.WalletTransaction, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	w, err := s.store.GetWallet(customerID)
	if err == domain.ErrNotFound {
		return []domain.WalletTransaction{}, nil
	}
	if err != nil {
		return nil, err
	}
	return s.store.GetTransactions(w.ID, limit)
}
