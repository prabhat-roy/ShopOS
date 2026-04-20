package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/shopos/warehouse-service/internal/domain"
	"github.com/shopos/warehouse-service/internal/service"
)

// Handler holds the HTTP mux and a reference to the service layer.
type Handler struct {
	mux *http.ServeMux
	svc service.Servicer
}

// New wires all routes onto a fresh ServeMux and returns the Handler.
func New(svc service.Servicer) *Handler {
	h := &Handler{mux: http.NewServeMux(), svc: svc}
	h.routes()
	return h
}

// ServeHTTP satisfies http.Handler so Handler can be passed directly to http.Server.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func (h *Handler) routes() {
	h.mux.HandleFunc("/healthz", h.healthz)
	h.mux.HandleFunc("/warehouses", h.warehouses)
	h.mux.HandleFunc("/warehouses/", h.warehouseByID)
}

// ---- helpers ----------------------------------------------------------------

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func decode(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}

// ---- /healthz ---------------------------------------------------------------

func (h *Handler) healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ---- /warehouses ------------------------------------------------------------

func (h *Handler) warehouses(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listWarehouses(w, r)
	case http.MethodPost:
		h.createWarehouse(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) createWarehouse(w http.ResponseWriter, r *http.Request) {
	var body domain.Warehouse
	if err := decode(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	created, err := h.svc.CreateWarehouse(&body)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *Handler) listWarehouses(w http.ResponseWriter, r *http.Request) {
	activeOnly := r.URL.Query().Get("active") == "true"
	warehouses, err := h.svc.ListWarehouses(activeOnly)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if warehouses == nil {
		warehouses = []*domain.Warehouse{}
	}
	writeJSON(w, http.StatusOK, warehouses)
}

// ---- /warehouses/{id} and sub-routes ----------------------------------------

// warehouseByID dispatches based on the URL path segments after /warehouses/.
//
// Handled patterns:
//
//	/warehouses/{id}                   GET, PATCH
//	/warehouses/{id}/receive           POST
//	/warehouses/{id}/ship              POST
//	/warehouses/{id}/stock             GET  (?productId=)
//	/warehouses/{id}/movements         GET
func (h *Handler) warehouseByID(w http.ResponseWriter, r *http.Request) {
	// Trim leading slash and split: ["warehouses", "{id}", ...]
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")
	// parts[0] == "warehouses"
	if len(parts) < 2 || parts[1] == "" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	id := parts[1]

	if len(parts) == 2 {
		// /warehouses/{id}
		switch r.Method {
		case http.MethodGet:
			h.getWarehouse(w, r, id)
		case http.MethodPatch:
			h.updateWarehouse(w, r, id)
		default:
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	subResource := parts[2]
	switch subResource {
	case "receive":
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		h.receiveStock(w, r, id)

	case "ship":
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		h.shipStock(w, r, id)

	case "stock":
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		h.getStock(w, r, id)

	case "movements":
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		h.listMovements(w, r, id)

	default:
		writeError(w, http.StatusNotFound, "not found")
	}
}

func (h *Handler) getWarehouse(w http.ResponseWriter, r *http.Request, id string) {
	warehouse, err := h.svc.GetWarehouse(id)
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, "warehouse not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, warehouse)
}

func (h *Handler) updateWarehouse(w http.ResponseWriter, r *http.Request, id string) {
	var body domain.Warehouse
	if err := decode(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	updated, err := h.svc.UpdateWarehouse(id, &body)
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, "warehouse not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (h *Handler) receiveStock(w http.ResponseWriter, r *http.Request, warehouseID string) {
	var body domain.StockMovement
	if err := decode(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	body.WarehouseID = warehouseID
	movement, err := h.svc.ReceiveStock(&body)
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, "warehouse not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, movement)
}

func (h *Handler) shipStock(w http.ResponseWriter, r *http.Request, warehouseID string) {
	var body domain.StockMovement
	if err := decode(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	body.WarehouseID = warehouseID
	movement, err := h.svc.ShipStock(&body)
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, "warehouse not found")
		return
	}
	if errors.Is(err, domain.ErrInsufficientStock) {
		writeError(w, http.StatusConflict, "insufficient stock")
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, movement)
}

func (h *Handler) getStock(w http.ResponseWriter, r *http.Request, warehouseID string) {
	productID := r.URL.Query().Get("productId")
	if productID == "" {
		writeError(w, http.StatusBadRequest, "productId query parameter is required")
		return
	}
	level, err := h.svc.GetStockLevel(warehouseID, productID)
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, "warehouse not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]int{"stock": level})
}

func (h *Handler) listMovements(w http.ResponseWriter, r *http.Request, warehouseID string) {
	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}
	movements, err := h.svc.ListMovements(warehouseID, limit)
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, "warehouse not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if movements == nil {
		movements = []*domain.StockMovement{}
	}
	writeJSON(w, http.StatusOK, movements)
}
