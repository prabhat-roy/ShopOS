package handler

import (
	"encoding/json"
	"net/http"

	"github.com/shopos/address-validation-service/domain"
	"github.com/shopos/address-validation-service/validator"
)

// Handler holds HTTP route logic for the address-validation-service.
type Handler struct{}

// New creates a Handler and registers routes on mux.
func New() *Handler { return &Handler{} }

// RegisterRoutes wires all HTTP routes.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/healthz", h.Healthz)
	mux.HandleFunc("/addresses/validate", h.validateSingle)
	mux.HandleFunc("/addresses/validate/batch", h.validateBatch)
}

// Healthz returns service health.
func (h *Handler) Healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// validateSingle handles POST /addresses/validate.
func (h *Handler) validateSingle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var addr domain.Address
	if err := json.NewDecoder(r.Body).Decode(&addr); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	result := validator.Validate(addr)
	writeJSON(w, http.StatusOK, result)
}

// validateBatch handles POST /addresses/validate/batch.
func (h *Handler) validateBatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var addrs []domain.Address
	if err := json.NewDecoder(r.Body).Decode(&addrs); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	results := make([]domain.ValidationResult, 0, len(addrs))
	for _, addr := range addrs {
		results = append(results, validator.Validate(addr))
	}
	writeJSON(w, http.StatusOK, results)
}

// helpers ----------------------------------------------------------------

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}

func writeError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}
