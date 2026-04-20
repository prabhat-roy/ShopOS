package domain

import (
	"errors"
	"time"
)

var (
	ErrNotFound      = errors.New("saga not found")
	ErrInvalidInput  = errors.New("invalid input")
	ErrInvalidState  = errors.New("invalid saga state transition")
)

// SagaState represents the current lifecycle state of a saga instance.
type SagaState string

const (
	StateStarted             SagaState = "started"
	StateInventoryPending    SagaState = "inventory_pending"
	StateInventoryReserved   SagaState = "inventory_reserved"
	StatePaymentPending      SagaState = "payment_pending"
	StatePaymentProcessed    SagaState = "payment_processed"
	StateShipmentPending     SagaState = "shipment_pending"
	StateCompleted           SagaState = "completed"
	StateCompensating        SagaState = "compensating"
	StateCompensated         SagaState = "compensated"
	StateFailed              SagaState = "failed"
)

// SagaType identifies which business process this saga orchestrates.
type SagaType string

const (
	TypeOrderFulfillment SagaType = "order_fulfillment"
)

// Step is one unit of work within a saga, with its own state and compensation.
type Step struct {
	Name        string    `json:"name"`
	State       StepState `json:"state"`
	StartedAt   time.Time `json:"started_at"`
	CompletedAt time.Time `json:"completed_at,omitempty"`
	Error       string    `json:"error,omitempty"`
}

// StepState tracks an individual step.
type StepState string

const (
	StepPending    StepState = "pending"
	StepInProgress StepState = "in_progress"
	StepSucceeded  StepState = "succeeded"
	StepFailed     StepState = "failed"
	StepSkipped    StepState = "skipped"
)

// Saga is the root aggregate — one instance per business transaction.
type Saga struct {
	ID          string            `json:"id"`
	Type        SagaType          `json:"type"`
	OrderID     string            `json:"order_id"`
	State       SagaState         `json:"state"`
	Steps       []Step            `json:"steps"`
	Payload     map[string]string `json:"payload"` // arbitrary context (user_id, amounts, etc.)
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	CompletedAt *time.Time        `json:"completed_at,omitempty"`
	FailedAt    *time.Time        `json:"failed_at,omitempty"`
	ErrorMsg    string            `json:"error_msg,omitempty"`
}

// ValidTransitions maps allowed state transitions.
var ValidTransitions = map[SagaState][]SagaState{
	StateStarted:           {StateInventoryPending, StateCompensating, StateFailed},
	StateInventoryPending:  {StateInventoryReserved, StateCompensating, StateFailed},
	StateInventoryReserved: {StatePaymentPending, StateCompensating, StateFailed},
	StatePaymentPending:    {StatePaymentProcessed, StateCompensating, StateFailed},
	StatePaymentProcessed:  {StateShipmentPending, StateCompensating, StateFailed},
	StateShipmentPending:   {StateCompleted, StateCompensating, StateFailed},
	StateCompensating:      {StateCompensated, StateFailed},
	StateCompleted:         {},
	StateCompensated:       {},
	StateFailed:            {},
}

// CanTransition checks whether moving from current to next is allowed.
func CanTransition(current, next SagaState) bool {
	for _, allowed := range ValidTransitions[current] {
		if allowed == next {
			return true
		}
	}
	return false
}

// StartSagaRequest is the payload to begin a new saga.
type StartSagaRequest struct {
	Type    SagaType          `json:"type"`
	OrderID string            `json:"order_id"`
	Payload map[string]string `json:"payload"`
}
