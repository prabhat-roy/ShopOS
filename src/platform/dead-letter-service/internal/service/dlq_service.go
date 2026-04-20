package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/shopos/dead-letter-service/internal/domain"
)

// Storer is the persistence interface required by DLQService.
type Storer interface {
	Save(msg *domain.DeadMessage) error
	Get(id string) (*domain.DeadMessage, error)
	List(topic string, status domain.MessageStatus, limit, offset int) ([]*domain.DeadMessage, error)
	UpdateStatus(id string, status domain.MessageStatus) error
	Stats() (map[string]int64, error)
}

// DLQService implements the business logic for the dead-letter queue.
type DLQService struct {
	store Storer
}

// New creates a DLQService backed by the provided Storer.
func New(store Storer) *DLQService {
	return &DLQService{store: store}
}

// Save persists a new dead-lettered message, generating an ID and timestamps.
func (s *DLQService) Save(msg *domain.DeadMessage) error {
	if msg.ID == "" {
		msg.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	msg.CreatedAt = now
	msg.UpdatedAt = now
	if msg.Status == "" {
		msg.Status = domain.StatusPending
	}
	if err := s.store.Save(msg); err != nil {
		return fmt.Errorf("service.Save: %w", err)
	}
	return nil
}

// Get retrieves a single message by ID.
func (s *DLQService) Get(id string) (*domain.DeadMessage, error) {
	msg, err := s.store.Get(id)
	if err != nil {
		return nil, fmt.Errorf("service.Get: %w", err)
	}
	return msg, nil
}

// List returns a filtered, paginated list of messages.
func (s *DLQService) List(topic string, status domain.MessageStatus, limit, offset int) ([]*domain.DeadMessage, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	msgs, err := s.store.List(topic, status, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("service.List: %w", err)
	}
	return msgs, nil
}

// Retry marks the message as retried, incrementing its retry counter.
func (s *DLQService) Retry(id string) error {
	if err := s.store.UpdateStatus(id, domain.StatusRetried); err != nil {
		return fmt.Errorf("service.Retry: %w", err)
	}
	return nil
}

// Discard marks the message as permanently discarded.
func (s *DLQService) Discard(id string) error {
	if err := s.store.UpdateStatus(id, domain.StatusDiscarded); err != nil {
		return fmt.Errorf("service.Discard: %w", err)
	}
	return nil
}

// Stats returns aggregate counts per status.
func (s *DLQService) Stats() (map[string]int64, error) {
	result, err := s.store.Stats()
	if err != nil {
		return nil, fmt.Errorf("service.Stats: %w", err)
	}
	return result, nil
}
