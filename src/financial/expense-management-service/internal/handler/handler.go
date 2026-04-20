package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/shopos/expense-management-service/internal/domain"
	"github.com/shopos/expense-management-service/internal/service"
)

// Handler wires HTTP routes to the service layer.
type Handler struct {
	svc service.Servicer
	mux *http.ServeMux
}

// New creates and registers all routes.
func New(svc service.Servicer) *Handler {
	h := &Handler{svc: svc, mux: http.NewServeMux()}
	h.mux.HandleFunc("/healthz", h.healthz)
	h.mux.HandleFunc("/expenses", h.expensesCollection)
	h.mux.HandleFunc("/expenses/", h.expenseRouter)
	return h
}

// ServeHTTP implements http.Handler.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

// GET /healthz
func (h *Handler) healthz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// /expenses  — POST (create) or GET (list)
func (h *Handler) expensesCollection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.createExpense(w, r)
	case http.MethodGet:
		h.listExpenses(w, r)
	default:
		methodNotAllowed(w)
	}
}

// POST /expenses — creates a new expense (status DRAFT), returns 201.
func (h *Handler) createExpense(w http.ResponseWriter, r *http.Request) {
	var e domain.Expense
	if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}
	created, err := h.svc.CreateExpense(&e)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

// GET /expenses?employeeId=&status=&category=
func (h *Handler) listExpenses(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := domain.ListFilter{
		EmployeeID: q.Get("employeeId"),
		Status:     q.Get("status"),
		Category:   q.Get("category"),
	}
	expenses, err := h.svc.ListExpenses(f)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if expenses == nil {
		expenses = []*domain.Expense{}
	}
	writeJSON(w, http.StatusOK, expenses)
}

// /expenses/{id}[/action] — route based on path suffix and method.
func (h *Handler) expenseRouter(w http.ResponseWriter, r *http.Request) {
	// Strip leading "/expenses/"
	rest := strings.TrimPrefix(r.URL.Path, "/expenses/")
	if rest == "" {
		writeError(w, http.StatusBadRequest, "expense id is required")
		return
	}

	// Split into id + optional action: "{id}" or "{id}/submit" etc.
	parts := strings.SplitN(rest, "/", 2)
	rawID := parts[0]
	action := ""
	if len(parts) == 2 {
		action = parts[1]
	}

	id, err := uuid.Parse(rawID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid expense id: must be a UUID")
		return
	}

	switch action {
	case "":
		h.expenseItem(w, r, id)
	case "submit":
		h.submitExpense(w, r, id)
	case "approve":
		h.approveExpense(w, r, id)
	case "reject":
		h.rejectExpense(w, r, id)
	case "reimburse":
		h.reimburseExpense(w, r, id)
	default:
		writeError(w, http.StatusNotFound, "unknown action: "+action)
	}
}

// GET /expenses/{id}  or  DELETE /expenses/{id}
func (h *Handler) expenseItem(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	switch r.Method {
	case http.MethodGet:
		e, err := h.svc.GetExpense(id)
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, "expense not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, e)

	case http.MethodDelete:
		err := h.svc.DeleteExpense(id)
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, "expense not found")
			return
		}
		if errors.Is(err, domain.ErrInvalidTransition) {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.WriteHeader(http.StatusNoContent)

	default:
		methodNotAllowed(w)
	}
}

// POST /expenses/{id}/submit — DRAFT → SUBMITTED, 204.
func (h *Handler) submitExpense(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	_, err := h.svc.SubmitExpense(id)
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, "expense not found")
		return
	}
	if errors.Is(err, domain.ErrInvalidTransition) {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// POST /expenses/{id}/approve — SUBMITTED → APPROVED, body: {approverId}, 204.
func (h *Handler) approveExpense(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	var req struct {
		ApproverID string `json:"approverId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}
	approverID, err := uuid.Parse(req.ApproverID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "approverId must be a valid UUID")
		return
	}
	_, err = h.svc.ApproveExpense(id, approverID)
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, "expense not found")
		return
	}
	if errors.Is(err, domain.ErrInvalidTransition) {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// POST /expenses/{id}/reject — SUBMITTED → REJECTED, body: {reason}, 204.
func (h *Handler) rejectExpense(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	var req struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}
	_, err := h.svc.RejectExpense(id, req.Reason)
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, "expense not found")
		return
	}
	if errors.Is(err, domain.ErrInvalidTransition) {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// POST /expenses/{id}/reimburse — APPROVED → REIMBURSED, 204.
func (h *Handler) reimburseExpense(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	_, err := h.svc.ReimburseExpense(id)
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, "expense not found")
		return
	}
	if errors.Is(err, domain.ErrInvalidTransition) {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ---- helpers ----

func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}

func methodNotAllowed(w http.ResponseWriter) {
	writeError(w, http.StatusMethodNotAllowed, "method not allowed")
}
