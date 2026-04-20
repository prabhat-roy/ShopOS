package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/shopos/marketplace-connector-service/internal/adapter"
	"github.com/shopos/marketplace-connector-service/internal/domain"
	"github.com/shopos/marketplace-connector-service/internal/service"
)

// Handler holds all HTTP handler dependencies.
type Handler struct {
	svc     *service.Servicer
	adapter *adapter.MarketplaceAdapter
}

// New returns an initialised Handler.
func New(svc *service.Servicer, ad *adapter.MarketplaceAdapter) *Handler {
	return &Handler{svc: svc, adapter: ad}
}

// RegisterRoutes attaches all routes to mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/healthz", h.healthz)
	mux.HandleFunc("/marketplace/products/sync", h.syncProducts)
	mux.HandleFunc("/marketplace/orders/sync", h.syncOrders)
	mux.HandleFunc("/marketplace/syncs/", h.syncByID)
	mux.HandleFunc("/marketplace/syncs", h.listSyncs)
	mux.HandleFunc("/marketplace/stats", h.stats)
	mux.HandleFunc("/marketplace/field-mappings/", h.fieldMappings)
}

// healthz returns service health.
func (h *Handler) healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// syncProducts handles POST /marketplace/products/sync
func (h *Handler) syncProducts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var body struct {
		Marketplace domain.Marketplace       `json:"marketplace"`
		Products    []map[string]interface{} `json:"products"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if !domain.ValidMarketplace(body.Marketplace) {
		writeError(w, http.StatusBadRequest, "invalid marketplace; must be AMAZON, EBAY, ETSY, or WALMART")
		return
	}
	if len(body.Products) == 0 {
		writeError(w, http.StatusBadRequest, "products list must not be empty")
		return
	}

	rec := h.svc.SyncProducts(body.Marketplace, body.Products)
	writeJSON(w, http.StatusAccepted, rec)
}

// syncOrders handles POST /marketplace/orders/sync
func (h *Handler) syncOrders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var body struct {
		Marketplace domain.Marketplace `json:"marketplace"`
		Limit       int                `json:"limit"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if !domain.ValidMarketplace(body.Marketplace) {
		writeError(w, http.StatusBadRequest, "invalid marketplace; must be AMAZON, EBAY, ETSY, or WALMART")
		return
	}

	rec := h.svc.SyncOrders(body.Marketplace, body.Limit)
	writeJSON(w, http.StatusAccepted, rec)
}

// syncByID handles GET /marketplace/syncs/{id}
func (h *Handler) syncByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/marketplace/syncs/")
	if id == "" {
		writeError(w, http.StatusBadRequest, "sync id required")
		return
	}

	rec, err := h.svc.GetSync(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, rec)
}

// listSyncs handles GET /marketplace/syncs?marketplace=&limit=
func (h *Handler) listSyncs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	q := r.URL.Query()
	marketplace := domain.Marketplace(q.Get("marketplace"))
	if marketplace != "" && !domain.ValidMarketplace(marketplace) {
		writeError(w, http.StatusBadRequest, "invalid marketplace filter")
		return
	}

	limit := 20
	if l := q.Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}

	records := h.svc.ListSyncs(marketplace, limit)
	writeJSON(w, http.StatusOK, map[string]interface{}{"syncs": records, "count": len(records)})
}

// stats handles GET /marketplace/stats
func (h *Handler) stats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	writeJSON(w, http.StatusOK, h.svc.GetStats())
}

// fieldMappings handles GET /marketplace/field-mappings/{marketplace}
func (h *Handler) fieldMappings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	mp := domain.Marketplace(strings.TrimPrefix(r.URL.Path, "/marketplace/field-mappings/"))
	if !domain.ValidMarketplace(mp) {
		writeError(w, http.StatusBadRequest, "invalid marketplace; must be AMAZON, EBAY, ETSY, or WALMART")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"marketplace": mp,
		"mappings":    h.adapter.GetFieldMapping(mp),
	})
}

// --- helpers ---

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
