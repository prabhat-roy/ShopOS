package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/shopos/returns-logistics-service/internal/domain"
	"github.com/shopos/returns-logistics-service/internal/store"
)

// Servicer is the interface exposing all return logistics business operations.
type Servicer interface {
	CreateReturnAuth(orderID, customerID, reason string, items []domain.ReturnItem) (*domain.ReturnAuth, error)
	GetReturnAuth(id string) (*domain.ReturnAuth, error)
	ListReturnAuths(customerID string) ([]*domain.ReturnAuth, error)
	ApproveReturn(id string) (*domain.ReturnAuth, error)
	RejectReturn(id, reason string) (*domain.ReturnAuth, error)
	IssueLabel(id string) (*domain.ReturnAuth, error)
	MarkInTransit(id string) (*domain.ReturnAuth, error)
	MarkReceived(id string) (*domain.ReturnAuth, error)
	StartInspection(id string) (*domain.ReturnAuth, error)
	CompleteReturn(id, notes string) (*domain.ReturnAuth, error)
	Cancel(id string) (*domain.ReturnAuth, error)
}

// ReturnService implements Servicer.
type ReturnService struct {
	store store.Storer
}

// New creates a new ReturnService backed by the supplied store.
func New(s store.Storer) *ReturnService {
	return &ReturnService{store: s}
}

// CreateReturnAuth creates a new return authorisation in PENDING state.
func (svc *ReturnService) CreateReturnAuth(orderID, customerID, reason string, items []domain.ReturnItem) (*domain.ReturnAuth, error) {
	if orderID == "" || customerID == "" {
		return nil, domain.ErrInvalidRequest
	}
	if len(items) == 0 {
		return nil, domain.ErrInvalidRequest
	}

	now := time.Now().UTC()
	ra := &domain.ReturnAuth{
		ID:         uuid.New().String(),
		OrderID:    orderID,
		CustomerID: customerID,
		Items:      items,
		Reason:     reason,
		Status:     domain.StatusPending,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := svc.store.Create(ra); err != nil {
		return nil, fmt.Errorf("create return auth: %w", err)
	}
	return ra, nil
}

// GetReturnAuth retrieves a return authorisation by ID.
func (svc *ReturnService) GetReturnAuth(id string) (*domain.ReturnAuth, error) {
	return svc.store.Get(id)
}

// ListReturnAuths lists all return authorisations, optionally filtered by customer.
func (svc *ReturnService) ListReturnAuths(customerID string) ([]*domain.ReturnAuth, error) {
	return svc.store.List(customerID)
}

// ApproveReturn transitions an authorisation from PENDING → APPROVED.
func (svc *ReturnService) ApproveReturn(id string) (*domain.ReturnAuth, error) {
	return svc.transition(id, domain.StatusApproved, func(ra *domain.ReturnAuth) error {
		return svc.store.UpdateStatus(id, domain.StatusApproved)
	})
}

// RejectReturn transitions an authorisation from PENDING → REJECTED and records the reason.
func (svc *ReturnService) RejectReturn(id, reason string) (*domain.ReturnAuth, error) {
	return svc.transition(id, domain.StatusRejected, func(ra *domain.ReturnAuth) error {
		if err := svc.store.SetRejectionReason(id, reason); err != nil {
			return err
		}
		return svc.store.UpdateStatus(id, domain.StatusRejected)
	})
}

// IssueLabel generates a return label + tracking number and records them,
// transitioning the authorisation to LABEL_ISSUED.
func (svc *ReturnService) IssueLabel(id string) (*domain.ReturnAuth, error) {
	ra, err := svc.store.Get(id)
	if err != nil {
		return nil, err
	}
	if !domain.CanTransition(ra.Status, domain.StatusLabelIssued) {
		return nil, fmt.Errorf("%w: %s → %s", domain.ErrInvalidTransition, ra.Status, domain.StatusLabelIssued)
	}

	trackingNumber := generateReturnTrackingNumber()
	labelURL := fmt.Sprintf("https://returns.shopos.internal/labels/%s.pdf", trackingNumber)

	if err := svc.store.IssueLabel(id, labelURL, trackingNumber); err != nil {
		return nil, fmt.Errorf("issue label: %w", err)
	}
	return svc.store.Get(id)
}

// MarkInTransit transitions an authorisation from LABEL_ISSUED → IN_TRANSIT.
func (svc *ReturnService) MarkInTransit(id string) (*domain.ReturnAuth, error) {
	return svc.transition(id, domain.StatusInTransit, func(_ *domain.ReturnAuth) error {
		return svc.store.UpdateStatus(id, domain.StatusInTransit)
	})
}

// MarkReceived transitions an authorisation from IN_TRANSIT → RECEIVED.
func (svc *ReturnService) MarkReceived(id string) (*domain.ReturnAuth, error) {
	return svc.transition(id, domain.StatusReceived, func(_ *domain.ReturnAuth) error {
		return svc.store.UpdateStatus(id, domain.StatusReceived)
	})
}

// StartInspection transitions an authorisation from RECEIVED → INSPECTING.
func (svc *ReturnService) StartInspection(id string) (*domain.ReturnAuth, error) {
	return svc.transition(id, domain.StatusInspecting, func(_ *domain.ReturnAuth) error {
		return svc.store.UpdateStatus(id, domain.StatusInspecting)
	})
}

// CompleteReturn records inspection notes, transitions to COMPLETED.
func (svc *ReturnService) CompleteReturn(id, notes string) (*domain.ReturnAuth, error) {
	ra, err := svc.store.Get(id)
	if err != nil {
		return nil, err
	}
	if !domain.CanTransition(ra.Status, domain.StatusCompleted) {
		return nil, fmt.Errorf("%w: %s → %s", domain.ErrInvalidTransition, ra.Status, domain.StatusCompleted)
	}
	if err := svc.store.SetInspectionNotes(id, notes, ra.WarehouseID); err != nil {
		return nil, fmt.Errorf("set inspection notes: %w", err)
	}
	if err := svc.store.UpdateStatus(id, domain.StatusCompleted); err != nil {
		return nil, err
	}
	return svc.store.Get(id)
}

// Cancel cancels a return authorisation (allowed from PENDING, APPROVED, LABEL_ISSUED).
func (svc *ReturnService) Cancel(id string) (*domain.ReturnAuth, error) {
	return svc.transition(id, domain.StatusCancelled, func(_ *domain.ReturnAuth) error {
		return svc.store.UpdateStatus(id, domain.StatusCancelled)
	})
}

// ─── helpers ─────────────────────────────────────────────────────────────────

// transition is a generic helper that loads the RA, validates the transition,
// executes the supplied mutation function, then returns the refreshed RA.
func (svc *ReturnService) transition(
	id string,
	next domain.ReturnAuthStatus,
	mutate func(*domain.ReturnAuth) error,
) (*domain.ReturnAuth, error) {
	ra, err := svc.store.Get(id)
	if err != nil {
		return nil, err
	}
	if !domain.CanTransition(ra.Status, next) {
		return nil, fmt.Errorf("%w: %s → %s", domain.ErrInvalidTransition, ra.Status, next)
	}
	if err := mutate(ra); err != nil {
		return nil, err
	}
	return svc.store.Get(id)
}

// generateReturnTrackingNumber creates a tracking number in the format RET-XXXXXXXX.
func generateReturnTrackingNumber() string {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("RET-%d", time.Now().UnixNano())
	}
	return "RET-" + strings.ToUpper(hex.EncodeToString(b))
}
