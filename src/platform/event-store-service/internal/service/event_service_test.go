package service_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/shopos/event-store-service/internal/domain"
	"github.com/shopos/event-store-service/internal/service"
	"go.uber.org/zap"
)

// --- in-memory mock store ---

type memStore struct {
	events    []*domain.Event
	snapshots map[string]*domain.Snapshot
	seq       int64
}

func newMem() *memStore {
	return &memStore{snapshots: make(map[string]*domain.Snapshot)}
}

func (m *memStore) Append(_ context.Context, req *domain.AppendRequest) ([]*domain.Event, error) {
	// find current max version for stream
	var current int64 = -1
	for _, e := range m.events {
		if e.StreamID == req.StreamID && e.Version > current {
			current = e.Version
		}
	}
	if req.ExpectedVersion >= 0 && current != req.ExpectedVersion {
		return nil, fmt.Errorf("%w: expected %d current %d", domain.ErrVersionConflict, req.ExpectedVersion, current)
	}

	now := time.Now().UTC()
	var out []*domain.Event
	for i, ne := range req.Events {
		m.seq++
		ev := &domain.Event{
			ID:         fmt.Sprintf("ev-%d", m.seq),
			StreamID:   req.StreamID,
			StreamType: req.StreamType,
			EventType:  ne.EventType,
			Version:    current + int64(i) + 1,
			GlobalSeq:  m.seq,
			Payload:    ne.Payload,
			Metadata:   ne.Metadata,
			OccurredAt: now,
			RecordedAt: now,
		}
		m.events = append(m.events, ev)
		out = append(out, ev)
	}
	return out, nil
}

func (m *memStore) Read(_ context.Context, req domain.ReadRequest) ([]*domain.Event, error) {
	var out []*domain.Event
	for _, e := range m.events {
		if e.StreamID != req.StreamID || e.Version < req.FromVersion {
			continue
		}
		if req.ToVersion > 0 && e.Version > req.ToVersion {
			continue
		}
		out = append(out, e)
		if req.MaxCount > 0 && len(out) >= req.MaxCount {
			break
		}
	}
	return out, nil
}

func (m *memStore) ReadAll(_ context.Context, req domain.ReadAllRequest) ([]*domain.Event, error) {
	var out []*domain.Event
	for _, e := range m.events {
		if e.GlobalSeq < req.FromGlobalSeq {
			continue
		}
		if req.StreamType != "" && e.StreamType != req.StreamType {
			continue
		}
		if req.EventType != "" && e.EventType != req.EventType {
			continue
		}
		out = append(out, e)
		if req.MaxCount > 0 && len(out) >= req.MaxCount {
			break
		}
	}
	return out, nil
}

