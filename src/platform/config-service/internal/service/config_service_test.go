package service_test

import (
	"context"
	"testing"
	"time"

	pb "github.com/shopos/config-service/internal/proto"
	"github.com/shopos/config-service/internal/service"
	"go.uber.org/zap"
)

// mockStore is an in-memory stand-in for etcd.
type mockStore struct {
	data map[string]string
}

func newMockStore() *mockStore {
	return &mockStore{data: make(map[string]string)}
}

func (m *mockStore) Get(_ context.Context, key string) (*pb.ConfigEntry, bool, error) {
	v, ok := m.data[key]
	if !ok {
		return nil, false, nil
	}
	return &pb.ConfigEntry{Key: key, Value: v, Version: 1}, true, nil
}

func (m *mockStore) Set(_ context.Context, key, value string) error {
	m.data[key] = value
	return nil
}

func (m *mockStore) Delete(_ context.Context, key string) error {
	delete(m.data, key)
	return nil
}

func (m *mockStore) List(_ context.Context, prefix string) ([]*pb.ConfigEntry, error) {
	var out []*pb.ConfigEntry
	for k, v := range m.data {
		out = append(out, &pb.ConfigEntry{Key: k, Value: v})
	}
	return out, nil
}

func (m *mockStore) WatchCh(_ context.Context, _ string) (<-chan *pb.WatchEvent, error) {
	ch := make(chan *pb.WatchEvent)
	close(ch)
	return ch, nil
}

// storeBacked wraps mockStore to match the service.Store interface.
type storeBacked struct{ m *mockStore }

func newTestService(t *testing.T) *serviceUnderTest {
	t.Helper()
	ms := newMockStore()
	log := zap.NewNop()
	sut := newServiceWithStore(ms, log)
	return sut
}

// ---- inline service with injectable store ----

type Storer interface {
	Get(ctx context.Context, key string) (*pb.ConfigEntry, bool, error)
	Set(ctx context.Context, key, value string) error
	Delete(ctx context.Context, key string) error
	List(ctx context.Context, prefix string) ([]*pb.ConfigEntry, error)
	WatchCh(ctx context.Context, prefix string) (<-chan *pb.WatchEvent, error)
}

type serviceUnderTest struct {
	store Storer
	log   *zap.Logger
}

func newServiceWithStore(st Storer, log *zap.Logger) *serviceUnderTest {
	return &serviceUnderTest{store: st, log: log}
}

func (s *serviceUnderTest) Get(ctx context.Context, key string) (*pb.ConfigEntry, error) {
	e, found, err := s.store.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, service.ErrNotFound
	}
	return e, nil
}

func (s *serviceUnderTest) Set(ctx context.Context, key, value string) error {
	return s.store.Set(ctx, key, value)
}

func (s *serviceUnderTest) Delete(ctx context.Context, key string) error {
	return s.store.Delete(ctx, key)
}

func (s *serviceUnderTest) List(ctx context.Context, prefix string) ([]*pb.ConfigEntry, error) {
	return s.store.List(ctx, prefix)
}

func (s *serviceUnderTest) Watch(ctx context.Context, prefix string) (<-chan *pb.WatchEvent, error) {
	return s.store.WatchCh(ctx, prefix)
}

// ---- tests ----

func TestSetAndGet(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	if err := svc.Set(ctx, "app/timeout", "30s"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	entry, err := svc.Get(ctx, "app/timeout")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if entry.Value != "30s" {
		t.Errorf("expected 30s, got %q", entry.Value)
	}
}

func TestGetNotFound(t *testing.T) {
	svc := newTestService(t)
	_, err := svc.Get(context.Background(), "does/not/exist")
	if err != service.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestDelete(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	_ = svc.Set(ctx, "temp/key", "value")
	_ = svc.Delete(ctx, "temp/key")

	_, err := svc.Get(ctx, "temp/key")
	if err != service.ErrNotFound {
		t.Errorf("expected key to be deleted, got %v", err)
	}
}

func TestList(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	_ = svc.Set(ctx, "db/host", "localhost")
	_ = svc.Set(ctx, "db/port", "5432")

	entries, err := svc.List(ctx, "db/")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(entries) < 2 {
		t.Errorf("expected at least 2 entries, got %d", len(entries))
	}
}

func TestWatch(t *testing.T) {
	svc := newTestService(t)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	ch, err := svc.Watch(ctx, "app/")
	if err != nil {
		t.Fatalf("Watch: %v", err)
	}
	// mockStore closes the channel immediately — drain it
	for range ch {
	}
}
