package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/shopos/approval-workflow-service/internal/domain"
	"github.com/shopos/approval-workflow-service/internal/service"
)

// Handler bundles the HTTP mux and service layer.
type Handler struct {
	mux *http.ServeMux
	svc service.Servicer
}

// New wires up all routes and returns a ready Handler.
func New(svc service.Servicer) *Handler {
	h := &Handler{mux: http.NewServeMux(), svc: svc}
	h.routes()
	return h
}

// ServeHTTP implements http.Handler.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func (h *Handler) routes() {
	h.mux.HandleFunc("GET /healthz", h.healthz)
	h.mux.HandleFunc("POST /workflows", h.createWorkflow)
	h.mux.HandleFunc("GET /workflows", h.listWorkflows)
	h.mux.HandleFunc("GET /workflows/{id}", h.getWorkflow)
	h.mux.HandleFunc("GET /workflows/entity/{entityId}", h.getByEntityID)
	h.mux.HandleFunc("POST /workflows/{id}/approve", h.approveWorkflow)
	h.mux.HandleFunc("POST /workflows/{id}/reject", h.rejectWorkflow)
	h.mux.HandleFunc("DELETE /workflows/{id}", h.cancelWorkflow)
}

func (h *Handler) healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) createWorkflow(w http.ResponseWriter, r *http.Request) {
	var req service.CreateWorkflowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %s", err))
		return
	}
	wf, err := h.svc.CreateWorkflow(req)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, wf)
}

func (h *Handler) listWorkflows(w http.ResponseWriter, r *http.Request) {
	var entityType *domain.EntityType
	var status *domain.WorkflowStatus
	var orgID *uuid.UUID

	if raw := r.URL.Query().Get("entity_type"); raw != "" {
		et := domain.EntityType(strings.ToLower(raw))
		entityType = &et
	}
	if raw := r.URL.Query().Get("status"); raw != "" {
		st := domain.WorkflowStatus(strings.ToUpper(raw))
		status = &st
	}
	if raw := r.URL.Query().Get("org_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid org_id")
			return
		}
		orgID = &id
	}

	workflows, err := h.svc.ListWorkflows(entityType, status, orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if workflows == nil {
		workflows = []*domain.ApprovalWorkflow{}
	}
	writeJSON(w, http.StatusOK, workflows)
}

func (h *Handler) getWorkflow(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r, "id")
	if !ok {
		return
	}
	wf, err := h.svc.GetWorkflow(id)
	if err != nil {
		writeWorkflowError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, wf)
}

func (h *Handler) getByEntityID(w http.ResponseWriter, r *http.Request) {
	entityID, ok := parseUUID(w, r, "entityId")
	if !ok {
		return
	}
	wf, err := h.svc.GetByEntityID(entityID)
	if err != nil {
		writeWorkflowError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, wf)
}

// decisionBody is the expected JSON body for approve/reject endpoints.
type decisionBody struct {
	ApproverID string `json:"approver_id"`
	Comment    string `json:"comment"`
}

func (h *Handler) approveWorkflow(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r, "id")
	if !ok {
		return
	}
	var body decisionBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %s", err))
		return
	}
	approverID, err := uuid.Parse(body.ApproverID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid approver_id")
		return
	}
	if err := h.svc.Approve(id, approverID, body.Comment); err != nil {
		writeWorkflowError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) rejectWorkflow(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r, "id")
	if !ok {
		return
	}
	var body decisionBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %s", err))
		return
	}
	approverID, err := uuid.Parse(body.ApproverID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid approver_id")
		return
	}
	if err := h.svc.Reject(id, approverID, body.Comment); err != nil {
		writeWorkflowError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) cancelWorkflow(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r, "id")
	if !ok {
		return
	}
	if err := h.svc.Cancel(id); err != nil {
		writeWorkflowError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- helpers ---

func parseUUID(w http.ResponseWriter, r *http.Request, key string) (uuid.UUID, bool) {
	raw := r.PathValue(key)
	id, err := uuid.Parse(raw)
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid %s: %s", key, raw))
		return uuid.Nil, false
	}
	return id, true
}

func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

type errorResponse struct {
	Error string `json:"error"`
}

func writeError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, errorResponse{Error: msg})
}

func writeWorkflowError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, domain.ErrNotCurrentStep):
		writeError(w, http.StatusForbidden, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, err.Error())
	}
}
