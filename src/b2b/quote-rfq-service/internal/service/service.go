package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopos/quote-rfq-service/internal/domain"
	"github.com/shopos/quote-rfq-service/internal/store"
)

// Servicer exposes all business operations for the quote/RFQ domain.
type Servicer interface {
	CreateRFQ(req CreateRFQRequest) (*domain.Quote, error)
	GetQuote(id uuid.UUID) (*domain.Quote, error)
	ListQuotes(orgID *uuid.UUID, status *domain.QuoteStatus) ([]*domain.Quote, error)
	SubmitRFQ(id uuid.UUID) error
	ReviewQuote(id uuid.UUID) error
	ProvideQuote(id uuid.UUID, req ProvideQuoteRequest) (*domain.Quote, error)
	AcceptQuote(id uuid.UUID) error
	RejectQuote(id uuid.UUID) error
	ExpireQuote(id uuid.UUID) error
	CancelQuote(id uuid.UUID) error
	UpdateNotes(id uuid.UUID, notes string) error
}

// CreateRFQRequest carries the data needed to open a new RFQ.
type CreateRFQRequest struct {
	OrgID             uuid.UUID        `json:"org_id"`
	Title             string           `json:"title"`
	Description       string           `json:"description"`
	Items             domain.QuoteItems `json:"items"`
	RequestedDelivery time.Time        `json:"requested_delivery"`
	Currency          string           `json:"currency"`
	Notes             string           `json:"notes"`
	CreatedBy         string           `json:"created_by"`
}

// ProvideQuoteRequest carries the vendor's response data.
type ProvideQuoteRequest struct {
	Items       domain.QuoteItems `json:"items"`
	TotalAmount float64           `json:"total_amount"`
	ValidUntil  *time.Time        `json:"valid_until,omitempty"`
}

// Service is the concrete implementation of Servicer.
type Service struct {
	store store.Storer
}

// New constructs a Service backed by the supplied Storer.
func New(s store.Storer) *Service {
	return &Service{store: s}
}

// CreateRFQ creates a new quote request in DRAFT status.
func (svc *Service) CreateRFQ(req CreateRFQRequest) (*domain.Quote, error) {
	if req.Title == "" {
		return nil, fmt.Errorf("title is required")
	}
	if req.OrgID == uuid.Nil {
		return nil, fmt.Errorf("org_id is required")
	}
	if req.CreatedBy == "" {
		return nil, fmt.Errorf("created_by is required")
	}
	if len(req.Items) == 0 {
		return nil, fmt.Errorf("at least one item is required")
	}
	if req.Currency == "" {
		req.Currency = "USD"
	}

	now := time.Now().UTC()
	q := &domain.Quote{
		ID:                uuid.New(),
		OrgID:             req.OrgID,
		Title:             req.Title,
		Description:       req.Description,
		Items:             req.Items,
		RequestedDelivery: req.RequestedDelivery,
		Status:            domain.QuoteStatusDraft,
		Currency:          req.Currency,
		Notes:             req.Notes,
		CreatedBy:         req.CreatedBy,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	if err := svc.store.Create(q); err != nil {
		return nil, fmt.Errorf("create RFQ: %w", err)
	}
	return q, nil
}

// GetQuote retrieves a quote by ID.
func (svc *Service) GetQuote(id uuid.UUID) (*domain.Quote, error) {
	return svc.store.Get(id)
}

// ListQuotes lists quotes with optional filters.
func (svc *Service) ListQuotes(orgID *uuid.UUID, status *domain.QuoteStatus) ([]*domain.Quote, error) {
	return svc.store.List(orgID, status)
}

// SubmitRFQ transitions a DRAFT quote to SUBMITTED.
func (svc *Service) SubmitRFQ(id uuid.UUID) error {
	return svc.transition(id, domain.QuoteStatusSubmitted)
}

// ReviewQuote transitions a SUBMITTED quote to UNDER_REVIEW.
func (svc *Service) ReviewQuote(id uuid.UUID) error {
	return svc.transition(id, domain.QuoteStatusUnderReview)
}

// ProvideQuote records vendor pricing and transitions to QUOTED.
func (svc *Service) ProvideQuote(id uuid.UUID, req ProvideQuoteRequest) (*domain.Quote, error) {
	q, err := svc.store.Get(id)
	if err != nil {
		return nil, err
	}
	if !domain.CanTransition(q.Status, domain.QuoteStatusQuoted) {
		return nil, fmt.Errorf("%w: %s → QUOTED", domain.ErrInvalidTransition, q.Status)
	}
	if err := svc.store.SetQuotedPrices(id, req.Items, req.TotalAmount, req.ValidUntil); err != nil {
		return nil, fmt.Errorf("provide quote: %w", err)
	}
	return svc.store.Get(id)
}

// AcceptQuote transitions a QUOTED quote to ACCEPTED.
func (svc *Service) AcceptQuote(id uuid.UUID) error {
	return svc.transition(id, domain.QuoteStatusAccepted)
}

// RejectQuote transitions a QUOTED or UNDER_REVIEW quote to REJECTED.
func (svc *Service) RejectQuote(id uuid.UUID) error {
	return svc.transition(id, domain.QuoteStatusRejected)
}

// ExpireQuote transitions a QUOTED quote to EXPIRED.
func (svc *Service) ExpireQuote(id uuid.UUID) error {
	return svc.transition(id, domain.QuoteStatusExpired)
}

// CancelQuote transitions any non-terminal quote to CANCELLED.
func (svc *Service) CancelQuote(id uuid.UUID) error {
	return svc.transition(id, domain.QuoteStatusCancelled)
}

// UpdateNotes sets a freeform notes string on a quote.
func (svc *Service) UpdateNotes(id uuid.UUID, notes string) error {
	return svc.store.UpdateNotes(id, notes)
}

// transition is a helper that loads a quote, validates the transition, and persists.
func (svc *Service) transition(id uuid.UUID, next domain.QuoteStatus) error {
	q, err := svc.store.Get(id)
	if err != nil {
		return err
	}
	if err := q.Transition(next); err != nil {
		return err
	}
	return svc.store.UpdateStatus(id, next)
}
