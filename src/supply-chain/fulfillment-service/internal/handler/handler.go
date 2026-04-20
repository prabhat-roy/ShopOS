package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/shopos/fulfillment-service/internal/domain"
	"github.com/shopos/fulfillment-service/internal/service"
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
	h.mux.HandleFunc("/fulfillments", h.fulfillments)
	h.mux.HandleFunc("/fulfillments/", h.fulfillmentDispatch)
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

func errStatus(err error) int {
	if errors.Is(err, domain.ErrNotFound) {
		return http.StatusNotFound
	}
	if errors.Is(err, domain.ErrInvalidTransition) {
		return http.StatusConflict
	}
	return http.StatusInternalServerError
}

// ---- /healthz ---------------------------------------------------------------

func (h *Handler) healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ---- /fulfillments ----------------------------------------------------------

func (h *Handler) fulfillments(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listFulfillments(w, r)
	case http.MethodPost:
		h.createFulfillment(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) createFulfillment(w http.ResponseWriter, r *http.Request) {
	var body domain.FulfillmentOrder
	if err := decode(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	created, err := h.svc.CreateFulfillment(&body)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *Handler) listFulfillments(w http.ResponseWriter, r *http.Request) {
	orderID := r.URL.Query().Get("orderId")
	list, err := h.svc.ListFulfillments(orderID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if list == nil {
		list = []*domain.FulfillmentOrder{}
	}
	writeJSON(w, http.StatusOK, list)
}

// ---- /fulfillments/{id} and sub-routes --------------------------------------
//
// Handled patterns:
//
//	/fulfillments/{id}                     GET
//	/fulfillments/order/{orderId}          GET
//	/fulfillments/{id}/pick                POST → 204
//	/fulfillments/{id}/pack                POST → 204
//	/fulfillments/{id}/ready               POST → 204
//	/fulfillments/{id}/ship                POST → 204 (body: {trackingNumber,carrier})
//	/fulfillments/{id}/deliver             POST → 204
//	/fulfillments/{id}                     DELETE → cancel → 204
func (h *Handler) fulfillmentDispatch(w http.ResponseWriter, r *http.Request) {
	// Trim leading slash and split: ["fulfillments", seg1, seg2?, ...]
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")
	// parts[0] == "fulfillments"
	if len(parts) < 2 || parts[1] == "" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	seg1 := parts[1] // could be "order" or an ID
	if seg1 == "order" {
		// /fulfillments/order/{orderId}
		if len(parts) < 3 || parts[2] == "" {
			writeError(w, http.StatusBadRequest, "orderId is required")
			return
		}
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		h.getFulfillmentByOrder(w, r, parts[2])
		return
	}

	id := seg1

	if len(parts) == 2 {
		// /fulfillments/{id}
		switch r.Method {
		case http.MethodGet:
			h.getFulfillment(w, r, id)
		case http.MethodDelete:
			h.cancelFulfillment(w, r, id)
		default:
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	action := parts[2]
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	switch action {
	case "pick":
		h.pickFulfillment(w, r, id)
	case "pack":
		h.packFulfillment(w, r, id)
	case "ready":
		h.readyFulfillment(w, r, id)
	case "ship":
		h.shipFulfillment(w, r, id)
	case "deliver":
		h.deliverFulfillment(w, r, id)
	default:
		writeError(w, http.StatusNotFound, "not found")
	}
}

func (h *Handler) getFulfillment(w http.ResponseWriter, r *http.Request, id string) {
	f, err := h.svc.GetFulfillment(id)
	if err != nil {
		writeError(w, errStatus(err), err.Error())
		return
	}
	writeJSON(w, http.StatusOK, f)
}

func (h *Handler) getFulfillmentByOrder(w http.ResponseWriter, r *http.Request, orderID string) {
	f, err := h.svc.GetByOrderID(orderID)
	if err != nil {
		writeError(w, errStatus(err), err.Error())
		return
	}
	writeJSON(w, http.StatusOK, f)
}

func (h *Handler) pickFulfillment(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.svc.StartPicking(id); err != nil {
		writeError(w, errStatus(err), err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) packFulfillment(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.svc.StartPacking(id); err != nil {
		writeError(w, errStatus(err), err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) readyFulfillment(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.svc.MarkReadyToShip(id); err != nil {
		writeError(w, errStatus(err), err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type shipRequest struct {
	TrackingNumber string `json:"trackingNumber"`
	Carrier        string `json:"carrier"`
}

func (h *Handler) shipFulfillment(w http.ResponseWriter, r *http.Request, id string) {
	var body shipRequest
	if err := decode(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if err := h.svc.Ship(id, body.TrackingNumber, body.Carrier); err != nil {
		writeError(w, errStatus(err), err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) deliverFulfillment(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.svc.Deliver(id); err != nil {
		writeError(w, errStatus(err), err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) cancelFulfillment(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.svc.Cancel(id); err != nil {
		writeError(w, errStatus(err), err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
