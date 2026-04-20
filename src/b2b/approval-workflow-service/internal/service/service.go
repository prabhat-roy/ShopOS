package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopos/approval-workflow-service/internal/domain"
	"github.com/shopos/approval-workflow-service/internal/store"
)

// Servicer exposes all business operations for approval workflows.
type Servicer interface {
	CreateWorkflow(req CreateWorkflowRequest) (*domain.ApprovalWorkflow, error)
	GetWorkflow(id uuid.UUID) (*domain.ApprovalWorkflow, error)
	ListWorkflows(entityType *domain.EntityType, status *domain.WorkflowStatus, orgID *uuid.UUID) ([]*domain.ApprovalWorkflow, error)
	GetByEntityID(entityID uuid.UUID) (*domain.ApprovalWorkflow, error)
	Approve(id uuid.UUID, approverID uuid.UUID, comment string) error
	Reject(id uuid.UUID, approverID uuid.UUID, comment string) error
	Cancel(id uuid.UUID) error
}

// CreateWorkflowRequest carries the data needed to start a new workflow.
type CreateWorkflowRequest struct {
	EntityID    uuid.UUID          `json:"entity_id"`
	EntityType  domain.EntityType  `json:"entity_type"`
	OrgID       uuid.UUID          `json:"org_id"`
	TotalAmount float64            `json:"total_amount"`
	CreatedBy   string             `json:"created_by"`
}

// approvalRoles maps step index (0-based) to the required approver role.
var approvalRoles = []string{"MANAGER", "DIRECTOR", "VP"}

// stepsForAmount determines the approval chain based on total amount.
// < $1,000  → 1 step (MANAGER)
// $1,000–$10,000 → 2 steps (MANAGER + DIRECTOR)
// > $10,000 → 3 steps (MANAGER + DIRECTOR + VP)
func stepsForAmount(amount float64) domain.ApprovalSteps {
	numSteps := 1
	if amount >= 10_000 {
		numSteps = 3
	} else if amount >= 1_000 {
		numSteps = 2
	}

	steps := make(domain.ApprovalSteps, numSteps)
	for i := 0; i < numSteps; i++ {
		steps[i] = domain.ApprovalStep{
			StepIndex:    i,
			ApproverRole: approvalRoles[i],
			Status:       domain.StepStatusPending,
		}
	}
	return steps
}

// Service is the concrete implementation of Servicer.
type Service struct {
	store store.Storer
}

// New constructs a Service backed by the supplied Storer.
func New(s store.Storer) *Service {
	return &Service{store: s}
}

// CreateWorkflow creates a new approval workflow and determines its steps based on amount.
func (svc *Service) CreateWorkflow(req CreateWorkflowRequest) (*domain.ApprovalWorkflow, error) {
	if req.EntityID == uuid.Nil {
		return nil, fmt.Errorf("entity_id is required")
	}
	if req.OrgID == uuid.Nil {
		return nil, fmt.Errorf("org_id is required")
	}
	if req.CreatedBy == "" {
		return nil, fmt.Errorf("created_by is required")
	}
	validTypes := map[domain.EntityType]bool{
		domain.EntityTypePurchaseOrder: true,
		domain.EntityTypeQuote:         true,
		domain.EntityTypeContract:      true,
		domain.EntityTypeExpense:       true,
	}
	if !validTypes[req.EntityType] {
		return nil, fmt.Errorf("invalid entity_type: %s", req.EntityType)
	}

	steps := stepsForAmount(req.TotalAmount)
	now := time.Now().UTC()
	w := &domain.ApprovalWorkflow{
		ID:               uuid.New(),
		EntityID:         req.EntityID,
		EntityType:       req.EntityType,
		OrgID:            req.OrgID,
		TotalAmount:      req.TotalAmount,
		Status:           domain.WorkflowStatusInProgress,
		Steps:            steps,
		CurrentStepIndex: 0,
		CreatedBy:        req.CreatedBy,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := svc.store.Create(w); err != nil {
		return nil, fmt.Errorf("create workflow: %w", err)
	}
	return w, nil
}

// GetWorkflow retrieves a workflow by ID.
func (svc *Service) GetWorkflow(id uuid.UUID) (*domain.ApprovalWorkflow, error) {
	return svc.store.Get(id)
}

// ListWorkflows returns workflows with optional filters.
func (svc *Service) ListWorkflows(entityType *domain.EntityType, status *domain.WorkflowStatus, orgID *uuid.UUID) ([]*domain.ApprovalWorkflow, error) {
	return svc.store.List(entityType, status, orgID)
}

// GetByEntityID retrieves the most recent workflow for a given entity.
func (svc *Service) GetByEntityID(entityID uuid.UUID) (*domain.ApprovalWorkflow, error) {
	return svc.store.GetByEntityID(entityID)
}

// Approve approves the current step. If all steps are approved, marks the workflow APPROVED.
func (svc *Service) Approve(id uuid.UUID, approverID uuid.UUID, comment string) error {
	w, err := svc.store.Get(id)
	if err != nil {
		return err
	}
	if w.IsTerminal() {
		return fmt.Errorf("workflow is already in a terminal state: %s", w.Status)
	}

	step := w.CurrentStep()
	if step == nil {
		return fmt.Errorf("no current step to approve")
	}

	now := time.Now().UTC()
	step.ApproverID = &approverID
	step.Status = domain.StepStatusApproved
	step.Comment = comment
	step.DecidedAt = &now
	w.Steps[w.CurrentStepIndex] = *step
	w.UpdatedAt = now

	// Advance to next step or mark completed.
	nextIndex := w.CurrentStepIndex + 1
	if nextIndex >= len(w.Steps) {
		w.Status = domain.WorkflowStatusApproved
		w.CompletedAt = &now
	} else {
		w.CurrentStepIndex = nextIndex
	}

	return svc.store.UpdateWorkflow(w)
}

// Reject rejects the current step and marks the entire workflow REJECTED.
func (svc *Service) Reject(id uuid.UUID, approverID uuid.UUID, comment string) error {
	w, err := svc.store.Get(id)
	if err != nil {
		return err
	}
	if w.IsTerminal() {
		return fmt.Errorf("workflow is already in a terminal state: %s", w.Status)
	}

	step := w.CurrentStep()
	if step == nil {
		return fmt.Errorf("no current step to reject")
	}

	now := time.Now().UTC()
	step.ApproverID = &approverID
	step.Status = domain.StepStatusRejected
	step.Comment = comment
	step.DecidedAt = &now
	w.Steps[w.CurrentStepIndex] = *step

	w.Status = domain.WorkflowStatusRejected
	w.CompletedAt = &now
	w.UpdatedAt = now

	return svc.store.UpdateWorkflow(w)
}

// Cancel marks a non-terminal workflow as CANCELLED.
func (svc *Service) Cancel(id uuid.UUID) error {
	w, err := svc.store.Get(id)
	if err != nil {
		return err
	}
	if w.IsTerminal() {
		return fmt.Errorf("workflow is already in a terminal state: %s", w.Status)
	}
	now := time.Now().UTC()
	w.Status = domain.WorkflowStatusCancelled
	w.CompletedAt = &now
	w.UpdatedAt = now
	return svc.store.UpdateWorkflow(w)
}
