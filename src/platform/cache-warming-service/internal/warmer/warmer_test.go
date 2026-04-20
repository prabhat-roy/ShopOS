package warmer_test

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/shopos/cache-warming-service/internal/cache"
	"github.com/shopos/cache-warming-service/internal/warmer"
)

// mockCache is an in-memory implementation of cache.Cacher.
type mockCache struct {
	mu      sync.Mutex
	data    map[string]any
	counts  map[string]int64
	expires map[string]time.Duration
}

func newMockCache() *mockCache {
	return &mockCache{
		data:    make(map[string]any),
		counts:  make(map[string]int64),
		expires: make(map[string]time.Duration),
	}
}

func (m *mockCache) Set(_ context.Context, key string, value any, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
	m.expires[key] = ttl
	return nil
}

func (m *mockCache) Incr(_ context.Context, key string) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.counts[key]++
	return m.counts[key], nil
}

func (m *mockCache) Expire(_ context.Context, key string, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.expires[key] = ttl
	return nil
}

var _ cache.Cacher = (*mockCache)(nil)

func marshal(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}

func TestHandleProductViewed(t *testing.T) {
	mc := newMockCache()
	w := warmer.New(mc, 10*time.Minute)

	msg := marshal(map[string]string{"product_id": "prod-1"})
	if err := w.HandleProductViewed(context.Background(), msg); err != nil {
		t.Fatal(err)
	}

	mc.mu.Lock()
	defer mc.mu.Unlock()

	if mc.counts["warm:product:popular:prod-1"] != 1 {
		t.Errorf("expected popular counter 1, got %d", mc.counts["warm:product:popular:prod-1"])
	}
	if _, ok := mc.data["warm:product:pending:prod-1"]; !ok {
		t.Error("expected pending warm key to be set")
	}
}

func TestHandleProductViewed_NoID(t *testing.T) {
	mc := newMockCache()
	w := warmer.New(mc, 10*time.Minute)

	if err := w.HandleProductViewed(context.Background(), marshal(map[string]string{})); err != nil {
		t.Fatal(err)
	}
	if len(mc.data) != 0 {
		t.Error("expected no keys set for empty product_id")
	}
}

func TestHandleProductViewed_BadJSON(t *testing.T) {
	mc := newMockCache()
	w := warmer.New(mc, 10*time.Minute)

	if err := w.HandleProductViewed(context.Background(), []byte("bad")); err == nil {
		t.Error("expected error for bad JSON")
	}
}

func TestHandleCartAbandoned(t *testing.T) {
	mc := newMockCache()
	w := warmer.New(mc, 10*time.Minute)

	msg := marshal(map[string]string{"cart_id": "cart-99"})
	if err := w.HandleCartAbandoned(context.Background(), msg); err != nil {
		t.Fatal(err)
	}

	mc.mu.Lock()
	defer mc.mu.Unlock()

	if _, ok := mc.data["warm:cart:abandoned:cart-99"]; !ok {
		t.Error("expected cart warm key to be set")
	}
}

func TestHandleOrderPlaced(t *testing.T) {
	mc := newMockCache()
	w := warmer.New(mc, 10*time.Minute)

	msg := marshal(map[string]string{"order_id": "ord-42"})
	if err := w.HandleOrderPlaced(context.Background(), msg); err != nil {
		t.Fatal(err)
	}

	mc.mu.Lock()
	defer mc.mu.Unlock()

	if _, ok := mc.data["warm:order:ord-42"]; !ok {
		t.Error("expected order warm key to be set")
	}
}

func TestHandleInventoryLow(t *testing.T) {
	mc := newMockCache()
	w := warmer.New(mc, 10*time.Minute)

	msg := marshal(map[string]string{"product_id": "prod-low"})
	if err := w.HandleInventoryLow(context.Background(), msg); err != nil {
		t.Fatal(err)
	}

	mc.mu.Lock()
	defer mc.mu.Unlock()

	if _, ok := mc.data["warm:inventory:low:prod-low"]; !ok {
		t.Error("expected inventory low warm key to be set")
	}
}

func TestHandleSearchPerformed(t *testing.T) {
	mc := newMockCache()
	w := warmer.New(mc, 10*time.Minute)

	msg := marshal(map[string]string{"query": "blue sneakers"})
	if err := w.HandleSearchPerformed(context.Background(), msg); err != nil {
		t.Fatal(err)
	}

	mc.mu.Lock()
	defer mc.mu.Unlock()

	if mc.counts["warm:search:popular:blue sneakers"] != 1 {
		t.Errorf("expected search counter 1, got %d", mc.counts["warm:search:popular:blue sneakers"])
	}
}

func TestHandleSearchPerformed_Increment(t *testing.T) {
	mc := newMockCache()
	w := warmer.New(mc, 10*time.Minute)

	msg := marshal(map[string]string{"query": "running shoes"})
	for i := 0; i < 3; i++ {
		if err := w.HandleSearchPerformed(context.Background(), msg); err != nil {
			t.Fatal(err)
		}
	}

	mc.mu.Lock()
	defer mc.mu.Unlock()

	if mc.counts["warm:search:popular:running shoes"] != 3 {
		t.Errorf("expected counter 3, got %d", mc.counts["warm:search:popular:running shoes"])
	}
}
