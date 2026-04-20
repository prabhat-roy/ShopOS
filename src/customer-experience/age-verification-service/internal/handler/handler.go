package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/shopos/age-verification-service/internal/domain"
	"github.com/shopos/age-verification-service/internal/verifier"
)

// Handler holds the HTTP mux and age verifier.
type Handler struct {
	v   *verifier.AgeVerifier
	mux *http.ServeMux
}

// New creates and wires all routes.
func New(v *verifier.AgeVerifier) *Handler {
	h := &Handler{v: v, mux: http.NewServeMux()}
	h.mux.HandleFunc("/healthz", h.healthz)
	h.mux.HandleFunc("/verify/batch", h.batchVerify)
	h.mux.HandleFunc("/verify", h.singleVerify)
	h.mux.HandleFunc("/min-age", h.minAge)
	return h
}

// ServeHTTP satisfies http.Handler.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

// GET /healthz
func (h *Handler) healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// POST /verify
func (h *Handler) singleVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req domain.VerificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	result, err := h.v.Verify(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// POST /verify/batch
func (h *Handler) batchVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var reqs []domain.VerificationRequest
	if err := json.NewDecoder(r.Body).Decode(&reqs); err != nil || len(reqs) == 0 {
		writeError(w, http.StatusBadRequest, "request body must be a non-empty JSON array of verification requests")
		return
	}
	results := h.v.BatchVerify(reqs)
	writeJSON(w, http.StatusOK, results)
}

// GET /min-age?country=UK&category=alcohol
func (h *Handler) minAge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	country := strings.TrimSpace(r.URL.Query().Get("country"))
	category := strings.TrimSpace(r.URL.Query().Get("category"))
	if country == "" || category == "" {
		writeError(w, http.StatusBadRequest, "query params 'country' and 'category' are required")
		return
	}
	minAge := verifier.MinAgeFor(country, category)
	writeJSON(w, http.StatusOK, map[string]any{
		"country":         country,
		"productCategory": category,
		"minAge":          minAge,
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
