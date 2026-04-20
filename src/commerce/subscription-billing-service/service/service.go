package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopos/subscription-billing-service/domain"
	"github.com/shopos/subscription-billing-service/store"
)

// SubscribeRequest carries the data needed to create a new subscription.
type SubscribeRequest struct {
	CustomerID  string
	PlanID      string
	ProductID   string
	Cycle       domain.BillingCycle
	Price       float64
	Currency    string
	TrialDays   int
}

// Service provides subscription lifecycle operations.
type Service struct {
	store store.Storer
}

// New constructs a Service with the given Storer.
func New(s store.Storer) *Service {
	return &Service{store: s}
}

// Subscribe creates a new active subscription.
func (svc *Service) Subscribe(req SubscribeRequest) (*domain.Subscription, error) {
	now := time.Now().UTC()

	var trialEndsAt *time.Time
	if req.TrialDays > 0 {
		t := now.AddDate(0, 0, req.TrialDays)
		trialEndsAt = &t
	}

	nextBilling := nextBillingDate(now, req.Cycle)

	sub := &domain.Subscription{
		ID:            uuid.New().String(),
		CustomerID:    req.CustomerID,
		PlanID:        req.PlanID,
		ProductID:     req.ProductID,
		Status:        domain.SubActive,
		Cycle:         req.Cycle,
		Price:         req.Price,
		Currency:      req.Currency,
		TrialEndsAt:   trialEndsAt,
		NextBillingAt: nextBilling,
		StartedAt:     now,
		CreatedAt:     now,
	}

	if err := svc.store.Create(sub); err != nil {
		return nil, fmt.Errorf("create subscription: %w", err)
	}
	return sub, nil
}

// Cancel marks a subscription as cancelled.
func (svc *Service) Cancel(id string) error {
	now := time.Now().UTC()
	if err := svc.store.UpdateStatus(id, domain.SubCancelled, &now); err != nil {
		return fmt.Errorf("cancel subscription %s: %w", id, err)
	}
	return nil
}

// Pause suspends billing for a subscription.
func (svc *Service) Pause(id string) error {
	if err := svc.store.UpdateStatus(id, domain.SubPaused, nil); err != nil {
		return fmt.Errorf("pause subscription %s: %w", id, err)
	}
	return nil
}

// Resume reactivates a paused subscription and recalculates next billing date.
func (svc *Service) Resume(id string) error {
	sub, err := svc.store.Get(id)
	if err != nil {
		return fmt.Errorf("get subscription %s: %w", id, err)
	}

	if err := svc.store.UpdateStatus(id, domain.SubActive, nil); err != nil {
		return fmt.Errorf("resume subscription %s: %w", id, err)
	}

	next := nextBillingDate(time.Now().UTC(), sub.Cycle)
	if err := svc.store.UpdateNextBilling(id, next); err != nil {
		return fmt.Errorf("update next billing for %s: %w", id, err)
	}
	return nil
}

// ProcessBilling creates a billing record for a subscription and advances the billing cycle.
func (svc *Service) ProcessBilling(subID string) (*domain.BillingRecord, error) {
	sub, err := svc.store.Get(subID)
	if err != nil {
		return nil, fmt.Errorf("get subscription for billing %s: %w", subID, err)
	}

	rec := &domain.BillingRecord{
		ID:             uuid.New().String(),
		SubscriptionID: subID,
		Amount:         sub.Price,
		Currency:       sub.Currency,
		Status:         "success",
		CreatedAt:      time.Now().UTC(),
	}

	if err := svc.store.SaveBillingRecord(rec); err != nil {
		rec.Status = "failed"
		return rec, fmt.Errorf("save billing record for %s: %w", subID, err)
	}

	next := nextBillingDate(time.Now().UTC(), sub.Cycle)
	if err := svc.store.UpdateNextBilling(subID, next); err != nil {
		return rec, fmt.Errorf("advance billing cycle for %s: %w", subID, err)
	}

	return rec, nil
}

// GetSubscription retrieves a subscription by ID.
func (svc *Service) GetSubscription(id string) (*domain.Subscription, error) {
	return svc.store.Get(id)
}

// ListSubscriptions returns all subscriptions for a customer.
func (svc *Service) ListSubscriptions(customerID string) ([]*domain.Subscription, error) {
	return svc.store.List(customerID)
}

// ListBillingRecords returns all billing records for a subscription.
func (svc *Service) ListBillingRecords(subID string) ([]*domain.BillingRecord, error) {
	return svc.store.ListBillingRecords(subID)
}

// nextBillingDate computes the next billing date from a reference time.
func nextBillingDate(from time.Time, cycle domain.BillingCycle) time.Time {
	switch cycle {
	case domain.CycleQuarterly:
		return from.AddDate(0, 3, 0)
	case domain.CycleAnnual:
		return from.AddDate(1, 0, 0)
	default: // CycleMonthly
		return from.AddDate(0, 1, 0)
	}
}
