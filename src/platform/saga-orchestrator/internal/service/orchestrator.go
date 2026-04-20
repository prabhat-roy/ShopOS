package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopos/saga-orchestrator/internal/domain"
	"go.uber.org/zap"
)

// Storer is satisfied by store.SagaStore.
type Storer interface {
	Create(ctx context.Context, saga *domain.Saga) error
	GetByID(ctx context.Context, id string) (*domain.Saga, error)
	GetByOrderID(ctx context.Context, orderID string) (*domain.Saga, error)
	UpdateState(ctx context.Context, id string, state domain.SagaState, steps []domain.Step, errMsg string, completedAt, failedAt *time.Time) error
	ListByState(ctx context.Context, state domain.SagaState, limit int) ([]*domain.Saga, error)
}

// Publisher sends commands to downstream services via Kafka.
type Publisher interface {
	Publish(ctx context.Context, topic, key string, payload any) error
}

// Orchestrator drives the order-fulfillment saga state machine.
type Orchestrator struct {
	store  Storer
	pub    Publisher
	topics Topics
	log    *zap.Logger
}

// Topics holds all topic names the orchestrator produces to.
type Topics struct {
	ReserveInventory string
	ProcessPayment   string
	CreateShipment   string
	OrderCancelled   string
}

func New(st Storer, pub Publisher, topics Topics, log *zap.Logger) *Orchestrator {
	return &Orchestrator{store: st, pub: pub, topics: topics, log: log}
}

