package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/shopos/geolocation-service/internal/domain"
)

// Looker is the interface the handler depends on for IP resolution.
// Using an interface allows the lookup layer to be replaced or mocked in tests.
type Looker interface {
	Resolve(ip string) (*domain.Location, error)
	ResolveMany(ips []string) ([]*domain.Location, error)
}

// Handler holds the HTTP mux and a reference to the lookup dependency.
type Handler struct {
	mux    *http.ServeMux
	looker Looker
}

// New creates a Handler, wires all routes, and returns it.
func New(looker Looker) *Handler {
	h := &Handler{
		mux:    http.NewServeMux(),
		looker: looker,
	}
	h.routes()
	return h
}

// ServeHTTP implements http.Handler so Handler can be passed directly to
// http.ListenAndServe.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

// routes registers all HTTP endpoints.
func (h *Handler) routes() {
	h.mux.HandleFunc("/healthz", h.handleHealthz)
	h.mux.HandleFunc("/locate/batch", h.handleBatch)
	// Catch-all for /locate/{ip} and /locate?ip=...
	h.mux.HandleFunc("/locate", h.handleLocate)
	h.mux.HandleFunc("/locate/", h.handleLocatePath)
}

// handleHealthz returns a simple liveness response.
func (h *Handler) handleHealthz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handleLocate handles GET /locate?ip={ip}.
func (h *Handler) handleLocate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	ip := r.URL.Query().Get("ip")
	if ip == "" {
		writeError(w, http.StatusBadRequest, "query parameter 'ip' is required")
		return
	}

	h.resolveAndRespond(w, ip)
}

// handleLocatePath handles GET /locate/{ip}.
func (h *Handler) handleLocatePath(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Strip the leading "/locate/" prefix to get the IP segment.
	ip := strings.TrimPrefix(r.URL.Path, "/locate/")
	if ip == "" {
		writeError(w, http.StatusBadRequest, "IP address is required in path")
		return
	}

	h.resolveAndRespond(w, ip)
}

// handleBatch handles POST /locate/batch with a JSON body {"ips":["...","..."]}.
func (h *Handler) handleBatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		IPs []string `json:"ips"`
	}

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body: "+err.Error())
		return
	}

	if len(req.IPs) == 0 {
		writeError(w, http.StatusBadRequest, "'ips' array must not be empty")
		return
	}

	locations, err := h.looker.ResolveMany(req.IPs)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "lookup error: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"results": locations,
		"count":   len(locations),
	})
}

// resolveAndRespond is a shared helper for single-IP resolution handlers.
func (h *Handler) resolveAndRespond(w http.ResponseWriter, ip string) {
	loc, err := h.looker.Resolve(ip)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidIP) {
			writeError(w, http.StatusBadRequest, "invalid IP address: "+ip)
			return
		}
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, "location not found for IP: "+ip)
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, loc)
}

// writeJSON encodes v as JSON and writes it to w with the given status code.
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError writes a JSON error response.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
