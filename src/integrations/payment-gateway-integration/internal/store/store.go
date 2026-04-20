package store

import (
	"fmt"
	"sync"

	"github.com/shopos/payment-gateway-integration/internal/domain"
)

// PaymentStore is a thread-safe in-memory store for payment intents and refunds.
type PaymentStore struct {
	mu      sync.RWMutex
	intents map[string]*domain.PaymentIntent
	refunds map[string]*domain.RefundResponse
}

// New returns an initialised PaymentStore.
func New() *PaymentStore {
	return &PaymentStore{
		intents: make(map[string]*domain.PaymentIntent),
		refunds: make(map[string]*domain.RefundResponse),
	}
}

// SaveIntent persists or replaces a PaymentIntent.
func (s *PaymentStore) SaveIntent(pi *domain.PaymentIntent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.intents[pi.ID] = pi
}

// GetIntent retrieves a PaymentIntent by its ShopOS ID.
func (s *PaymentStore) GetIntent(id string) (*domain.PaymentIntent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	pi, ok := s.intents[id]
	if !ok {
		return nil, fmt.Errorf("payment intent %q not found", id)
	}
	return pi, nil
}

// ListByOrderId returns all intents for a given order.
func (s *PaymentStore) ListByOrderId(orderID string) []*domain.PaymentIntent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []*domain.PaymentIntent{}
	for _, pi := range s.intents {
		if pi.OrderID == orderID {
			out = append(out, pi)
		}
	}
	return out
}

// ListByCustomerId returns all intents for a given customer.
func (s *PaymentStore) ListByCustomerId(customerID string) []*domain.PaymentIntent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []*domain.PaymentIntent{}
	for _, pi := range s.intents {
		if pi.CustomerID == customerID {
			out = append(out, pi)
		}
	}
	return out
}

// SaveRefund persists a RefundResponse.
func (s *PaymentStore) SaveRefund(r *domain.RefundResponse) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.refunds[r.RefundID] = r
}

// GetRefund retrieves a RefundResponse by refund ID.
func (s *PaymentStore) GetRefund(refundID string) (*domain.RefundResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.refunds[refundID]
	if !ok {
		return nil, fmt.Errorf("refund %q not found", refundID)
	}
	return r, nil
}
