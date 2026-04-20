package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopos/budget-service/internal/domain"
	"github.com/shopos/budget-service/internal/service"
)

// Handler holds the service dependency and registers HTTP routes.
type Handler struct {
	svc service.Servicer
}

// New returns a Handler wired to svc.
func New(svc service.Servicer) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes registers all endpoints on the provided mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/healthz", h.healthz)
	mux.HandleFunc("/budgets", h.budgets)
	mux.HandleFunc("/budgets/", h.budgetsRouter)
}

// ---- helpers ----------------------------------------------------------------

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("handler: writeJSON: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func parseUUID(s string) (uuid.UUID, error) { return uuid.Parse(s) }

// ---- /healthz ---------------------------------------------------------------

func (h *Handler) healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ---- /budgets ---------------------------------------------------------------

// budgets handles GET /budgets and POST /budgets.
func (h *Handler) budgets(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listBudgets(w, r)
	case http.MethodPost:
		h.createBudget(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

type createBudgetRequest struct {
	Department  string `json:"department"`
	Name        string `json:"name"`
	Period      string `json:"period"`
	FiscalYear  int    `json:"fiscal_year"`
	StartDate   string `json:"start_date"` // RFC3339
	EndDate     string `json:"end_date"`
	TotalAmount float64 `json:"total_amount"`
	Currency    string `json:"currency"`
}

func (h *Handler) createBudget(w http.ResponseWriter, r *http.Request) {
	var req createBudgetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Department == "" || req.Name == "" {
		writeError(w, http.StatusBadRequest, "department and name are required")
		return
	}
	startDate, err := time.Parse(time.RFC3339, req.StartDate)
	if err != nil {
		writeError(w, http.StatusBadRequest, "start_date must be RFC3339")
		return
	}
	endDate, err := time.Parse(time.RFC3339, req.EndDate)
	if err != nil {
		writeError(w, http.StatusBadRequest, "end_date must be RFC3339")
		return
	}
	period := domain.BudgetPeriod(req.Period)
	b, err := h.svc.CreateBudget(req.Department, req.Name, period, req.FiscalYear, startDate, endDate, req.TotalAmount, req.Currency)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, b)
}

func (h *Handler) listBudgets(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	department := q.Get("department")
	status := domain.BudgetStatus(q.Get("status"))
	fiscalYear := 0
	if fy := q.Get("fiscal_year"); fy != "" {
		if v, err := strconv.Atoi(fy); err == nil {
			fiscalYear = v
		}
	}
	budgets, err := h.svc.ListBudgets(department, status, fiscalYear)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"budgets": budgets})
}

// ---- /budgets/{id}[/...] router --------------------------------------------

func (h *Handler) budgetsRouter(w http.ResponseWriter, r *http.Request) {
	// parts: ["budgets", id, sub-resource]
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 2 || parts[1] == "" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	budgetID, err := parseUUID(parts[1])
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid budget id")
		return
	}

	// GET /budgets/{id}
	if len(parts) == 2 {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		h.getBudget(w, r, budgetID)
		return
	}

	switch parts[2] {
	case "activate":
		if r.Method != http.MethodPatch {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		h.activateBudget(w, r, budgetID)
	case "close":
		if r.Method != http.MethodPatch {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		h.closeBudget(w, r, budgetID)
	case "allocations":
		switch r.Method {
		case http.MethodPost:
			h.createAllocation(w, r, budgetID)
		case http.MethodGet:
			h.listAllocations(w, r, budgetID)
		default:
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	case "spending":
		switch r.Method {
		case http.MethodPost:
			h.recordSpending(w, r, budgetID)
		case http.MethodGet:
			h.listSpending(w, r, budgetID)
		default:
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	case "summary":
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		h.getBudgetSummary(w, r, budgetID)
	default:
		writeError(w, http.StatusNotFound, "not found")
	}
}

// GET /budgets/{id}
func (h *Handler) getBudget(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	b, err := h.svc.GetBudget(id)
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, "budget not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, b)
}

// PATCH /budgets/{id}/activate
func (h *Handler) activateBudget(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	err := h.svc.ActivateBudget(id)
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, "budget not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// PATCH /budgets/{id}/close
func (h *Handler) closeBudget(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	err := h.svc.CloseBudget(id)
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, "budget not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// POST /budgets/{id}/allocations
type createAllocationRequest struct {
	Category string  `json:"category"`
	Amount   float64 `json:"amount"`
	Notes    string  `json:"notes"`
}

func (h *Handler) createAllocation(w http.ResponseWriter, r *http.Request, budgetID uuid.UUID) {
	var req createAllocationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Category == "" {
		writeError(w, http.StatusBadRequest, "category is required")
		return
	}
	alloc, err := h.svc.CreateAllocation(budgetID, req.Category, req.Amount, req.Notes)
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, "budget not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, alloc)
}

// GET /budgets/{id}/allocations
func (h *Handler) listAllocations(w http.ResponseWriter, r *http.Request, budgetID uuid.UUID) {
	allocs, err := h.svc.ListAllocations(budgetID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"allocations": allocs})
}

// POST /budgets/{id}/spending
type recordSpendingRequest struct {
	AllocationID string  `json:"allocation_id,omitempty"`
	Category     string  `json:"category"`
	Description  string  `json:"description"`
	Amount       float64 `json:"amount"`
	Reference    string  `json:"reference"`
}

func (h *Handler) recordSpending(w http.ResponseWriter, r *http.Request, budgetID uuid.UUID) {
	var req recordSpendingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Amount <= 0 {
		writeError(w, http.StatusBadRequest, "amount must be positive")
		return
	}

	var allocID *uuid.UUID
	if req.AllocationID != "" {
		id, err := parseUUID(req.AllocationID)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid allocation_id UUID")
			return
		}
		allocID = &id
	}

	sr, err := h.svc.RecordSpending(budgetID, allocID, req.Category, req.Description, req.Amount, req.Reference)
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, "budget or allocation not found")
		return
	}
	if errors.Is(err, domain.ErrBudgetExceeded) {
		writeError(w, http.StatusUnprocessableEntity, "spending would exceed budget")
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, sr)
}

// GET /budgets/{id}/spending
func (h *Handler) listSpending(w http.ResponseWriter, r *http.Request, budgetID uuid.UUID) {
	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}
	records, err := h.svc.ListSpending(budgetID, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"spending": records})
}

// GET /budgets/{id}/summary
func (h *Handler) getBudgetSummary(w http.ResponseWriter, r *http.Request, budgetID uuid.UUID) {
	summary, err := h.svc.GetBudgetSummary(budgetID)
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, "budget not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, summary)
}
