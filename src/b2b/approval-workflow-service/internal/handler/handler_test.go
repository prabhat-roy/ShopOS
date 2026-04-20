package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopos/approval-workflow-service/internal/domain"
	"github.com/shopos/approval-workflow-service/internal/handler"
	"github.com/shopos/approval-workflow-service/internal/service"
)

// --- mock service ---

type mockService struct {
	createFn       func(req service.CreateWorkflowRequest) (*domain.ApprovalWorkflow, error)
	getFn          func(id uuid.UUID) (*domain.ApprovalWorkflow, error)
	listFn         func(*domain.EntityType, *domain.WorkflowStatus, *uuid.UUID) ([]*domain.ApprovalWorkflow, error)
	getByEntityFn  func(entityID uuid.UUID) (*domain.ApprovalWorkflow, error)
	approveFn      func(id uuid.UUID, approverID uuid.UUID, comment string) error
	rejectFn       func(id uuid.UUID, approverID uuid.UUID, comment string) error
	cancelFn       func(id uuid.UUID) error
}

func (m *mockService) CreateWorkflow(req service.CreateWorkflowRequest) (*domain.ApprovalWorkflow, error) {
	return m.createFn(req)
}
func (m *mockService) GetWorkflow(id uuid.UUID) (*domain.ApprovalWorkflow, error) { return m.getFn(id) }
func (m *mockService) ListWorkflows(et *domain.EntityType, st *domain.WorkflowStatus, org *uuid.UUID) ([]*domain.ApprovalWorkflow, error) {
	return m.listFn(et, st, org)
}
func (m *mockService) GetByEntityID(entityID uuid.UUID) (*domain.ApprovalWorkflow, error) {
	return m.getByEntityFn(entityID)
}
func (m *mockService) Approve(id uuid.UUID, approverID uuid.UUID, comment string) error {
	return m.approveFn(id, approverID, comment)
}
func (m *mockService) Reject(id uuid.UUID, approverID uuid.UUID, comment string) error {
	return m.rejectFn(id, approverID, comment)
}
func (m *mockService) Cancel(id uuid.UUID) error { return m.cancelFn(id) }

// --- helpers ---

func sampleWorkflow() *domain.ApprovalWorkflow {
	return &domain.ApprovalWorkflow{
		ID:         uuid.New(),
		EntityID:   uuid.New(),
		EntityType: domain.EntityTypePurchaseOrder,
		OrgID:      uuid.New(),
		TotalAmount: 5000,
		Status:     domain.WorkflowStatusInProgress,
		Steps: domain.ApprovalSteps{
			{StepIndex: 0, ApproverRole: "MANAGER", Status: domain.StepStatusPending},
			{StepIndex: 1, ApproverRole: "DIRECTOR", Status: domain.StepStatusPending},
		},
		CurrentStepIndex: 0,
		CreatedBy:        "user-1",
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}
}

func doRequest(h http.Handler, method, path string, body interface{}) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	return rr
}

// --- tests ---

func TestHealthz(t *testing.T) {
	h := handler.New(&mockService{})
	rr := doRequest(h, http.MethodGet, "/healthz", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var resp map[string]string
	_ = json.NewDecoder(rr.Body).Decode(&resp)
	if resp["status"] != "ok" {
		t.Errorf("expected status=ok, got %q", resp["status"])
	}
}

func TestCreateWorkflow_Success(t *testing.T) {
	wf := sampleWorkflow()
	svc := &mockService{
		createFn: func(req service.CreateWorkflowRequest) (*domain.ApprovalWorkflow, error) { return wf, nil },
	}
	h := handler.New(svc)
	body := map[string]interface{}{
		"entity_id":    uuid.New().String(),
		"entity_type":  "purchase_order",
		"org_id":       uuid.New().String(),
		"total_amount": 5000.0,
		"created_by":   "user-1",
	}
	rr := doRequest(h, http.MethodPost, "/workflows", body)
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestCreateWorkflow_ValidationError(t *testing.T) {
	svc := &mockService{
		createFn: func(req service.CreateWorkflowRequest) (*domain.ApprovalWorkflow, error) {
			return nil, domain.ErrNotFound
		},
	}
	h := handler.New(svc)
	rr := doRequest(h, http.MethodPost, "/workflows", map[string]interface{}{})
	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", rr.Code)
	}
}

func TestGetWorkflow_Found(t *testing.T) {
	wf := sampleWorkflow()
	svc := &mockService{
		getFn: func(id uuid.UUID) (*domain.ApprovalWorkflow, error) { return wf, nil },
	}
	h := handler.New(svc)
	rr := doRequest(h, http.MethodGet, "/workflows/"+wf.ID.String(), nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestGetWorkflow_NotFound(t *testing.T) {
	svc := &mockService{
		getFn: func(id uuid.UUID) (*domain.ApprovalWorkflow, error) { return nil, domain.ErrNotFound },
	}
	h := handler.New(svc)
	rr := doRequest(h, http.MethodGet, "/workflows/"+uuid.New().String(), nil)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestListWorkflows_NoFilter(t *testing.T) {
	wf := sampleWorkflow()
	svc := &mockService{
		listFn: func(*domain.EntityType, *domain.WorkflowStatus, *uuid.UUID) ([]*domain.ApprovalWorkflow, error) {
			return []*domain.ApprovalWorkflow{wf}, nil
		},
	}
	h := handler.New(svc)
	rr := doRequest(h, http.MethodGet, "/workflows", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestGetByEntityID_Found(t *testing.T) {
	wf := sampleWorkflow()
	svc := &mockService{
		getByEntityFn: func(entityID uuid.UUID) (*domain.ApprovalWorkflow, error) { return wf, nil },
	}
	h := handler.New(svc)
	rr := doRequest(h, http.MethodGet, "/workflows/entity/"+wf.EntityID.String(), nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestApproveWorkflow_Success(t *testing.T) {
	svc := &mockService{
		approveFn: func(id uuid.UUID, approverID uuid.UUID, comment string) error { return nil },
	}
	h := handler.New(svc)
	body := map[string]interface{}{
		"approver_id": uuid.New().String(),
		"comment":     "Looks good",
	}
	rr := doRequest(h, http.MethodPost, "/workflows/"+uuid.New().String()+"/approve", body)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestRejectWorkflow_Success(t *testing.T) {
	svc := &mockService{
		rejectFn: func(id uuid.UUID, approverID uuid.UUID, comment string) error { return nil },
	}
	h := handler.New(svc)
	body := map[string]interface{}{
		"approver_id": uuid.New().String(),
		"comment":     "Budget exceeded",
	}
	rr := doRequest(h, http.MethodPost, "/workflows/"+uuid.New().String()+"/reject", body)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestCancelWorkflow_Success(t *testing.T) {
	svc := &mockService{
		cancelFn: func(id uuid.UUID) error { return nil },
	}
	h := handler.New(svc)
	rr := doRequest(h, http.MethodDelete, "/workflows/"+uuid.New().String(), nil)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}
}

func TestApproveWorkflow_NotFound(t *testing.T) {
	svc := &mockService{
		approveFn: func(id uuid.UUID, approverID uuid.UUID, comment string) error {
			return domain.ErrNotFound
		},
	}
	h := handler.New(svc)
	body := map[string]interface{}{
		"approver_id": uuid.New().String(),
		"comment":     "ok",
	}
	rr := doRequest(h, http.MethodPost, "/workflows/"+uuid.New().String()+"/approve", body)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}