func (m *memStore) GetByID(_ context.Context, id string) (*domain.Event, error) {
	for _, e := range m.events {
		if e.ID == id {
			return e, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (m *memStore) SaveSnapshot(_ context.Context, snap *domain.Snapshot) error {
	m.snapshots[snap.StreamID] = snap
	return nil
}

func (m *memStore) GetSnapshot(_ context.Context, streamID string) (*domain.Snapshot, error) {
	s, ok := m.snapshots[streamID]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return s, nil
}

func newSvc() *service.EventService {
	return service.New(newMem(), zap.NewNop())
}

// --- tests ---

func TestAppendAndRead(t *testing.T) {
	svc := newSvc()
	ctx := context.Background()

	events, err := svc.Append(ctx, &domain.AppendRequest{
		StreamID:        "order-1",
		StreamType:      "Order",
		ExpectedVersion: -1,
		Events: []domain.NewEvent{
			{EventType: "OrderPlaced", Payload: []byte(`{"amount":100}`)},
			{EventType: "OrderConfirmed", Payload: []byte(`{"confirmed":true}`)},
		},
	})
	if err != nil {
		t.Fatalf("Append: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	if events[0].Version != 0 || events[1].Version != 1 {
		t.Errorf("unexpected versions: %d, %d", events[0].Version, events[1].Version)
	}

	read, err := svc.ReadStream(ctx, domain.ReadRequest{StreamID: "order-1"})
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if len(read) != 2 {
		t.Errorf("expected 2, got %d", len(read))
	}
}

func TestOptimisticConcurrencyConflict(t *testing.T) {
	svc := newSvc()
	ctx := context.Background()

	svc.Append(ctx, &domain.AppendRequest{
		StreamID: "order-2", StreamType: "Order", ExpectedVersion: -1,
		Events: []domain.NewEvent{{EventType: "OrderPlaced", Payload: []byte(`{}`)}},
	})

	// Try to append again with wrong expected version
	_, err := svc.Append(ctx, &domain.AppendRequest{
		StreamID: "order-2", StreamType: "Order", ExpectedVersion: 5, // wrong
		Events: []domain.NewEvent{{EventType: "OrderUpdated", Payload: []byte(`{}`)}},
	})
	if err == nil {
		t.Fatal("expected version conflict error")
	}
}

func TestAppendInvalidInput(t *testing.T) {
	svc := newSvc()
	cases := []domain.AppendRequest{
		{StreamID: "", StreamType: "Order", Events: []domain.NewEvent{{EventType: "X", Payload: []byte(`{}`)}}},
		{StreamID: "s1", StreamType: "", Events: []domain.NewEvent{{EventType: "X", Payload: []byte(`{}`)}}},
		{StreamID: "s1", StreamType: "T", Events: nil},
		{StreamID: "s1", StreamType: "T", Events: []domain.NewEvent{{EventType: "", Payload: []byte(`{}`)}}},
		{StreamID: "s1", StreamType: "T", Events: []domain.NewEvent{{EventType: "X", Payload: nil}}},
	}
	for _, req := range cases {
		_, err := svc.Append(context.Background(), &req)
		if err != domain.ErrInvalidInput {
			t.Errorf("expected ErrInvalidInput for %+v, got %v", req, err)
		}
	}
}

func TestReadStreamFromVersion(t *testing.T) {
	svc := newSvc()
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		svc.Append(ctx, &domain.AppendRequest{
			StreamID: "order-3", StreamType: "Order", ExpectedVersion: int64(i - 1),
			Events: []domain.NewEvent{{EventType: "Event", Payload: []byte(`{}`)}},
		})
	}

	events, _ := svc.ReadStream(ctx, domain.ReadRequest{StreamID: "order-3", FromVersion: 2, ToVersion: 4})
	if len(events) != 3 {
		t.Errorf("expected 3 events (v2,v3,v4), got %d", len(events))
	}
}

func TestReadAll(t *testing.T) {
	svc := newSvc()
	ctx := context.Background()

	svc.Append(ctx, &domain.AppendRequest{
		StreamID: "order-4", StreamType: "Order", ExpectedVersion: -1,
		Events: []domain.NewEvent{
			{EventType: "OrderPlaced", Payload: []byte(`{}`)},
			{EventType: "OrderPaid", Payload: []byte(`{}`)},
		},
	})
	svc.Append(ctx, &domain.AppendRequest{
		StreamID: "product-1", StreamType: "Product", ExpectedVersion: -1,
		Events: []domain.NewEvent{{EventType: "ProductUpdated", Payload: []byte(`{}`)}},
	})

	// All events
	all, _ := svc.ReadAll(ctx, domain.ReadAllRequest{FromGlobalSeq: 0})
	if len(all) != 3 {
		t.Errorf("expected 3 total events, got %d", len(all))
	}

	// Filter by stream type
	orders, _ := svc.ReadAll(ctx, domain.ReadAllRequest{StreamType: "Order"})
	if len(orders) != 2 {
		t.Errorf("expected 2 order events, got %d", len(orders))
	}

	// Filter by event type
	placed, _ := svc.ReadAll(ctx, domain.ReadAllRequest{EventType: "OrderPlaced"})
	if len(placed) != 1 {
		t.Errorf("expected 1 OrderPlaced, got %d", len(placed))
	}
}

func TestGetEvent(t *testing.T) {
	svc := newSvc()
	ctx := context.Background()

	events, _ := svc.Append(ctx, &domain.AppendRequest{
		StreamID: "order-5", StreamType: "Order", ExpectedVersion: -1,
		Events: []domain.NewEvent{{EventType: "OrderPlaced", Payload: []byte(`{}`)}},
	})

	ev, err := svc.GetEvent(ctx, events[0].ID)
	if err != nil {
		t.Fatalf("GetEvent: %v", err)
	}
	if ev.ID != events[0].ID {
		t.Errorf("ID mismatch")
	}
}

func TestGetEventNotFound(t *testing.T) {
	svc := newSvc()
	_, err := svc.GetEvent(context.Background(), "nonexistent")
	if err != domain.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestSnapshot(t *testing.T) {
	svc := newSvc()
	ctx := context.Background()

	state := []byte(`{"total":500,"status":"active"}`)
	if err := svc.SaveSnapshot(ctx, "order-6", "Order", 10, state); err != nil {
		t.Fatalf("SaveSnapshot: %v", err)
	}

	snap, err := svc.GetSnapshot(ctx, "order-6")
	if err != nil {
		t.Fatalf("GetSnapshot: %v", err)
	}
	if snap.Version != 10 {
		t.Errorf("expected version 10, got %d", snap.Version)
	}
}

func TestSnapshotNotFound(t *testing.T) {
	svc := newSvc()
	_, err := svc.GetSnapshot(context.Background(), "no-stream")
	if err != domain.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
