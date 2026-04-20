package service

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/shopos/voucher-service/domain"
	"github.com/shopos/voucher-service/store"
)

// VoucherService implements business logic for vouchers.
type VoucherService struct {
	store *store.Store
}

// New creates a VoucherService.
func New(s *store.Store) *VoucherService {
	return &VoucherService{store: s}
}

// IssueVoucher creates and persists a new voucher.
func (svc *VoucherService) IssueVoucher(ctx context.Context, customerID string, amount float64, currency string, expiresAt time.Time) (*domain.Voucher, error) {
	if currency == "" {
		currency = "USD"
	}
	v := &domain.Voucher{
		ID:         uuid.New().String(),
		Code:       generateCode(),
		CustomerID: customerID,
		Amount:     amount,
		Currency:   strings.ToUpper(currency),
		Used:       false,
		ExpiresAt:  expiresAt,
		CreatedAt:  time.Now().UTC(),
	}
	if err := svc.store.Issue(ctx, v); err != nil {
		return nil, fmt.Errorf("issue voucher: %w", err)
	}
	return v, nil
}

// GetVoucher retrieves a voucher by code.
func (svc *VoucherService) GetVoucher(ctx context.Context, code string) (*domain.Voucher, error) {
	return svc.store.GetByCode(ctx, code)
}

// UseVoucher validates and marks a voucher as used for the given order.
func (svc *VoucherService) UseVoucher(ctx context.Context, code, orderID string) (*domain.Voucher, error) {
	v, err := svc.store.GetByCode(ctx, code)
	if err != nil {
		return nil, err // ErrNotFound propagates
	}
	if v.Used {
		return nil, domain.ErrAlreadyUsed
	}
	if time.Now().UTC().After(v.ExpiresAt) {
		return nil, domain.ErrExpired
	}
	if err := svc.store.Use(ctx, code, orderID); err != nil {
		return nil, err
	}
	// Return updated voucher
	return svc.store.GetByCode(ctx, code)
}

// ListCustomerVouchers returns all vouchers for a customer.
func (svc *VoucherService) ListCustomerVouchers(ctx context.Context, customerID string) ([]*domain.Voucher, error) {
	return svc.store.ListByCustomer(ctx, customerID)
}

// generateCode produces a random 10-character alphanumeric voucher code.
func generateCode() string {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, 10)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}
