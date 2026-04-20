package domain

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// WorkflowStatus represents the overall state of an approval workflow.
type WorkflowStatus string

const (
	WorkflowStatusPending    WorkflowStatus = "PENDING"
	WorkflowStatusInProgress WorkflowStatus = "IN_PROGRESS"
	WorkflowStatusApproved   WorkflowStatus = "APPROVED"
	WorkflowStatusRejected   WorkflowStatus = "REJECTED"
	WorkflowStatusCancelled  WorkflowStatus = "CANCELLED"
)

// StepStatus represents the state of an individual approval step.
type StepStatus string

const (
	StepStatusPending  StepStatus = "PENDING"
	StepStatusApproved StepStatus = "APPROVED"
	StepStatusRejected StepStatus = "REJECTED"
	StepStatusSkipped  StepStatus = "SKIPPED"
)

// EntityType enumerates supported entity kinds that can trigger a workflow.
type EntityType string

const (
	EntityTypePurchaseOrder EntityType = "purchase_order"
	EntityTypeQuote         EntityType = "quote"
	EntityTypeContract      EntityType = "contract"
	EntityTypeExpense       EntityType = "expense"
)

// Sentinel errors.
var (
	ErrNotFound      = errors.New("workflow not found")
	ErrNotCurrentStep = errors.New("approver is not the current step's approver role")
)

// ApprovalStep is a single node in the multi-step approval chain.
type ApprovalStep struct {
	StepIndex    int        `json:"step_index"`
	ApproverRole string     `json:"approver_role"`
	ApproverID   *uuid.UUID `json:"approver_id,omitempty"`
	Status       StepStatus `json:"status"`
	Comment      string     `json:"comment,omitempty"`
	DecidedAt    *time.Time `json:"decided_at,omitempty"`
}

// ApprovalSteps is a slice of ApprovalStep with JSONB persistence support.
type ApprovalSteps []ApprovalStep

// Value implements driver.Valuer.
func (as ApprovalSteps) Value() (driver.Value, error) {
	if as == nil {
		return "[]", nil
	}
	b, err := json.Marshal(as)
	if err != nil {
		return nil, fmt.Errorf("marshal ApprovalSteps: %w", err)
	}
	return string(b), nil
}

// Scan implements sql.Scanner.
func (as *ApprovalSteps) Scan(src interface{}) error {
	var data []byte
	switch v := src.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	case nil:
		*as = ApprovalSteps{}
		return nil
	default:
		return fmt.Errorf("unsupported type for ApprovalSteps: %T", src)
	}
	return json.Unmarshal(data, as)
}

// ApprovalWorkflow is the aggregate root for an approval lifecycle.
type ApprovalWorkflow struct {
	ID               uuid.UUID      `json:"id"`
	EntityID         uuid.UUID      `json:"entity_id"`
	EntityType       EntityType     `json:"entity_type"`
	OrgID            uuid.UUID      `json:"org_id"`
	TotalAmount      float64        `json:"total_amount"`
	Status           WorkflowStatus `json:"status"`
	Steps            ApprovalSteps  `json:"steps"`
	CurrentStepIndex int            `json:"current_step_index"`
	CreatedBy        string         `json:"created_by"`
	CompletedAt      *time.Time     `json:"completed_at,omitempty"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
}

// CurrentStep returns the step at CurrentStepIndex, or nil if the workflow is complete.
func (w *ApprovalWorkflow) CurrentStep() *ApprovalStep {
	if w.CurrentStepIndex >= len(w.Steps) {
		return nil
	}
	return &w.Steps[w.CurrentStepIndex]
}

// IsTerminal returns true when the workflow cannot advance further.
func (w *ApprovalWorkflow) IsTerminal() bool {
	switch w.Status {
	case WorkflowStatusApproved, WorkflowStatusRejected, WorkflowStatusCancelled:
		return true
	}
	return false
}
