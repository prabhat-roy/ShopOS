package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/shopos/tax-provider-integration/internal/domain"
	"github.com/shopos/tax-provider-integration/internal/service"
)

// Handler holds the HTTP multiplexer and a reference to the service layer.
type Handler struct {
	mux *http.ServeMux
	svc service.Servicer
}

// New wires all routes and returns the Handler.
func New(svc service.Servicer) *Handler {
	h := &Handler{mux: http.NewServeMux(), svc: svc}
	h.mux.HandleFunc("/healthz", h.handleHealth)
	h.mux.HandleFunc("/tax/calculate", h.handleCalculate)
	h.mux.HandleFunc("/tax/commit", h.handleCommit)
	h.mux.HandleFunc("/tax/providers", h.handleProviders)
	h.mux.HandleFunc("/tax/providers/", h.handleProviderByID)
	h.mux.HandleFunc("/tax/validate-address", h.handleValidateAddress)
	return h
}

// ServeHTTP satisfies http.Handler.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

// ---------------------------------------------------------------------------
// route handlers
// ---------------------------------------------------------------------------

// GET /healthz
func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// POST /tax/calculate
func (h *Handler) handleCalculate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}

	var req domain.TaxCalculationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	resp, err := h.svc.CalculateTax(req)
	if err != nil {
		code := http.StatusBadRequest
		if strings.Contains(err.Error(), "unsupported provider") {
			code = http.StatusUnprocessableEntity
		}
		writeError(w, code, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// POST /tax/commit
func (h *Handler) handleCommit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}

	var req domain.CommitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	resp, err := h.svc.CommitTransaction(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// GET /tax/providers
func (h *Handler) handleProviders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	providers := h.svc.ListProviders()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"providers": providers,
		"count":     len(providers),
	})
}

// GET /tax/providers/{provider}
func (h *Handler) handleProviderByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}

	providerStr := strings.TrimPrefix(r.URL.Path, "/tax/providers/")
	providerStr = strings.Trim(providerStr, "/")
	if providerStr == "" {
		// Fall through to list handler.
		h.handleProviders(w, r)
		return
	}

	provider := domain.TaxProvider(strings.ToUpper(providerStr))
	info, err := h.svc.GetProviderInfo(provider)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, info)
}

// POST /tax/validate-address
func (h *Handler) handleValidateAddress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}

	var req struct {
		Provider domain.TaxProvider `json:"provider"`
		Address  domain.TaxAddress  `json:"address"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	valid, normalized, err := h.svc.ValidateAddress(req.Provider, req.Address)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"valid":      valid,
		"normalized": normalized,
	})
}

// ---------------------------------------------------------------------------
// JSON helpers
// ---------------------------------------------------------------------------

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
