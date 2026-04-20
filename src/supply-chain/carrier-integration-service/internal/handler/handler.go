package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/shopos/carrier-integration-service/internal/domain"
	"github.com/shopos/carrier-integration-service/internal/registry"
)

// Registry is the interface the handler depends on. This makes testing easy.
type Registry interface {
	GetCarrier(id string) (*domain.Carrier, error)
	ListCarriers() []*domain.Carrier
	ListActive() []*domain.Carrier
	GetAllRates(req domain.RateQuoteRequest) ([]domain.RateQuoteResponse, error)
	CreateShipment(req domain.ShipmentRequest) (*domain.ShipmentResponse, error)
	GetTracking(trackingNumber string) (*domain.TrackResponse, error)
}

// Handler holds the HTTP handler state.
type Handler struct {
	reg Registry
	mux *http.ServeMux
}

// New creates a Handler with all routes registered.
func New(reg *registry.CarrierRegistry) *Handler {
	h := &Handler{reg: reg}
	h.mux = http.NewServeMux()
	h.registerRoutes()
	return h
}

// NewWithRegistry creates a Handler using the Registry interface (useful for testing).
func NewWithRegistry(reg Registry) *Handler {
	h := &Handler{reg: reg}
	h.mux = http.NewServeMux()
	h.registerRoutes()
	return h
}

func (h *Handler) registerRoutes() {
	h.mux.HandleFunc("/healthz", h.handleHealthz)
	h.mux.HandleFunc("/carriers", h.handleCarriers)
	h.mux.HandleFunc("/carriers/rates", h.handleRates)
	h.mux.HandleFunc("/carriers/tracking/", h.handleTracking)
	// catch-all for /carriers/{id} and /carriers/{id}/shipments
	h.mux.HandleFunc("/carriers/", h.handleCarrierDispatch)
}

// ServeHTTP implements http.Handler so the Handler can be used directly.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

// ─── route handlers ───────────────────────────────────────────────────────────

func (h *Handler) handleHealthz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handleCarriers handles GET /carriers (list all) — active-only when ?active=true.
func (h *Handler) handleCarriers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	var carriers []*domain.Carrier
	if r.URL.Query().Get("active") == "true" {
		carriers = h.reg.ListActive()
	} else {
		carriers = h.reg.ListCarriers()
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"carriers": carriers,
		"total":    len(carriers),
	})
}

// handleRates handles POST /carriers/rates — rate shopping across all carriers.
func (h *Handler) handleRates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	var req domain.RateQuoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		badRequest(w, "invalid JSON body: "+err.Error())
		return
	}
	if req.FromPostal == "" || req.ToPostal == "" {
		badRequest(w, "fromPostal and toPostal are required")
		return
	}
	if req.WeightKg <= 0 {
		badRequest(w, "weightKg must be greater than 0")
		return
	}

	quotes, err := h.reg.GetAllRates(req)
	if err != nil {
		internalError(w, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"quotes": quotes,
		"total":  len(quotes),
	})
}

// handleTracking handles GET /carriers/tracking/{trackingNumber}.
func (h *Handler) handleTracking(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	trackingNumber := strings.TrimPrefix(r.URL.Path, "/carriers/tracking/")
	if trackingNumber == "" {
		badRequest(w, "trackingNumber is required")
		return
	}

	resp, err := h.reg.GetTracking(trackingNumber)
	if err != nil {
		if errors.Is(err, domain.ErrTrackingNotFound) {
			notFound(w, "tracking number not found")
			return
		}
		internalError(w, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// handleCarrierDispatch dispatches /carriers/{id} and /carriers/{id}/shipments.
func (h *Handler) handleCarrierDispatch(w http.ResponseWriter, r *http.Request) {
	// strip leading /carriers/
	path := strings.TrimPrefix(r.URL.Path, "/carriers/")
	parts := strings.SplitN(path, "/", 2)

	carrierID := parts[0]
	if carrierID == "" {
		notFound(w, "not found")
		return
	}

	if len(parts) == 1 {
		// GET /carriers/{id}
		h.handleGetCarrier(w, r, carrierID)
		return
	}

	sub := parts[1]
	if sub == "shipments" {
		// POST /carriers/{id}/shipments
		h.handleCreateShipment(w, r, carrierID)
		return
	}

	notFound(w, "not found")
}

func (h *Handler) handleGetCarrier(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	carrier, err := h.reg.GetCarrier(id)
	if err != nil {
		if errors.Is(err, domain.ErrCarrierNotFound) {
			notFound(w, "carrier not found")
			return
		}
		internalError(w, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, carrier)
}

func (h *Handler) handleCreateShipment(w http.ResponseWriter, r *http.Request, carrierID string) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	var req domain.ShipmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		badRequest(w, "invalid JSON body: "+err.Error())
		return
	}
	req.CarrierID = carrierID
	if req.Service == "" {
		badRequest(w, "service is required")
		return
	}

	resp, err := h.reg.CreateShipment(req)
	if err != nil {
		if errors.Is(err, domain.ErrCarrierNotFound) {
			notFound(w, "carrier not found")
			return
		}
		if errors.Is(err, domain.ErrServiceNotFound) {
			badRequest(w, "service not supported by this carrier")
			return
		}
		internalError(w, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, resp)
}

// ─── response helpers ─────────────────────────────────────────────────────────

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func badRequest(w http.ResponseWriter, msg string) {
	writeJSON(w, http.StatusBadRequest, map[string]string{"error": msg})
}

func notFound(w http.ResponseWriter, msg string) {
	writeJSON(w, http.StatusNotFound, map[string]string{"error": msg})
}

func methodNotAllowed(w http.ResponseWriter) {
	writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
}

func internalError(w http.ResponseWriter, msg string) {
	writeJSON(w, http.StatusInternalServerError, map[string]string{"error": msg})
}
