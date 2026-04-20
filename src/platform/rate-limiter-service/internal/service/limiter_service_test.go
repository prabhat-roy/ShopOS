package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/shopos/rate-limiter-service/internal/domain"
	"github.com/shopos/rate-limiter-service/internal/service"
	"github.com/shopos/rate-limiter-service/internal/store"
	"go.uber.org/zap"
)

// --- mock policy store ---

type mockPolicyStore struct {
	data map[string]*domain.Policy
}

func newMockPS() *mockPolicyStore {
	return &mockPolicyStore{data: make(map[string]*domain.Policy)}
}

func (m *mockPolicyStore) Save(_ context.Context, p *domain.Policy) error {
	m.data[p.ID] = p
	return nil
}

func (m *mockPolicyStore) Get(_ context.Context, id string) (*domain.Policy, error) {
	p, ok := m.data[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return p, nil
}

func (m *mockPolicyStore) List(_ context.Context) ([]*domain.Policy, error) {
	out := make([]*domain.Policy, 0, len(m.data))
	for _, p := range m.data {
		out = append(out, p)
	}
	return out, nil
}

func (m *mockPolicyStore) Delete(_ context.Context, id string) error {
	if _, ok := m.data[id]; !ok {
		return domain.ErrNotFound
	}
	delete(m.data, id)
	return nil
}

// --- mock counter store ---

type mockCounterStore struct {
	callCount int
	limit     int
}

func (m *mockCounterStore) SlidingWindowAllow(_ context.Context, _ string, limit int, _ time.Duration, cost int) (bool, int64, int64, error) {
	m.callCount += cost
	remaining := int64(limit - m.callCount)
	if remaining < 0 {
		return false, 0, 60, nil
	}
	return true, remaining, 60, nil
}

func (m *mockCounterStore) FixedWindowAllow(_ context.Context, _ string, limit int, _ time.Duration, cost int) (bool, int64, int64, error) {
	m.callCount += cost
	remaining := int64(limit - m.callCount)
	if remaining < 0 {
		return false, 0, 60, nil
	}
	return true, remaining, 60, nil
}

// --- helpers ---

func newSvc() *service.LimiterService {
	return service.New(newMockPS(), &mockCounterStore{}, store.NewInMemoryTokenBucket(), zap.NewNop())
}

// --- tests ---

func TestCreateAndGetPolicy(t *testing.T) {
	svc := newSvc()
	ctx := context.Background()

	p, err := svc.CreatePolicy(ctx, &domain.CreatePolicyRequest{
		Key:        "api:search",
		Name:       "Search API",
		Algorithm:  domain.AlgoSlidingWindow,
		Limit:      100,
		WindowSecs: 60,
		Enabled:    true,
	})
	if err != nil {
		t.Fatalf("CreatePolicy: %v", err)
	}
	if p.ID == "" {
		t.Error("expected non-empty ID")
	}

	got, err := svc.GetPolicy(ctx, p.ID)
	if err != nil {
		t.Fatalf("GetPolicy: %v", err)
	}
	if got.Key != "api:search" {
		t.Errorf("expected key api:search, got %q", got.Key)
	}
}

func TestCreatePolicyInvalidInput(t *testing.T) {
	svc := newSvc()
	_, err := svc.CreatePolicy(context.Background(), &domain.CreatePolicyRequest{Name: "no-key"})
	if err != domain.ErrInvalidInput {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
}

func TestCreatePolicyZeroLimit(t *testing.T) {
	svc := newSvc()
	_, err := svc.CreatePolicy(context.Background(), &domain.CreatePolicyRequest{
		Key: "x", Name: "X", Limit: 0, WindowSecs: 60,
	})
	if err != domain.ErrInvalidInput {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
}

func TestListPolicies(t *testing.T) {
	svc := newSvc()
	ctx := context.Background()

	_, _ = svc.CreatePolicy(ctx, &domain.CreatePolicyRequest{Key: "a", Name: "A", Limit: 10, WindowSecs: 60, Enabled: true})
	_, _ = svc.CreatePolicy(ctx, &domain.CreatePolicyRequest{Key: "b", Name: "B", Limit: 20, WindowSecs: 60, Enabled: true})

	list, err := svc.ListPolicies(ctx)
	if err != nil {
		t.Fatalf("ListPolicies: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("expected 2, got %d", len(list))
	}
}

func TestUpdatePolicy(t *testing.T) {
	svc := newSvc()
	ctx := context.Background()

	p, _ := svc.CreatePolicy(ctx, &domain.CreatePolicyRequest{Key: "x", Name: "X", Limit: 10, WindowSecs: 60, Enabled: false})

	enabled := true
	newLimit := 50
	updated, err := svc.UpdatePolicy(ctx, p.ID, &domain.UpdatePolicyRequest{Enabled: &enabled, Limit: &newLimit})
	if err != nil {
		t.Fatalf("UpdatePolicy: %v", err)
	}
	if !updated.Enabled || updated.Limit != 50 {
		t.Errorf("unexpected result: %+v", updated)
	}
}

func TestDeletePolicy(t *testing.T) {
	svc := newSvc()
	ctx := context.Background()

	p, _ := svc.CreatePolicy(ctx, &domain.CreatePolicyRequest{Key: "del", Name: "Del", Limit: 5, WindowSecs: 60, Enabled: true})
	if err := svc.DeletePolicy(ctx, p.ID); err != nil {
		t.Fatalf("DeletePolicy: %v", err)
	}
	if err := svc.DeletePolicy(ctx, p.ID); err != domain.ErrNotFound {
		t.Errorf("expected ErrNotFound on second delete, got %v", err)
	}
}

func TestCheckSlidingWindowAllowed(t *testing.T) {
	svc := newSvc()
	ctx := context.Background()

	_, _ = svc.CreatePolicy(ctx, &domain.CreatePolicyRequest{
		Key: "svc:login", Name: "Login", Algorithm: domain.AlgoSlidingWindow,
		Limit: 10, WindowSecs: 60, Enabled: true,
	})

	resp, err := svc.Check(ctx, domain.CheckRequest{PolicyKey: "svc:login", Subject: "192.168.1.1", Cost: 1})
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if !resp.Allowed {
		t.Error("expected allowed=true")
	}
}

func TestCheckDisabledPolicyAlwaysAllows(t *testing.T) {
	svc := newSvc()
	ctx := context.Background()

	_, _ = svc.CreatePolicy(ctx, &domain.CreatePolicyRequest{
		Key: "off:route", Name: "Off", Algorithm: domain.AlgoSlidingWindow,
		Limit: 1, WindowSecs: 60, Enabled: false,
	})

	for i := 0; i < 5; i++ {
		resp, err := svc.Check(ctx, domain.CheckRequest{PolicyKey: "off:route", Subject: "u1"})
		if err != nil || !resp.Allowed {
			t.Errorf("disabled policy should always allow; iter %d: %v %v", i, resp, err)
		}
	}
}

func TestCheckPolicyNotFound(t *testing.T) {
	svc := newSvc()
	_, err := svc.Check(context.Background(), domain.CheckRequest{PolicyKey: "missing", Subject: "u1"})
	if err != domain.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestCheckTokenBucket(t *testing.T) {
	svc := newSvc()
	ctx := context.Background()

	_, _ = svc.CreatePolicy(ctx, &domain.CreatePolicyRequest{
		Key: "tb:api", Name: "TB", Algorithm: domain.AlgoTokenBucket,
		Limit: 5, BurstSize: 5, WindowSecs: 1, Enabled: true,
	})

	allowed := 0
	for i := 0; i < 10; i++ {
		resp, _ := svc.Check(ctx, domain.CheckRequest{PolicyKey: "tb:api", Subject: "u1", Cost: 1})
		if resp.Allowed {
			allowed++
		}
	}
	// First 5 should be allowed, next 5 denied
	if allowed != 5 {
		t.Errorf("expected 5 allowed, got %d", allowed)
	}
}

func TestCheckMissingSubject(t *testing.T) {
	svc := newSvc()
	_, err := svc.Check(context.Background(), domain.CheckRequest{PolicyKey: "x"})
	if err != domain.ErrInvalidInput {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
}

func TestTokenBucketCleanup(t *testing.T) {
	tb := store.NewInMemoryTokenBucket()
	tb.Allow("key1", 10, 10, 1)
	tb.Cleanup(0) // cleanup everything older than 0 duration
	// Just verify no panic; actual key removal is internal
}
