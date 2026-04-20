package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopos/gift-card-service/domain"
)

// Storer is the persistence contract required by Service.
type Storer interface {
	Issue(card *domain.GiftCard) (*domain.GiftCard, error)
	GetByCode(code string) (*domain.GiftCard, error)
	GetByID(id string) (*domain.GiftCard, error)
	Redeem(code, orderID string, amount float64) (*domain.RedemptionRecord, error)
	ListRedemptions(cardID string) ([]domain.RedemptionRecord, error)
	Deactivate(id string) error
}

// Service implements the gift-card business logic.
type Service struct {
	store Storer
}

// New creates a new Service with the provided Storer.
func New(store Storer) *Service {
	return &Service{store: store}
}

// IssueCard creates a new gift card. Code is auto-generated when left empty.
func (s *Service) IssueCard(initialBalance float64, currency, issuedTo string, expiresAt *time.Time) (*domain.GiftCard, error) {
	if initialBalance <= 0 {
		return nil, fmt.Errorf("initial balance must be positive")
	}
	if currency == "" {
		currency = "USD"
	}

	code, err := generateCode()
	if err != nil {
		return nil, fmt.Errorf("generate code: %w", err)
	}

	card := &domain.GiftCard{
		ID:             uuid.NewString(),
		Code:           code,
		InitialBalance: initialBalance,
		CurrentBalance: initialBalance,
		Currency:       currency,
		IssuedTo:       issuedTo,
		Active:         true,
		ExpiresAt:      expiresAt,
	}
	return s.store.Issue(card)
}

// GetCard returns a gift card by its code.
func (s *Service) GetCard(code string) (*domain.GiftCard, error) {
	return s.store.GetByCode(strings.ToUpper(code))
}

// CheckBalance returns the current balance and metadata for a card.
func (s *Service) CheckBalance(code string) (*domain.GiftCard, error) {
	card, err := s.store.GetByCode(strings.ToUpper(code))
	if err != nil {
		return nil, err
	}
	if !card.Active {
		return nil, domain.ErrCardInactive
	}
	if card.ExpiresAt != nil && time.Now().UTC().After(*card.ExpiresAt) {
		return nil, domain.ErrCardExpired
	}
	return card, nil
}

// Redeem applies a redemption against the card.
func (s *Service) Redeem(code, orderID string, amount float64) (*domain.RedemptionRecord, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("amount must be positive")
	}
	return s.store.Redeem(strings.ToUpper(code), orderID, amount)
}

// Deactivate deactivates a gift card by code.
func (s *Service) Deactivate(code string) error {
	card, err := s.store.GetByCode(strings.ToUpper(code))
	if err != nil {
		return err
	}
	return s.store.Deactivate(card.ID)
}

// generateCode returns a random uppercase hex code in XXXX-XXXX-XXXX-XXXX format.
func generateCode() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	h := strings.ToUpper(hex.EncodeToString(b))
	return fmt.Sprintf("%s-%s-%s-%s", h[0:4], h[4:8], h[8:12], h[12:16]), nil
}
