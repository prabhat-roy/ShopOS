package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/shopos/wishlist-service/internal/domain"
	"github.com/shopos/wishlist-service/internal/service"
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
	// All wishlist routes are handled by a single pattern-matched handler.
	h.mux.HandleFunc("/wishlist/", h.routeWishlist)
}

// healthz returns a simple health check response.
func (h *Handler) healthz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// routeWishlist dispatches /wishlist/{customerId}/... routes.
//
// Supported routes:
//
//	POST   /wishlist/{customerId}/items              → addItem
//	GET    /wishlist/{customerId}/items              → listItems
//	GET    /wishlist/{customerId}/items/{productId}  → getItem
//	DELETE /wishlist/{customerId}/items/{productId}  → removeItem
//	DELETE /wishlist/{customerId}                    → clearWishlist
//	GET    /wishlist/{customerId}/check?productId=   → checkWishlist
func (h *Handler) routeWishlist(w http.ResponseWriter, r *http.Request) {
	// Strip leading /wishlist/
	path := strings.TrimPrefix(r.URL.Path, "/wishlist/")
	// path is now: {customerId}, {customerId}/items, {customerId}/items/{productId}, etc.

	parts := strings.SplitN(path, "/", 3)
	if len(parts) == 0 || parts[0] == "" {
		writeError(w, http.StatusBadRequest, "customerId is required")
		return
	}

	customerID, err := uuid.Parse(parts[0])
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid customerId: must be a UUID")
		return
	}

	switch {
	// DELETE /wishlist/{customerId}
	case len(parts) == 1 && r.Method == http.MethodDelete:
		h.clearWishlist(w, r, customerID)

	// GET /wishlist/{customerId}/check
	case len(parts) == 2 && parts[1] == "check" && r.Method == http.MethodGet:
		h.checkWishlist(w, r, customerID)

	// POST /wishlist/{customerId}/items
	case len(parts) == 2 && parts[1] == "items" && r.Method == http.MethodPost:
		h.addItem(w, r, customerID)

	// GET /wishlist/{customerId}/items
	case len(parts) == 2 && parts[1] == "items" && r.Method == http.MethodGet:
		h.listItems(w, r, customerID)

	// GET /wishlist/{customerId}/items/{productId}
	case len(parts) == 3 && parts[1] == "items" && r.Method == http.MethodGet:
		h.getItem(w, r, customerID, parts[2])

	// DELETE /wishlist/{customerId}/items/{productId}
	case len(parts) == 3 && parts[1] == "items" && r.Method == http.MethodDelete:
		h.removeItem(w, r, customerID, parts[2])

	default:
		writeError(w, http.StatusNotFound, "not found")
	}
}

// addItem handles POST /wishlist/{customerId}/items.
func (h *Handler) addItem(w http.ResponseWriter, r *http.Request, customerID uuid.UUID) {
	var req domain.AddItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	item, err := h.svc.AddToWishlist(r.Context(), customerID, &req)
	if err != nil {
		if errors.Is(err, domain.ErrAlreadyExists) {
			writeError(w, http.StatusConflict, "item already exists in wishlist")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

// listItems handles GET /wishlist/{customerId}/items.
func (h *Handler) listItems(w http.ResponseWriter, r *http.Request, customerID uuid.UUID) {
	limit := queryInt(r, "limit", 20)
	offset := queryInt(r, "offset", 0)

	page, err := h.svc.GetWishlist(r.Context(), customerID, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, page)
}

// getItem handles GET /wishlist/{customerId}/items/{productId}.
func (h *Handler) getItem(w http.ResponseWriter, r *http.Request, customerID uuid.UUID, productID string) {
	item, err := h.svc.GetWishlistItem(r.Context(), customerID, productID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, "item not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, item)
}

// removeItem handles DELETE /wishlist/{customerId}/items/{productId}.
func (h *Handler) removeItem(w http.ResponseWriter, r *http.Request, customerID uuid.UUID, productID string) {
	if err := h.svc.RemoveFromWishlist(r.Context(), customerID, productID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, "item not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// clearWishlist handles DELETE /wishlist/{customerId}.
func (h *Handler) clearWishlist(w http.ResponseWriter, r *http.Request, customerID uuid.UUID) {
	if err := h.svc.ClearWishlist(r.Context(), customerID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// checkWishlist handles GET /wishlist/{customerId}/check?productId=.
func (h *Handler) checkWishlist(w http.ResponseWriter, r *http.Request, customerID uuid.UUID) {
	productID := r.URL.Query().Get("productId")
	if productID == "" {
		writeError(w, http.StatusBadRequest, "productId query parameter is required")
		return
	}

	inWishlist, err := h.svc.CheckWishlist(r.Context(), customerID, productID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"inWishlist": inWishlist})
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

// queryInt parses an integer query parameter, returning def if absent or invalid.
func queryInt(r *http.Request, key string, def int) int {
	s := r.URL.Query().Get(key)
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return v
}
