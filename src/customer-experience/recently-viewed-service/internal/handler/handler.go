package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/shopos/recently-viewed-service/internal/domain"
	"github.com/shopos/recently-viewed-service/internal/service"
)

// Handler wires HTTP routes to service logic.
type Handler struct {
	svc service.Servicer
	mux *http.ServeMux
}

// New creates a new Handler and registers all routes.
func New(svc service.Servicer) *Handler {
	h := &Handler{svc: svc, mux: http.NewServeMux()}
	h.registerRoutes()
	return h
}

// ServeHTTP implements http.Handler.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func (h *Handler) registerRoutes() {
	h.mux.HandleFunc("/healthz", h.healthz)
	h.mux.HandleFunc("/recently-viewed/", h.routeRecentlyViewed)
}

// healthz returns a simple health check response.
func (h *Handler) healthz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// routeRecentlyViewed dispatches /recently-viewed/{customerId} routes.
//
// Supported routes:
//
//	POST   /recently-viewed/{customerId}   → recordView
//	GET    /recently-viewed/{customerId}?limit=  → getRecentlyViewed
//	DELETE /recently-viewed/{customerId}   → clearHistory
func (h *Handler) routeRecentlyViewed(w http.ResponseWriter, r *http.Request) {
	customerID := strings.TrimPrefix(r.URL.Path, "/recently-viewed/")
	// Strip any trailing slashes.
	customerID = strings.TrimRight(customerID, "/")

	if customerID == "" {
		writeError(w, http.StatusBadRequest, "customerId is required")
		return
	}

	// Only top-level /{customerId} is supported — no sub-paths.
	if strings.Contains(customerID, "/") {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	switch r.Method {
	case http.MethodPost:
		h.recordView(w, r, customerID)
	case http.MethodGet:
		h.getRecentlyViewed(w, r, customerID)
	case http.MethodDelete:
		h.clearHistory(w, r, customerID)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// recordView handles POST /recently-viewed/{customerId}.
func (h *Handler) recordView(w http.ResponseWriter, r *http.Request, customerID string) {
	var item domain.ViewedItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}
	if item.ProductID == "" {
		writeError(w, http.StatusBadRequest, "productId is required")
		return
	}
	// Set ViewedAt server-side if client did not provide it.
	if item.ViewedAt.IsZero() {
		item.ViewedAt = time.Now().UTC()
	}

	if err := h.svc.RecordView(r.Context(), customerID, item); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"status": "recorded"})
}

// getRecentlyViewed handles GET /recently-viewed/{customerId}?limit=.
func (h *Handler) getRecentlyViewed(w http.ResponseWriter, r *http.Request, customerID string) {
	limit := 10
	if v := r.URL.Query().Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err == nil && n > 0 {
			limit = n
		}
	}

	list, err := h.svc.GetRecentlyViewed(r.Context(), customerID, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, list)
}

// clearHistory handles DELETE /recently-viewed/{customerId}.
func (h *Handler) clearHistory(w http.ResponseWriter, r *http.Request, customerID string) {
	if err := h.svc.ClearHistory(r.Context(), customerID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// writeJSON encodes v as JSON and sends it with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError sends a JSON error response.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

