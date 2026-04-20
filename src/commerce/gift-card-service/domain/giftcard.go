package domain

import (
	"errors"
	"time"
)

// GiftCard represents an issued gift card.
type GiftCard struct {
	ID             string
	Code           string
	InitialBalance float64
	CurrentBalance float64
	Currency       string
	IssuedTo       string
	Active         bool
	ExpiresAt      *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// RedemptionRecord records a single redemption event against a gift card.
type RedemptionRecord struct {
	ID        string
	CardID    string
	OrderID   string
	Amount    float64
	CreatedAt time.Time
}

var (
	ErrNotFound           = errors.New("not found")
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrCardExpired        = errors.New("card expired")
	ErrCardInactive       = errors.New("card inactive or not found")
)
