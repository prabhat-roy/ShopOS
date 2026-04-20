package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopos/quote-rfq-service/internal/domain"
	"github.com/shopos/quote-rfq-service/internal/service"
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
	h.mux.HandleFunc("POST /quotes", h.createQuote)
	h.mux.HandleFunc("GET /quotes", h.listQuotes)
	h.mux.HandleFunc("GET /quotes/{id}", h.getQuote)
	h.mux.HandleFunc("POST /quotes/{id}/submit", h.submitQuote)
	h.mux.HandleFunc("POST /quotes/{id}/review", h.reviewQuote)
	h.mux.HandleFunc("POST /quotes/{id}/provide", h.provideQuote)
	h.mux.HandleFunc("POST /quotes/{id}/accept", h.acceptQuote)
	h.mux.HandleFunc("POST /quotes/{id}/reject", h.rejectQuote)
	h.mux.HandleFunc("DELETE /quotes/{id}", h.cancelQuote)
}

// healthz returns a simple liveness response.
func (h *Handler) healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// createQuote handles POST /quotes — creates a new RFQ in DRAFT status.
func (h *Handler) createQuote(w http.ResponseWriter, r *http.Request) {
	var req service.CreateRFQRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %s", err))
		return
	}
	q, err := h.svc.CreateRFQ(req)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, q)
}

// listQuotes handles GET /quotes with optional ?org_id= and ?status= filters.
func (h *Handler) listQuotes(w http.ResponseWriter, r *http.Request) {
	var orgID *uuid.UUID
	var status *domain.QuoteStatus

	if raw := r.URL.Query().Get("org_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid org_id")
			return
		}
		orgID = &id
	}
	if raw := r.URL.Query().Get("status"); raw != "" {
		s := domain.QuoteStatus(strings.ToUpper(raw))
		status = &s
	}

	quotes, err := h.svc.ListQuotes(orgID, status)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if quotes == nil {
		quotes = []*domain.Quote{}
	}
	writeJSON(w, http.StatusOK, quotes)
}

// getQuote handles GET /quotes/{id}.
func (h *Handler) getQuote(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r, "id")
	if !ok {
		return
	}
	q, err := h.svc.GetQuote(id)
	if err != nil {
		writeQuoteError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, q)
}

// submitQuote handles POST /quotes/{id}/submit.
func (h *Handler) submitQuote(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r, "id")
	if !ok {
		return
	}
	if err := h.svc.SubmitRFQ(id); err != nil {
		writeQuoteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// reviewQuote handles POST /quotes/{id}/review.
func (h *Handler) reviewQuote(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r, "id")
	if !ok {
		return
	}
	if err := h.svc.ReviewQuote(id); err != nil {
		writeQuoteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// provideQuote handles POST /quotes/{id}/provide — vendor sets prices.
func (h *Handler) provideQuote(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r, "id")
	if !ok {
		return
	}
	var req service.ProvideQuoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %s", err))
		return
	}
	q, err := h.svc.ProvideQuote(id, req)
	if err != nil {
		writeQuoteError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, q)
}

// acceptQuote handles POST /quotes/{id}/accept.
func (h *Handler) acceptQuote(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r, "id")
	if !ok {
		return
	}
	if err := h.svc.AcceptQuote(id); err != nil {
		writeQuoteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// rejectQuote handles POST /quotes/{id}/reject.
func (h *Handler) rejectQuote(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r, "id")
	if !ok {
		return
	}
	if err := h.svc.RejectQuote(id); err != nil {
		writeQuoteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// cancelQuote handles DELETE /quotes/{id}.
func (h *Handler) cancelQuote(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r, "id")
	if !ok {
		return
	}
	if err := h.svc.CancelQuote(id); err != nil {
		writeQuoteError(w, err)
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

func writeQuoteError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, domain.ErrInvalidTransition):
		writeError(w, http.StatusConflict, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, err.Error())
	}
}

// provideQuoteRequestWithTime is only used in test helpers; keep time.Time accessible.
var _ = time.Now
