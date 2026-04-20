package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/shopos/returns-logistics-service/internal/domain"
	"github.com/shopos/returns-logistics-service/internal/service"
)

// Servicer is the interface the handler depends on, allowing test injection.
type Servicer interface {
	CreateReturnAuth(orderID, customerID, reason string, items []domain.ReturnItem) (*domain.ReturnAuth, error)
	GetReturnAuth(id string) (*domain.ReturnAuth, error)
	ListReturnAuths(customerID string) ([]*domain.ReturnAuth, error)
	ApproveReturn(id string) (*domain.ReturnAuth, error)
	RejectReturn(id, reason string) (*domain.ReturnAuth, error)
	IssueLabel(id string) (*domain.ReturnAuth, error)
	MarkInTransit(id string) (*domain.ReturnAuth, error)
	MarkReceived(id string) (*domain.ReturnAuth, error)
	StartInspection(id string) (*domain.ReturnAuth, error)
	CompleteReturn(id, notes string) (*domain.ReturnAuth, error)
	Cancel(id string) (*domain.ReturnAuth, error)
}

// Handler holds the HTTP handler state.
type Handler struct {
	svc Servicer
	mux *http.ServeMux
}

// New creates a Handler wired to a concrete ReturnService.
func New(svc *service.ReturnService) *Handler {
	return newHandler(svc)
}

// NewWithServicer creates a Handler using the Servicer interface (for tests).
func NewWithServicer(svc Servicer) *Handler {
	return newHandler(svc)
}

func newHandler(svc Servicer) *Handler {
	h := &Handler{svc: svc}
	h.mux = http.NewServeMux()
	h.registerRoutes()
	return h
}

func (h *Handler) registerRoutes() {
	h.mux.HandleFunc("/healthz", h.handleHealthz)
	h.mux.HandleFunc("/returns", h.handleReturns)
	h.mux.HandleFunc("/returns/", h.handleReturnByID)
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

// ─── top-level collection handler ─────────────────────────────────────────────

func (h *Handler) handleHealthz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handleReturns handles:
//
//	POST /returns  — create return auth
//	GET  /returns  — list (optional ?customerId=)
func (h *Handler) handleReturns(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.createReturn(w, r)
	case http.MethodGet:
		h.listReturns(w, r)
	default:
		methodNotAllowed(w)
	}
}

type createReturnRequest struct {
	OrderID    string             `json:"orderId"`
	CustomerID string             `json:"customerId"`
	Reason     string             `json:"reason"`
	Items      []domain.ReturnItem `json:"items"`
}

func (h *Handler) createReturn(w http.ResponseWriter, r *http.Request) {
	var req createReturnRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		badRequest(w, "invalid JSON body: "+err.Error())
		return
	}
	if req.OrderID == "" || req.CustomerID == "" {
		badRequest(w, "orderId and customerId are required")
		return
	}
	if len(req.Items) == 0 {
		badRequest(w, "at least one item is required")
		return
	}

	ra, err := h.svc.CreateReturnAuth(req.OrderID, req.CustomerID, req.Reason, req.Items)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidRequest) {
			badRequest(w, err.Error())
			return
		}
		internalError(w, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, ra)
}

func (h *Handler) listReturns(w http.ResponseWriter, r *http.Request) {
	customerID := r.URL.Query().Get("customerId")
	ras, err := h.svc.ListReturnAuths(customerID)
	if err != nil {
		internalError(w, err.Error())
		return
	}
	if ras == nil {
		ras = []*domain.ReturnAuth{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"returns": ras,
		"total":   len(ras),
	})
}

// ─── per-resource handler ──────────────────────────────────────────────────────

// handleReturnByID dispatches sub-resource actions on /returns/{id}[/action].
func (h *Handler) handleReturnByID(w http.ResponseWriter, r *http.Request) {
	// Strip leading "/returns/"
	path := strings.TrimPrefix(r.URL.Path, "/returns/")
	parts := strings.SplitN(path, "/", 2)
	id := parts[0]
	if id == "" {
		notFound(w, "not found")
		return
	}

	if len(parts) == 1 {
		// /returns/{id}
		switch r.Method {
		case http.MethodGet:
			h.getReturn(w, r, id)
		case http.MethodDelete:
			h.cancelReturn(w, r, id)
		default:
			methodNotAllowed(w)
		}
		return
	}

	action := parts[1]
	switch action {
	case "approve":
		h.approveReturn(w, r, id)
	case "reject":
		h.rejectReturn(w, r, id)
	case "label":
		h.issueLabel(w, r, id)
	case "transit":
		h.markInTransit(w, r, id)
	case "receive":
		h.markReceived(w, r, id)
	case "inspect":
		h.startInspection(w, r, id)
	case "complete":
		h.completeReturn(w, r, id)
	default:
		notFound(w, "action not found")
	}
}

func (h *Handler) getReturn(w http.ResponseWriter, _ *http.Request, id string) {
	ra, err := h.svc.GetReturnAuth(id)
	if err != nil {
		handleServiceErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, ra)
}

func (h *Handler) approveReturn(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	if _, err := h.svc.ApproveReturn(id); err != nil {
		handleServiceErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type rejectRequest struct {
	Reason string `json:"reason"`
}

func (h *Handler) rejectReturn(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	var req rejectRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	if _, err := h.svc.RejectReturn(id, req.Reason); err != nil {
		handleServiceErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) issueLabel(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	ra, err := h.svc.IssueLabel(id)
	if err != nil {
		handleServiceErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"trackingNumber": ra.TrackingNumber,
		"labelUrl":       ra.ReturnLabel,
	})
}

func (h *Handler) markInTransit(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	if _, err := h.svc.MarkInTransit(id); err != nil {
		handleServiceErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) markReceived(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	if _, err := h.svc.MarkReceived(id); err != nil {
		handleServiceErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) startInspection(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	if _, err := h.svc.StartInspection(id); err != nil {
		handleServiceErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type completeRequest struct {
	Notes string `json:"notes"`
}

func (h *Handler) completeReturn(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	var req completeRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	if _, err := h.svc.CompleteReturn(id, req.Notes); err != nil {
		handleServiceErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) cancelReturn(w http.ResponseWriter, r *http.Request, id string) {
	if _, err := h.svc.Cancel(id); err != nil {
		handleServiceErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ─── response helpers ─────────────────────────────────────────────────────────

func handleServiceErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		notFound(w, err.Error())
	case errors.Is(err, domain.ErrInvalidTransition):
		writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
	case errors.Is(err, domain.ErrInvalidRequest):
		badRequest(w, err.Error())
	default:
		internalError(w, err.Error())
	}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func badRequest(w http.ResponseWriter, msg string) {
	writeJSON(w, http.StatusBadRequest, map[string]string{"error": msg})
}

func notFound(w http.ResponseWriter, msg string) {
	writeJSON(w, http.StatusNotFound, map[string]string{"error": msg})
}

func methodNotAllowed(w http.ResponseWriter) {
	writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
}

func internalError(w http.ResponseWriter, msg string) {
	writeJSON(w, http.StatusInternalServerError, map[string]string{"error": msg})
}
