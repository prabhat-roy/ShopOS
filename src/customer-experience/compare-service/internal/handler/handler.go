package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/shopos/compare-service/internal/domain"
	"github.com/shopos/compare-service/internal/service"
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
	h.mux.HandleFunc("/compare/", h.routeCompare)
}

// healthz returns a simple health check response.
func (h *Handler) healthz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// routeCompare dispatches /compare/{customerId}/... routes.
//
// Supported routes:
//
//	GET    /compare/{customerId}                  → getList
//	POST   /compare/{customerId}/items            → addItem
//	DELETE /compare/{customerId}/items/{productId}→ removeItem
//	DELETE /compare/{customerId}                  → clearList
func (h *Handler) routeCompare(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/compare/")
	parts := strings.SplitN(path, "/", 3)

	if len(parts) == 0 || parts[0] == "" {
		writeError(w, http.StatusBadRequest, "customerId is required")
		return
	}

	customerID := parts[0]

	switch {
	// GET /compare/{customerId}
	case len(parts) == 1 && r.Method == http.MethodGet:
		h.getList(w, r, customerID)

	// DELETE /compare/{customerId}
	case len(parts) == 1 && r.Method == http.MethodDelete:
		h.clearList(w, r, customerID)

	// POST /compare/{customerId}/items
	case len(parts) == 2 && parts[1] == "items" && r.Method == http.MethodPost:
		h.addItem(w, r, customerID)

	// DELETE /compare/{customerId}/items/{productId}
	case len(parts) == 3 && parts[1] == "items" && r.Method == http.MethodDelete:
		h.removeItem(w, r, customerID, parts[2])

	default:
		writeError(w, http.StatusNotFound, "not found")
	}
}

// getList handles GET /compare/{customerId}.
func (h *Handler) getList(w http.ResponseWriter, r *http.Request, customerID string) {
	list, err := h.svc.GetCompareList(r.Context(), customerID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, list)
}

// addItem handles POST /compare/{customerId}/items.
func (h *Handler) addItem(w http.ResponseWriter, r *http.Request, customerID string) {
	var item domain.CompareItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	list, err := h.svc.AddItem(r.Context(), customerID, item)
	if err != nil {
		if errors.Is(err, domain.ErrListFull) {
			writeError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, list)
}

// removeItem handles DELETE /compare/{customerId}/items/{productId}.
func (h *Handler) removeItem(w http.ResponseWriter, r *http.Request, customerID, productID string) {
	list, err := h.svc.RemoveItem(r.Context(), customerID, productID)
	if err != nil {
		if errors.Is(err, domain.ErrItemNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, list)
}

// clearList handles DELETE /compare/{customerId}.
func (h *Handler) clearList(w http.ResponseWriter, r *http.Request, customerID string) {
	if err := h.svc.ClearList(r.Context(), customerID); err != nil {
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
