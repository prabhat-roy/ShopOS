package domain

import (
	"errors"
	"time"
)

// SubStatus represents the lifecycle state of a subscription.
type SubStatus string

const (
	SubActive    SubStatus = "active"
	SubPaused    SubStatus = "paused"
	SubCancelled SubStatus = "cancelled"
	SubExpired   SubStatus = "expired"
)

// BillingCycle defines the recurring interval for billing.
type BillingCycle string

const (
	CycleMonthly   BillingCycle = "monthly"
	CycleQuarterly BillingCycle = "quarterly"
	CycleAnnual    BillingCycle = "annual"
)

// Subscription is the core domain entity.
type Subscription struct {
	ID            string
	CustomerID    string
	PlanID        string
	ProductID     string
	Status        SubStatus
	Cycle         BillingCycle
	Price         float64
	Currency      string
	TrialEndsAt   *time.Time
	NextBillingAt time.Time
	StartedAt     time.Time
	CancelledAt   *time.Time
	CreatedAt     time.Time
}

// BillingRecord captures the result of a billing attempt.
type BillingRecord struct {
	ID             string
	SubscriptionID string
	Amount         float64
	Currency       string
	// Status is one of: "success", "failed", "pending"
	Status    string
	CreatedAt time.Time
}

// ErrNotFound is returned when a requested resource does not exist.
var ErrNotFound = errors.New("not found")
