package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/shopos/saga-orchestrator/internal/domain"
	"github.com/shopos/saga-orchestrator/internal/service"
	"go.uber.org/zap"
)

// --- mock store ---

type mockStore struct {
	sagas map[string]*domain.Saga
}

func newMockStore() *mockStore { return &mockStore{sagas: make(map[string]*domain.Saga)} }

func (m *mockStore) Create(_ context.Context, s *domain.Saga) error {
	m.sagas[s.ID] = s
	return nil
}

func (m *mockStore) GetByID(_ context.Context, id string) (*domain.Saga, error) {
	s, ok := m.sagas[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	// Return a copy to avoid mutation issues between calls
	cp := *s
	steps := make([]domain.Step, len(s.Steps))
	copy(steps, s.Steps)
	cp.Steps = steps
	return &cp, nil
}

func (m *mockStore) GetByOrderID(_ context.Context, orderID string) (*domain.Saga, error) {
	for _, s := range m.sagas {
		if s.OrderID == orderID {
			return s, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (m *mockStore) UpdateState(_ context.Context, id string, state domain.SagaState, steps []domain.Step, errMsg string, completedAt, failedAt *time.Time) error {
	s, ok := m.sagas[id]
	if !ok {
		return domain.ErrNotFound
	}
	s.State = state
	s.Steps = steps
	s.ErrorMsg = errMsg
	s.CompletedAt = completedAt
	s.FailedAt = failedAt
	return nil
}

func (m *mockStore) ListByState(_ context.Context, state domain.SagaState, _ int) ([]*domain.Saga, error) {
	var out []*domain.Saga
	for _, s := range m.sagas {
		if s.State == state {
			out = append(out, s)
		}
	}
	return out, nil
}

// --- mock publisher ---

type mockPublisher struct {
	published []publishedMsg
}

type publishedMsg struct {
	topic, key string
	payload    any
}

func (m *mockPublisher) Publish(_ context.Context, topic, key string, payload any) error {
	m.published = append(m.published, publishedMsg{topic, key, payload})
	return nil
}

func newOrch(st *mockStore, pub *mockPublisher) *service.Orchestrator {
	return service.New(st, pub, service.Topics{
		ReserveInventory: "inventory.reserve",
		ProcessPayment:   "payment.process",
		CreateShipment:   "shipment.create",
		OrderCancelled:   "order.cancelled",
	}, zap.NewNop())
}

// --- tests ---

func TestStartSaga(t *testing.T) {
	st := newMockStore()
	pub := &mockPublisher{}
	orch := newOrch(st, pub)
	ctx := context.Background()

	saga, err := orch.Start(ctx, domain.StartSagaRequest{
		Type:    domain.TypeOrderFulfillment,
		OrderID: "order-1",
		Payload: map[string]string{"user_id": "u1"},
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	if saga.ID == "" {
		t.Error("expected non-empty saga ID")
	}
	if saga.State != domain.StateInventoryPending {
		t.Errorf("expected inventory_pending, got %s", saga.State)
	}
	if len(pub.published) == 0 {
		t.Error("expected inventory reserve message to be published")
	}
	if pub.published[0].topic != "inventory.reserve" {
		t.Errorf("expected inventory.reserve, got %q", pub.published[0].topic)
	}
}

func TestStartSagaMissingOrderID(t *testing.T) {
	orch := newOrch(newMockStore(), &mockPublisher{})
	_, err := orch.Start(context.Background(), domain.StartSagaRequest{Type: domain.TypeOrderFulfillment})
	if err != domain.ErrInvalidInput {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
}

func TestFullHappyPath(t *testing.T) {
	st := newMockStore()
	pub := &mockPublisher{}
	orch := newOrch(st, pub)
	ctx := context.Background()

	saga, _ := orch.Start(ctx, domain.StartSagaRequest{
		Type: domain.TypeOrderFulfillment, OrderID: "order-2",
	})

	// Step 1: inventory success
	if err := orch.OnInventoryResult(ctx, saga.ID, true, ""); err != nil {
		t.Fatalf("OnInventoryResult: %v", err)
	}
	got, _ := orch.GetSaga(ctx, saga.ID)
	if got.State != domain.StatePaymentPending {
		t.Errorf("after inventory: expected payment_pending, got %s", got.State)
	}

	// Step 2: payment success
	if err := orch.OnPaymentResult(ctx, saga.ID, true, ""); err != nil {
		t.Fatalf("OnPaymentResult: %v", err)
	}
	got, _ = orch.GetSaga(ctx, saga.ID)
	if got.State != domain.StateShipmentPending {
		t.Errorf("after payment: expected shipment_pending, got %s", got.State)
	}

	// Step 3: shipment success
	if err := orch.OnShipmentResult(ctx, saga.ID, true, ""); err != nil {
		t.Fatalf("OnShipmentResult: %v", err)
	}
	got, _ = orch.GetSaga(ctx, saga.ID)
	if got.State != domain.StateCompleted {
		t.Errorf("after shipment: expected completed, got %s", got.State)
	}
	if got.CompletedAt == nil {
		t.Error("expected CompletedAt to be set")
	}
}

func TestInventoryFailureTriggersCompensation(t *testing.T) {
	st := newMockStore()
	pub := &mockPublisher{}
	orch := newOrch(st, pub)
	ctx := context.Background()

	saga, _ := orch.Start(ctx, domain.StartSagaRequest{
		Type: domain.TypeOrderFulfillment, OrderID: "order-3",
	})

	if err := orch.OnInventoryResult(ctx, saga.ID, false, "out of stock"); err != nil {
		t.Fatalf("OnInventoryResult: %v", err)
	}

	got := st.sagas[saga.ID]
	if got.State != domain.StateCompensated {
		t.Errorf("expected compensated, got %s", got.State)
	}
	// Check cancellation was published
	var found bool
	for _, m := range pub.published {
		if m.topic == "order.cancelled" {
			found = true
		}
	}
	if !found {
		t.Error("expected order.cancelled to be published")
	}
}

func TestPaymentFailureTriggersCompensation(t *testing.T) {
	st := newMockStore()
	pub := &mockPublisher{}
	orch := newOrch(st, pub)
	ctx := context.Background()

	saga, _ := orch.Start(ctx, domain.StartSagaRequest{
		Type: domain.TypeOrderFulfillment, OrderID: "order-4",
	})
	orch.OnInventoryResult(ctx, saga.ID, true, "")
	orch.OnPaymentResult(ctx, saga.ID, false, "card declined")

	got := st.sagas[saga.ID]
	if got.State != domain.StateCompensated {
		t.Errorf("expected compensated after payment failure, got %s", got.State)
	}
}

func TestGetSagaByOrder(t *testing.T) {
	st := newMockStore()
	orch := newOrch(st, &mockPublisher{})
	ctx := context.Background()

	saga, _ := orch.Start(ctx, domain.StartSagaRequest{
		Type: domain.TypeOrderFulfillment, OrderID: "order-5",
	})

	got, err := orch.GetSagaByOrder(ctx, "order-5")
	if err != nil {
		t.Fatalf("GetSagaByOrder: %v", err)
	}
	if got.ID != saga.ID {
		t.Errorf("expected saga %s, got %s", saga.ID, got.ID)
	}
}

func TestGetSagaNotFound(t *testing.T) {
	orch := newOrch(newMockStore(), &mockPublisher{})
	_, err := orch.GetSaga(context.Background(), "nonexistent")
	if err != domain.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestCanTransition(t *testing.T) {
	cases := []struct {
		from, to domain.SagaState
		ok       bool
	}{
		{domain.StateStarted, domain.StateInventoryPending, true},
		{domain.StateInventoryPending, domain.StateInventoryReserved, true},
		{domain.StateCompleted, domain.StateStarted, false},
		{domain.StateFailed, domain.StateCompleted, false},
		{domain.StatePaymentPending, domain.StateCompensating, true},
	}
	for _, c := range cases {
		got := domain.CanTransition(c.from, c.to)
		if got != c.ok {
			t.Errorf("CanTransition(%s -> %s): expected %v, got %v", c.from, c.to, c.ok, got)
		}
	}
}
