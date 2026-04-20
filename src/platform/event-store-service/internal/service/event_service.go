package service

import (
	"context"
	"time"

	"github.com/shopos/event-store-service/internal/domain"
	"go.uber.org/zap"
)

// Storer is satisfied by store.EventStore.
type Storer interface {
	Append(ctx context.Context, req *domain.AppendRequest) ([]*domain.Event, error)
	Read(ctx context.Context, req domain.ReadRequest) ([]*domain.Event, error)
	ReadAll(ctx context.Context, req domain.ReadAllRequest) ([]*domain.Event, error)
	GetByID(ctx context.Context, id string) (*domain.Event, error)
	SaveSnapshot(ctx context.Context, snap *domain.Snapshot) error
	GetSnapshot(ctx context.Context, streamID string) (*domain.Snapshot, error)
}

// EventService contains all event-store business logic.
type EventService struct {
	store Storer
	log   *zap.Logger
}

func New(st Storer, log *zap.Logger) *EventService {
	return &EventService{store: st, log: log}
}

// Append validates and delegates to the store.
func (s *EventService) Append(ctx context.Context, req *domain.AppendRequest) ([]*domain.Event, error) {
	if req.StreamID == "" || req.StreamType == "" {
		return nil, domain.ErrInvalidInput
	}
	if len(req.Events) == 0 {
		return nil, domain.ErrInvalidInput
	}
	for _, e := range req.Events {
		if e.EventType == "" || len(e.Payload) == 0 {
			return nil, domain.ErrInvalidInput
		}
	}

	events, err := s.store.Append(ctx, req)
	if err != nil {
		s.log.Error("append failed",
			zap.String("stream_id", req.StreamID),
			zap.Error(err),
		)
		return nil, err
	}
	s.log.Info("events appended",
		zap.String("stream_id", req.StreamID),
		zap.Int("count", len(events)),
	)
	return events, nil
}

// ReadStream returns ordered events from a single stream.
func (s *EventService) ReadStream(ctx context.Context, req domain.ReadRequest) ([]*domain.Event, error) {
	if req.StreamID == "" {
		return nil, domain.ErrInvalidInput
	}
	return s.store.Read(ctx, req)
}

// ReadAll returns globally ordered events with optional type filters.
func (s *EventService) ReadAll(ctx context.Context, req domain.ReadAllRequest) ([]*domain.Event, error) {
	return s.store.ReadAll(ctx, req)
}

// GetEvent retrieves a single event by ID.
func (s *EventService) GetEvent(ctx context.Context, id string) (*domain.Event, error) {
	if id == "" {
		return nil, domain.ErrInvalidInput
	}
	return s.store.GetByID(ctx, id)
}

// SaveSnapshot stores the projected state of a stream at a version.
func (s *EventService) SaveSnapshot(ctx context.Context, streamID, streamType string, version int64, state []byte) error {
	if streamID == "" || version < 0 || len(state) == 0 {
		return domain.ErrInvalidInput
	}
	return s.store.SaveSnapshot(ctx, &domain.Snapshot{
		StreamID:   streamID,
		StreamType: streamType,
		Version:    version,
		State:      state,
		CreatedAt:  time.Now().UTC(),
	})
}

// GetSnapshot retrieves the latest snapshot for a stream.
func (s *EventService) GetSnapshot(ctx context.Context, streamID string) (*domain.Snapshot, error) {
	if streamID == "" {
		return nil, domain.ErrInvalidInput
	}
	return s.store.GetSnapshot(ctx, streamID)
}
