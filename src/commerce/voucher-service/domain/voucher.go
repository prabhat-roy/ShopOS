package domain

import (
	"errors"
	"time"
)

// Sentinel errors returned by the voucher service.
var (
	ErrNotFound    = errors.New("voucher not found")
	ErrAlreadyUsed = errors.New("voucher has already been used")
	ErrExpired     = errors.New("voucher has expired")
)

// Voucher represents a single-use promotional voucher code.
type Voucher struct {
	ID         string     `json:"id"`
	Code       string     `json:"code"`
	CustomerID string     `json:"customer_id"`
	Amount     float64    `json:"amount"`
	Currency   string     `json:"currency"`
	Used       bool       `json:"used"`
	UsedAt     *time.Time `json:"used_at,omitempty"`
	OrderID    string     `json:"order_id,omitempty"`
	ExpiresAt  time.Time  `json:"expires_at"`
	CreatedAt  time.Time  `json:"created_at"`
}
