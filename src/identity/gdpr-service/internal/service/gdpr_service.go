package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/shopos/gdpr-service/internal/domain"
)

// Storer is the persistence interface required by the service layer.
type Storer interface {
	CreateRequest(ctx context.Context, req *domain.DataRequest) (*domain.DataRequest, error)
	GetRequest(ctx context.Context, id string) (*domain.DataRequest, error)
	ListRequests(ctx context.Context, userID string) ([]*domain.DataRequest, error)
	UpdateRequestStatus(ctx context.Context, id string, status domain.RequestStatus, notes string) error
	UpsertConsent(ctx context.Context, consent *domain.Consent) error
	GetConsents(ctx context.Context, userID string) ([]*domain.Consent, error)
	GetConsent(ctx context.Context, userID string, consentType domain.ConsentType) (*domain.Consent, error)
}

// GDPRService implements all GDPR business operations.
type GDPRService struct {
	store Storer
}

// New creates a new GDPRService backed by the provided Storer.
func New(store Storer) *GDPRService {
	return &GDPRService{store: store}
}

// SubmitRequest creates a new data subject request in pending state.
// reqType must be one of the declared RequestType constants.
func (s *GDPRService) SubmitRequest(ctx context.Context, userID string, reqType domain.RequestType, reason string) (*domain.DataRequest, error) {
	if userID == "" {
		return nil, fmt.Errorf("user_id is required")
	}
	if reqType == "" {
		return nil, fmt.Errorf("type is required")
	}
	switch reqType {
	case domain.RequestExport, domain.RequestErasure, domain.RequestRectification:
		// valid
	default:
		return nil, fmt.Errorf("unknown request type: %s", reqType)
	}

	req := &domain.DataRequest{
		ID:     uuid.NewString(),
		UserID: userID,
		Type:   reqType,
		Status: domain.StatusPending,
		Reason: reason,
		Notes:  "",
	}
	return s.store.CreateRequest(ctx, req)
}

// GetRequest retrieves a single data subject request by ID.
func (s *GDPRService) GetRequest(ctx context.Context, id string) (*domain.DataRequest, error) {
	return s.store.GetRequest(ctx, id)
}

// ListRequests returns all data subject requests for a given user.
func (s *GDPRService) ListRequests(ctx context.Context, userID string) ([]*domain.DataRequest, error) {
	return s.store.ListRequests(ctx, userID)
}

// ProcessRequest transitions a request to the "processing" state.
func (s *GDPRService) ProcessRequest(ctx context.Context, id string, notes string) error {
	if _, err := s.store.GetRequest(ctx, id); err != nil {
		return err
	}
	return s.store.UpdateRequestStatus(ctx, id, domain.StatusProcessing, notes)
}

// CompleteRequest transitions a request to "completed" and records the
// completion timestamp.
func (s *GDPRService) CompleteRequest(ctx context.Context, id string, notes string) error {
	req, err := s.store.GetRequest(ctx, id)
	if err != nil {
		return err
	}
	if req.Status == domain.StatusCompleted {
		return fmt.Errorf("request %s is already completed", id)
	}
	return s.store.UpdateRequestStatus(ctx, id, domain.StatusCompleted, notes)
}

// RejectRequest transitions a request to "rejected" with an explanation.
func (s *GDPRService) RejectRequest(ctx context.Context, id string, notes string) error {
	if _, err := s.store.GetRequest(ctx, id); err != nil {
		return err
	}
	return s.store.UpdateRequestStatus(ctx, id, domain.StatusRejected, notes)
}

// UpdateConsent records or updates a user's consent decision for a specific
// processing purpose. The caller must supply the request IP address for
// the audit trail.
func (s *GDPRService) UpdateConsent(ctx context.Context, userID string, consentType domain.ConsentType, granted bool, ip string) error {
	if userID == "" {
		return fmt.Errorf("user_id is required")
	}
	switch consentType {
	case domain.ConsentMarketing, domain.ConsentAnalytics, domain.ConsentNecessary:
		// valid
	default:
		return fmt.Errorf("unknown consent type: %s", consentType)
	}

	consent := &domain.Consent{
		UserID:    userID,
		Type:      consentType,
		Granted:   granted,
		IPAddress: ip,
		UpdatedAt: time.Now().UTC(),
	}
	return s.store.UpsertConsent(ctx, consent)
}

// GetConsents returns all consent records for a user.
func (s *GDPRService) GetConsents(ctx context.Context, userID string) ([]*domain.Consent, error) {
	return s.store.GetConsents(ctx, userID)
}

// CheckConsent returns whether the user has granted consent for the specified
// processing type. Returns false (not an error) when no record exists.
func (s *GDPRService) CheckConsent(ctx context.Context, userID string, consentType domain.ConsentType) (bool, error) {
	c, err := s.store.GetConsent(ctx, userID, consentType)
	if err == domain.ErrNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return c.Granted, nil
}
