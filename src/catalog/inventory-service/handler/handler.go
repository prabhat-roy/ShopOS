package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/shopos/inventory-service/domain"
)

// Servicer is the application logic contract consumed by the HTTP layer.
type Servicer interface {
	GetStock(ctx context.Context, productID, warehouseID string) (*domain.StockLevel, error)
	ListStock(ctx context.Context, productID string) ([]*domain.StockLevel, error)
	UpsertStock(ctx context.Context, productID, sku, warehouseID string, available, reorder int) (*domain.StockLevel, error)
	Reserve(ctx context.Context, orderID, productID string, qty int) (*domain.Reservation, error)
	Release(ctx context.Context, reservationID string) error
	Commit(ctx context.Context, reservationID string) error
	GetReservation(ctx context.Context, id string) (*domain.Reservation, error)
}

// Handler holds dependencies for all HTTP handlers.
type Handler struct {
	svc Servicer
}

// New returns a Handler wired to the given Servicer.
func New(svc Servicer) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes attaches all routes to the provided ServeMux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/healthz", h.healthz)
	mux.HandleFunc("/inventory/reserve", h.reserve)
	mux.HandleFunc("/inventory/release/", h.release)
	mux.HandleFunc("/inventory/commit/", h.commit)
	mux.HandleFunc("/inventory/", h.inventoryRouter)
}

// -----------------------------------------------------------------------
// Health
// -----------------------------------------------------------------------

func (h *Handler) healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// -----------------------------------------------------------------------
// Inventory routing  (GET /inventory/{productID} and /inventory/{productID}/{warehouseID})
// -----------------------------------------------------------------------

func (h *Handler) inventoryRouter(w http.ResponseWriter, r *http.Request) {
	// Strip leading "/inventory/"
	path := strings.TrimPrefix(r.URL.Path, "/inventory/")
	parts := strings.SplitN(path, "/", 2)

	switch {
	case len(parts) == 1 && parts[0] != "":
		// /inventory/{productID}
		switch r.Method {
		case http.MethodGet:
			h.listStock(w, r, parts[0])
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	case len(parts) == 2 && parts[0] != "" && parts[1] != "":
		// /inventory/{productID}/{warehouseID}
		switch r.Method {
		case http.MethodGet:
			h.getStock(w, r, parts[0], parts[1])
		case http.MethodPut:
			h.upsertStock(w, r, parts[0], parts[1])
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	default:
		http.NotFound(w, r)
	}
}

// GET /inventory/{productID}
func (h *Handler) listStock(w http.ResponseWriter, r *http.Request, productID string) {
	levels, err := h.svc.ListStock(r.Context(), productID)
	if err != nil {
		handleError(w, err)
		return
	}
	if levels == nil {
		levels = []*domain.StockLevel{}
	}
	writeJSON(w, http.StatusOK, levels)
}

// GET /inventory/{productID}/{warehouseID}
func (h *Handler) getStock(w http.ResponseWriter, r *http.Request, productID, warehouseID string) {
	sl, err := h.svc.GetStock(r.Context(), productID, warehouseID)
	if err != nil {
		handleError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, sl)
}

// PUT /inventory/{productID}/{warehouseID}
func (h *Handler) upsertStock(w http.ResponseWriter, r *http.Request, productID, warehouseID string) {
	var body struct {
		SKU       string `json:"sku"`
		Available int    `json:"available"`
		Reorder   int    `json:"reorder_point"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	sl, err := h.svc.UpsertStock(r.Context(), productID, body.SKU, warehouseID, body.Available, body.Reorder)
	if err != nil {
		handleError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, sl)
}

// -----------------------------------------------------------------------
// Reservations
// -----------------------------------------------------------------------

// POST /inventory/reserve
func (h *Handler) reserve(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		OrderID   string `json:"order_id"`
		ProductID string `json:"product_id"`
		Quantity  int    `json:"quantity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if body.OrderID == "" || body.ProductID == "" || body.Quantity <= 0 {
		http.Error(w, "order_id, product_id and quantity > 0 are required", http.StatusBadRequest)
		return
	}

	res, err := h.svc.Reserve(r.Context(), body.OrderID, body.ProductID, body.Quantity)
	if err != nil {
		handleError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, res)
}

// POST /inventory/release/{reservationID}
func (h *Handler) release(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/inventory/release/")
	if id == "" {
		http.Error(w, "reservation id required", http.StatusBadRequest)
		return
	}
	if err := h.svc.Release(r.Context(), id); err != nil {
		handleError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// POST /inventory/commit/{reservationID}
func (h *Handler) commit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/inventory/commit/")
	if id == "" {
		http.Error(w, "reservation id required", http.StatusBadRequest)
		return
	}
	if err := h.svc.Commit(r.Context(), id); err != nil {
		handleError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// -----------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("writeJSON: %v", err)
	}
}

func handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, domain.ErrInsufficientStock):
		http.Error(w, err.Error(), http.StatusConflict)
	default:
		log.Printf("internal error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}
