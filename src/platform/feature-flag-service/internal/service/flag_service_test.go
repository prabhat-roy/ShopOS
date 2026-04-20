package service_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/shopos/feature-flag-service/internal/domain"
	"github.com/shopos/feature-flag-service/internal/service"
	"go.uber.org/zap"
)

// mockStore is an in-memory implementation of service.Storer.
type mockStore struct {
	flags map[string]*domain.Flag
	seq   int
}

func newMock() *mockStore {
	return &mockStore{flags: make(map[string]*domain.Flag)}
}

func (m *mockStore) GetByKey(_ context.Context, key string) (*domain.Flag, error) {
	for _, f := range m.flags {
		if f.Key == key {
			return f, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (m *mockStore) GetByID(_ context.Context, id string) (*domain.Flag, error) {
	f, ok := m.flags[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return f, nil
}

func (m *mockStore) List(_ context.Context) ([]*domain.Flag, error) {
	out := make([]*domain.Flag, 0, len(m.flags))
	for _, f := range m.flags {
		out = append(out, f)
	}
	return out, nil
}

func (m *mockStore) Create(_ context.Context, req *domain.CreateFlagRequest) (*domain.Flag, error) {
	for _, f := range m.flags {
		if f.Key == req.Key {
			return nil, domain.ErrAlreadyExists
		}
	}
	m.seq++
	id := fmt.Sprintf("flag-%d", m.seq)
	f := &domain.Flag{
		ID:          id,
		Key:         req.Key,
		Name:        req.Name,
		Description: req.Description,
		Enabled:     req.Enabled,
		Strategy:    req.Strategy,
		Percentage:  req.Percentage,
		UserIDs:     req.UserIDs,
		ContextKey:  req.ContextKey,
		ContextVal:  req.ContextVal,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	m.flags[id] = f
	return f, nil
}

func (m *mockStore) Update(_ context.Context, id string, req *domain.UpdateFlagRequest) (*domain.Flag, error) {
	f, ok := m.flags[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	if req.Name != nil {
		f.Name = *req.Name
	}
	if req.Enabled != nil {
		f.Enabled = *req.Enabled
	}
	if req.Strategy != nil {
		f.Strategy = *req.Strategy
	}
	if req.Percentage != nil {
		f.Percentage = *req.Percentage
	}
	if req.UserIDs != nil {
		f.UserIDs = req.UserIDs
	}
	f.UpdatedAt = time.Now()
	return f, nil
}

func (m *mockStore) Delete(_ context.Context, id string) error {
	if _, ok := m.flags[id]; !ok {
		return domain.ErrNotFound
	}
	delete(m.flags, id)
	return nil
}

func TestCreateAndGetFlag(t *testing.T) {
	svc := service.New(newMock(), zap.NewNop())
	ctx := context.Background()

	f, err := svc.CreateFlag(ctx, &domain.CreateFlagRequest{
		Key:     "dark-mode",
		Name:    "Dark Mode",
		Enabled: true,
		Strategy: domain.StrategyAll,
	})
	if err != nil {
		t.Fatalf("CreateFlag: %v", err)
	}
	if f.Key != "dark-mode" {
		t.Errorf("expected key dark-mode, got %q", f.Key)
	}

	got, err := svc.GetFlag(ctx, "dark-mode")
	if err != nil {
		t.Fatalf("GetFlag: %v", err)
	}
	if got.ID != f.ID {
		t.Errorf("ID mismatch")
	}
}

func TestCreateFlagMissingKey(t *testing.T) {
	svc := service.New(newMock(), zap.NewNop())
	_, err := svc.CreateFlag(context.Background(), &domain.CreateFlagRequest{Name: "No Key"})
	if err != domain.ErrInvalidInput {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
}

func TestGetFlagNotFound(t *testing.T) {
	svc := service.New(newMock(), zap.NewNop())
	_, err := svc.GetFlag(context.Background(), "missing")
	if err != domain.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestListFlags(t *testing.T) {
	m := newMock()
	svc := service.New(m, zap.NewNop())
	ctx := context.Background()

	_, _ = svc.CreateFlag(ctx, &domain.CreateFlagRequest{Key: "a", Name: "A", Strategy: domain.StrategyAll})
	_, _ = svc.CreateFlag(ctx, &domain.CreateFlagRequest{Key: "b", Name: "B", Strategy: domain.StrategyAll})

	flags, err := svc.ListFlags(ctx)
	if err != nil {
		t.Fatalf("ListFlags: %v", err)
	}
	if len(flags) != 2 {
		t.Errorf("expected 2, got %d", len(flags))
	}
}

func TestUpdateFlag(t *testing.T) {
	m := newMock()
	svc := service.New(m, zap.NewNop())
	ctx := context.Background()

	f, _ := svc.CreateFlag(ctx, &domain.CreateFlagRequest{Key: "x", Name: "X", Strategy: domain.StrategyAll})

	newName := "X Updated"
	enabled := true
	updated, err := svc.UpdateFlag(ctx, f.ID, &domain.UpdateFlagRequest{
		Name:    &newName,
		Enabled: &enabled,
	})
	if err != nil {
		t.Fatalf("UpdateFlag: %v", err)
	}
	if updated.Name != "X Updated" || !updated.Enabled {
		t.Errorf("unexpected update result: %+v", updated)
	}
}

func TestDeleteFlag(t *testing.T) {
	m := newMock()
	svc := service.New(m, zap.NewNop())
	ctx := context.Background()

	f, _ := svc.CreateFlag(ctx, &domain.CreateFlagRequest{Key: "del", Name: "Del", Strategy: domain.StrategyAll})
	if err := svc.DeleteFlag(ctx, f.ID); err != nil {
		t.Fatalf("DeleteFlag: %v", err)
	}
	if err := svc.DeleteFlag(ctx, f.ID); err != domain.ErrNotFound {
		t.Errorf("expected ErrNotFound on second delete, got %v", err)
	}
}

func TestEvaluateStrategyAll(t *testing.T) {
	m := newMock()
	svc := service.New(m, zap.NewNop())
	ctx := context.Background()

	_, _ = svc.CreateFlag(ctx, &domain.CreateFlagRequest{Key: "all-flag", Name: "All", Enabled: true, Strategy: domain.StrategyAll})

	ok, err := svc.Evaluate(ctx, domain.EvalRequest{Key: "all-flag", UserID: "u1"})
	if err != nil || !ok {
		t.Errorf("expected true, got %v %v", ok, err)
	}
}

func TestEvaluateDisabled(t *testing.T) {
	m := newMock()
	svc := service.New(m, zap.NewNop())
	ctx := context.Background()

	_, _ = svc.CreateFlag(ctx, &domain.CreateFlagRequest{Key: "off", Name: "Off", Enabled: false, Strategy: domain.StrategyAll})

	ok, err := svc.Evaluate(ctx, domain.EvalRequest{Key: "off"})
	if err != nil || ok {
		t.Errorf("expected false for disabled flag")
	}
}

func TestEvaluateUserList(t *testing.T) {
	m := newMock()
	svc := service.New(m, zap.NewNop())
	ctx := context.Background()

	_, _ = svc.CreateFlag(ctx, &domain.CreateFlagRequest{
		Key:      "beta",
		Name:     "Beta",
		Enabled:  true,
		Strategy: domain.StrategyUserList,
		UserIDs:  []string{"u1", "u2"},
	})

	if ok, _ := svc.Evaluate(ctx, domain.EvalRequest{Key: "beta", UserID: "u1"}); !ok {
		t.Error("u1 should be in user list")
	}
	if ok, _ := svc.Evaluate(ctx, domain.EvalRequest{Key: "beta", UserID: "u99"}); ok {
		t.Error("u99 should NOT be in user list")
	}
}

func TestEvaluateContext(t *testing.T) {
	m := newMock()
	svc := service.New(m, zap.NewNop())
	ctx := context.Background()

	ck := "region"
	cv := "eu"
	_, _ = svc.CreateFlag(ctx, &domain.CreateFlagRequest{
		Key:        "eu-feature",
		Name:       "EU Only",
		Enabled:    true,
		Strategy:   domain.StrategyContext,
		ContextKey: ck,
		ContextVal: cv,
	})

	if ok, _ := svc.Evaluate(ctx, domain.EvalRequest{Key: "eu-feature", Context: map[string]string{"region": "eu"}}); !ok {
		t.Error("eu context should match")
	}
	if ok, _ := svc.Evaluate(ctx, domain.EvalRequest{Key: "eu-feature", Context: map[string]string{"region": "us"}}); ok {
		t.Error("us context should not match")
	}
}

func TestEvaluatePercentage(t *testing.T) {
	m := newMock()
	svc := service.New(m, zap.NewNop())
	ctx := context.Background()

	pct := 50
	_, _ = svc.CreateFlag(ctx, &domain.CreateFlagRequest{
		Key:        "gradual",
		Name:       "Gradual",
		Enabled:    true,
		Strategy:   domain.StrategyPercentage,
		Percentage: pct,
	})

	// Run 100 different user IDs and check that roughly half get true
	trueCount := 0
	for i := 0; i < 100; i++ {
		uid := fmt.Sprintf("user-%d", i)
		if ok, _ := svc.Evaluate(ctx, domain.EvalRequest{Key: "gradual", UserID: uid}); ok {
			trueCount++
		}
	}
	// Expect between 30–70 out of 100 (FNV hash is deterministic but not perfectly uniform at 100 samples)
	if trueCount < 20 || trueCount > 80 {
		t.Errorf("percentage rollout looks off: %d/100 enabled", trueCount)
	}
}