// Start creates a new saga instance and kicks off the first step.
func (o *Orchestrator) Start(ctx context.Context, req domain.StartSagaRequest) (*domain.Saga, error) {
	if req.OrderID == "" || req.Type == "" {
		return nil, domain.ErrInvalidInput
	}

	saga := &domain.Saga{
		ID:      uuid.New().String(),
		Type:    req.Type,
		OrderID: req.OrderID,
		State:   domain.StateStarted,
		Payload: req.Payload,
		Steps: []domain.Step{
			{Name: "reserve_inventory", State: domain.StepPending},
			{Name: "process_payment", State: domain.StepPending},
			{Name: "create_shipment", State: domain.StepPending},
		},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	if err := o.store.Create(ctx, saga); err != nil {
		return nil, err
	}

	// Kick off step 1: reserve inventory
	saga.Steps[0].State = domain.StepInProgress
	saga.Steps[0].StartedAt = time.Now().UTC()
	saga.State = domain.StateInventoryPending
	_ = o.store.UpdateState(ctx, saga.ID, saga.State, saga.Steps, "", nil, nil)

	_ = o.pub.Publish(ctx, o.topics.ReserveInventory, saga.OrderID, map[string]any{
		"saga_id":  saga.ID,
		"order_id": saga.OrderID,
		"payload":  saga.Payload,
	})

	o.log.Info("saga started", zap.String("saga_id", saga.ID), zap.String("order_id", saga.OrderID))
	return saga, nil
}

// OnInventoryResult handles the reply from the inventory service.
func (o *Orchestrator) OnInventoryResult(ctx context.Context, sagaID string, success bool, errMsg string) error {
	saga, err := o.store.GetByID(ctx, sagaID)
	if err != nil {
		return err
	}
	if !domain.CanTransition(saga.State, domain.StateInventoryReserved) && success {
		return domain.ErrInvalidState
	}

	stepIdx := o.findStep(saga.Steps, "reserve_inventory")
	now := time.Now().UTC()

	if !success {
		saga.Steps[stepIdx].State = domain.StepFailed
		saga.Steps[stepIdx].Error = errMsg
		saga.Steps[stepIdx].CompletedAt = now
		return o.compensate(ctx, saga, errMsg)
	}

	saga.Steps[stepIdx].State = domain.StepSucceeded
	saga.Steps[stepIdx].CompletedAt = now

	// Advance to step 2: process payment
	payIdx := o.findStep(saga.Steps, "process_payment")
	saga.Steps[payIdx].State = domain.StepInProgress
	saga.Steps[payIdx].StartedAt = now
	saga.State = domain.StateInventoryReserved

	_ = o.store.UpdateState(ctx, saga.ID, saga.State, saga.Steps, "", nil, nil)

	saga.State = domain.StatePaymentPending
	_ = o.store.UpdateState(ctx, saga.ID, saga.State, saga.Steps, "", nil, nil)

	_ = o.pub.Publish(ctx, o.topics.ProcessPayment, saga.OrderID, map[string]any{
		"saga_id":  saga.ID,
		"order_id": saga.OrderID,
		"payload":  saga.Payload,
	})
	return nil
}

// OnPaymentResult handles the reply from the payment service.
func (o *Orchestrator) OnPaymentResult(ctx context.Context, sagaID string, success bool, errMsg string) error {
	saga, err := o.store.GetByID(ctx, sagaID)
	if err != nil {
		return err
	}

	stepIdx := o.findStep(saga.Steps, "process_payment")
	now := time.Now().UTC()

	if !success {
		saga.Steps[stepIdx].State = domain.StepFailed
		saga.Steps[stepIdx].Error = errMsg
		saga.Steps[stepIdx].CompletedAt = now
		return o.compensate(ctx, saga, errMsg)
	}

	saga.Steps[stepIdx].State = domain.StepSucceeded
	saga.Steps[stepIdx].CompletedAt = now

	// Advance to step 3: create shipment
	shipIdx := o.findStep(saga.Steps, "create_shipment")
	saga.Steps[shipIdx].State = domain.StepInProgress
	saga.Steps[shipIdx].StartedAt = now
	saga.State = domain.StatePaymentProcessed

	_ = o.store.UpdateState(ctx, saga.ID, saga.State, saga.Steps, "", nil, nil)

	saga.State = domain.StateShipmentPending
	_ = o.store.UpdateState(ctx, saga.ID, saga.State, saga.Steps, "", nil, nil)

	_ = o.pub.Publish(ctx, o.topics.CreateShipment, saga.OrderID, map[string]any{
		"saga_id":  saga.ID,
		"order_id": saga.OrderID,
		"payload":  saga.Payload,
	})
	return nil
}

// OnShipmentResult handles the reply from the shipping service.
func (o *Orchestrator) OnShipmentResult(ctx context.Context, sagaID string, success bool, errMsg string) error {
	saga, err := o.store.GetByID(ctx, sagaID)
	if err != nil {
		return err
	}

	stepIdx := o.findStep(saga.Steps, "create_shipment")
	now := time.Now().UTC()

	if !success {
		saga.Steps[stepIdx].State = domain.StepFailed
		saga.Steps[stepIdx].Error = errMsg
		saga.Steps[stepIdx].CompletedAt = now
		return o.compensate(ctx, saga, errMsg)
	}

	saga.Steps[stepIdx].State = domain.StepSucceeded
	saga.Steps[stepIdx].CompletedAt = now

	completedAt := now
	saga.State = domain.StateCompleted
	_ = o.store.UpdateState(ctx, saga.ID, saga.State, saga.Steps, "", &completedAt, nil)

	o.log.Info("saga completed", zap.String("saga_id", saga.ID))
	return nil
}

// GetSaga retrieves a saga by ID.
func (o *Orchestrator) GetSaga(ctx context.Context, id string) (*domain.Saga, error) {
	return o.store.GetByID(ctx, id)
}

// GetSagaByOrder retrieves the latest saga for an order.
func (o *Orchestrator) GetSagaByOrder(ctx context.Context, orderID string) (*domain.Saga, error) {
	return o.store.GetByOrderID(ctx, orderID)
}

// compensate triggers rollback of all completed steps.
func (o *Orchestrator) compensate(ctx context.Context, saga *domain.Saga, reason string) error {
	now := time.Now().UTC()
	saga.State = domain.StateCompensating
	_ = o.store.UpdateState(ctx, saga.ID, saga.State, saga.Steps, reason, nil, nil)

	// Publish cancellation event — downstream services handle their own rollback
	_ = o.pub.Publish(ctx, o.topics.OrderCancelled, saga.OrderID, map[string]any{
		"saga_id":  saga.ID,
		"order_id": saga.OrderID,
		"reason":   reason,
	})

	saga.State = domain.StateCompensated
	_ = o.store.UpdateState(ctx, saga.ID, saga.State, saga.Steps, reason, nil, &now)

	o.log.Warn("saga compensated", zap.String("saga_id", saga.ID), zap.String("reason", reason))
	return nil
}

func (o *Orchestrator) findStep(steps []domain.Step, name string) int {
	for i, s := range steps {
		if s.Name == name {
			return i
		}
	}
	return -1
}
