package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopos/session-service/internal/domain"
)

// Storer is the persistence contract required by SessionService.
type Storer interface {
	Create(ctx context.Context, session *domain.Session, ttl time.Duration) error
	Get(ctx context.Context, id string) (*domain.Session, error)
	Touch(ctx context.Context, id string, ttl time.Duration) error
	Delete(ctx context.Context, id string) error
	ListByUser(ctx context.Context, userID string) ([]*domain.Session, error)
	DeleteAllByUser(ctx context.Context, userID string) error
}

// SessionService contains the business logic for session management.
type SessionService struct {
	store      Storer
	sessionTTL time.Duration
}

// New creates a SessionService with the provided store and TTL.
func New(store Storer, ttl time.Duration) *SessionService {
	return &SessionService{store: store, sessionTTL: ttl}
}

// Create opens a new session for the user described in req. A unique ID is
// generated, session metadata is populated, and the record is persisted with
// the configured TTL.
func (s *SessionService) Create(ctx context.Context, req *domain.CreateSessionRequest) (*domain.Session, error) {
	if req.UserID == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	now := time.Now().UTC()
	session := &domain.Session{
		ID:           uuid.New().String(),
		UserID:       req.UserID,
		DeviceInfo:   req.DeviceInfo,
		IPAddress:    req.IPAddress,
		UserAgent:    req.UserAgent,
		CreatedAt:    now,
		LastActiveAt: now,
		ExpiresAt:    now.Add(s.sessionTTL),
	}

	if err := s.store.Create(ctx, session, s.sessionTTL); err != nil {
		return nil, fmt.Errorf("service.Create: %w", err)
	}

	return session, nil
}

// Validate retrieves a session by ID, verifies it has not expired, and then
// extends its TTL (touch). This is the hot path called on every authenticated
// request.
func (s *SessionService) Validate(ctx context.Context, id string) (*domain.Session, error) {
	session, err := s.store.Get(ctx, id)
	if err != nil {
		return nil, err // ErrNotFound propagates as-is
	}

	if time.Now().UTC().After(session.ExpiresAt) {
		// Clean up the expired record opportunistically.
		_ = s.store.Delete(ctx, id)
		return nil, domain.ErrExpired
	}

	// Extend TTL and refresh last-active timestamp.
	if err := s.store.Touch(ctx, id, s.sessionTTL); err != nil {
		// Non-fatal: we already have a valid session object.
		// Log the error but return the session.
		_ = err
	}

	// Re-fetch to get updated LastActiveAt.
	updated, err := s.store.Get(ctx, id)
	if err != nil {
		return session, nil // fall back to original if re-fetch fails
	}

	return updated, nil
}

// Get retrieves a session by ID without touching its TTL.
func (s *SessionService) Get(ctx context.Context, id string) (*domain.Session, error) {
	return s.store.Get(ctx, id)
}

// Delete terminates a single session.
func (s *SessionService) Delete(ctx context.Context, id string) error {
	return s.store.Delete(ctx, id)
}

// ListByUser returns all active sessions for the given user.
func (s *SessionService) ListByUser(ctx context.Context, userID string) ([]*domain.Session, error) {
	return s.store.ListByUser(ctx, userID)
}

// DeleteAllByUser terminates every session belonging to the specified user.
// This is typically called on logout-all or account compromise.
func (s *SessionService) DeleteAllByUser(ctx context.Context, userID string) error {
	return s.store.DeleteAllByUser(ctx, userID)
}
