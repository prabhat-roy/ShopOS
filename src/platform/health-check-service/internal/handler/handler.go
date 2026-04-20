package handler

import (
	"encoding/json"
	"net/http"

	"github.com/shopos/health-check-service/internal/domain"
)

// HealthSource is satisfied by checker.Checker.
type HealthSource interface {
	Overall() domain.OverallHealth
	Get(name string) (domain.TargetHealth, bool)
}

// Handler holds all HTTP handlers.
type Handler struct{ src HealthSource }

func New(src HealthSource) *Handler { return &Handler{src: src} }

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /healthz", h.selfHealth)
	mux.HandleFunc("GET /health", h.overall)
	mux.HandleFunc("GET /health/{name}", h.single)
}

// selfHealth is the service's own /healthz — always 200 if process is alive.
func (h *Handler) selfHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// overall returns the aggregated health of all registered targets.
func (h *Handler) overall(w http.ResponseWriter, r *http.Request) {
	oh := h.src.Overall()
	code := http.StatusOK
	if oh.Status != domain.StatusHealthy {
		code = http.StatusServiceUnavailable
	}
	writeJSON(w, code, oh)
}

// single returns the health of one named target.
func (h *Handler) single(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	th, ok := h.src.Get(name)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "target not found"})
		return
	}
	code := http.StatusOK
	if th.Status != domain.StatusHealthy {
		code = http.StatusServiceUnavailable
	}
	writeJSON(w, code, th)
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}
