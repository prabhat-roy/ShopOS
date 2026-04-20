package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/shopos/logistics-provider-integration/internal/domain"
	"github.com/shopos/logistics-provider-integration/internal/service"
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
	h.mux.HandleFunc("/providers", h.handleProviders)
	h.mux.HandleFunc("/shipments/rates", h.handleRates)  // must be before /shipments/{id}
	h.mux.HandleFunc("/shipments", h.handleShipments)
	h.mux.HandleFunc("/shipments/", h.handleShipmentByID)
	return h
}

// ServeHTTP satisfies http.Handler so the struct can be used directly.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

// ---------------------------------------------------------------------------
// route handlers
// ---------------------------------------------------------------------------

// handleHealth returns a simple liveness probe response.
// GET /healthz
func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handleProviders returns all supported logistics providers.
// GET /providers
func (h *Handler) handleProviders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"providers": domain.AllProviders,
	})
}

// handleShipments dispatches to create (POST) or list (GET).
// POST /shipments         — create a shipment
// GET  /shipments         — list shipments (query: provider, limit)
func (h *Handler) handleShipments(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.createShipment(w, r)
	case http.MethodGet:
		h.listShipments(w, r)
	default:
		methodNotAllowed(w)
	}
}

// handleRates handles rate-shopping requests.
// POST /shipments/rates
func (h *Handler) handleRates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}

	var req struct {
		FromPostal string  `json:"fromPostal"`
		ToPostal   string  `json:"toPostal"`
		WeightKg   float64 `json:"weightKg"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	quotes, err := h.svc.GetRates(req.FromPostal, req.ToPostal, req.WeightKg)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"rates": quotes})
}

// handleShipmentByID handles:
//   GET /shipments/{trackingNumber}           — fetch a stored shipment
//   GET /shipments/{trackingNumber}/track     — live tracking (query: provider)
func (h *Handler) handleShipmentByID(w http.ResponseWriter, r *http.Request) {
	// Strip leading /shipments/
	path := strings.TrimPrefix(r.URL.Path, "/shipments/")
	parts := strings.SplitN(path, "/", 2)
	trackingNumber := parts[0]

	if len(parts) == 2 && parts[1] == "track" {
		// GET /shipments/{trackingNumber}/track?provider=FEDEX
		if r.Method != http.MethodGet {
			methodNotAllowed(w)
			return
		}
		h.trackShipment(w, r, trackingNumber)
		return
	}

	// GET /shipments/{trackingNumber}
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	h.getShipment(w, r, trackingNumber)
}

// ---------------------------------------------------------------------------
// sub-handlers
// ---------------------------------------------------------------------------

func (h *Handler) createShipment(w http.ResponseWriter, r *http.Request) {
	var req domain.ShipmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	resp, err := h.svc.CreateShipment(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, resp)
}

func (h *Handler) getShipment(w http.ResponseWriter, r *http.Request, trackingNumber string) {
	resp, err := h.svc.GetShipment(trackingNumber)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) trackShipment(w http.ResponseWriter, r *http.Request, trackingNumber string) {
	providerStr := r.URL.Query().Get("provider")
	if providerStr == "" {
		writeError(w, http.StatusBadRequest, "query parameter 'provider' is required")
		return
	}
	provider := domain.Provider(strings.ToUpper(providerStr))

	resp, err := h.svc.TrackShipment(trackingNumber, provider)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) listShipments(w http.ResponseWriter, r *http.Request) {
	provider := r.URL.Query().Get("provider")
	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if n, err := strconv.Atoi(limitStr); err == nil && n > 0 {
			limit = n
		}
	}

	shipments := h.svc.ListShipments(provider, limit)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"shipments": shipments,
		"count":     len(shipments),
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
