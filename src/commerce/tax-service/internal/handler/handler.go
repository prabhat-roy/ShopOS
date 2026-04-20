package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/shopos/tax-service/internal/calculator"
	"github.com/shopos/tax-service/internal/domain"
)

// Handler holds a reference to the tax calculator and exposes HTTP routes.
type Handler struct {
	calc   *calculator.Calculator
	mux    *http.ServeMux
	logger *log.Logger
}

// New creates a Handler and registers all routes on a new ServeMux.
func New(calc *calculator.Calculator, logger *log.Logger) *Handler {
	h := &Handler{
		calc:   calc,
		mux:    http.NewServeMux(),
		logger: logger,
	}
	h.registerRoutes()
	return h
}

// ServeHTTP satisfies the http.Handler interface.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

// registerRoutes wires URL patterns to handler methods.
func (h *Handler) registerRoutes() {
	h.mux.HandleFunc("/healthz", h.healthz)
	h.mux.HandleFunc("/tax/calculate", h.calculate)
	h.mux.HandleFunc("/tax/rates", h.rates)
}

// ─── /healthz ────────────────────────────────────────────────────────────────

func (h *Handler) healthz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ─── POST /tax/calculate ─────────────────────────────────────────────────────

func (h *Handler) calculate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}

	var req domain.TaxRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	if req.ShipTo.Country == "" {
		writeError(w, http.StatusBadRequest, "ship_to.country is required")
		return
	}
	if len(req.Items) == 0 {
		writeError(w, http.StatusBadRequest, "items must not be empty")
		return
	}

	// Validate items
	for i, item := range req.Items {
		if item.Amount < 0 {
			writeError(w, http.StatusBadRequest, "item amount must be non-negative")
			return
		}
		if item.Quantity <= 0 {
			h.logger.Printf("warning: item[%d] quantity %d defaulted to 1", i, item.Quantity)
			req.Items[i].Quantity = 1
		}
	}

	resp := h.calc.Calculate(req)
	writeJSON(w, http.StatusOK, resp)
}

// ─── GET /tax/rates ──────────────────────────────────────────────────────────

func (h *Handler) rates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}

	country := r.URL.Query().Get("country")
	if country == "" {
		writeError(w, http.StatusBadRequest, "country query parameter is required")
		return
	}
	state := r.URL.Query().Get("state")

	info := h.calc.RateInfo(country, state)
	writeJSON(w, http.StatusOK, info)
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("writeJSON encode error: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, domain.ErrorResponse{Error: msg})
}

func methodNotAllowed(w http.ResponseWriter) {
	writeError(w, http.StatusMethodNotAllowed, "method not allowed")
}
