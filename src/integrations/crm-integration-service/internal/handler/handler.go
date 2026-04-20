package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/shopos/crm-integration-service/internal/domain"
	"github.com/shopos/crm-integration-service/internal/service"
)

// Handler holds all HTTP handler dependencies.
type Handler struct {
	svc *service.Servicer
}

// New returns an initialised Handler.
func New(svc *service.Servicer) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes attaches all routes to mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/healthz", h.healthz)
	mux.HandleFunc("/crm/contacts/sync", h.syncContacts)
	mux.HandleFunc("/crm/deals/sync", h.syncDeals)
	mux.HandleFunc("/crm/contacts/", h.contactByCrmSystemAndID)
	mux.HandleFunc("/crm/contacts", h.listContacts)
	mux.HandleFunc("/crm/sync-history", h.syncHistory)
	mux.HandleFunc("/crm/field-mappings/", h.fieldMappings)
}

// healthz returns service health.
func (h *Handler) healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// syncContacts handles POST /crm/contacts/sync
func (h *Handler) syncContacts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var body struct {
		CrmSystem domain.CrmSystem         `json:"crmSystem"`
		Customers []map[string]interface{} `json:"customers"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if !domain.ValidCrmSystem(body.CrmSystem) {
		writeError(w, http.StatusBadRequest, "invalid crmSystem; must be SALESFORCE, HUBSPOT, ZOHO, or PIPEDRIVE")
		return
	}
	if len(body.Customers) == 0 {
		writeError(w, http.StatusBadRequest, "customers list must not be empty")
		return
	}

	result := h.svc.SyncContacts(body.CrmSystem, body.Customers)
	writeJSON(w, http.StatusAccepted, result)
}

// syncDeals handles POST /crm/deals/sync
func (h *Handler) syncDeals(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var body struct {
		CrmSystem domain.CrmSystem         `json:"crmSystem"`
		Orders    []map[string]interface{} `json:"orders"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if !domain.ValidCrmSystem(body.CrmSystem) {
		writeError(w, http.StatusBadRequest, "invalid crmSystem; must be SALESFORCE, HUBSPOT, ZOHO, or PIPEDRIVE")
		return
	}
	if len(body.Orders) == 0 {
		writeError(w, http.StatusBadRequest, "orders list must not be empty")
		return
	}

	result := h.svc.SyncDeals(body.CrmSystem, body.Orders)
	writeJSON(w, http.StatusAccepted, result)
}

// contactByCrmSystemAndID handles GET /crm/contacts/{crmSystem}/{crmId}
func (h *Handler) contactByCrmSystemAndID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Path: /crm/contacts/{crmSystem}/{crmId}
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/crm/contacts/"), "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		writeError(w, http.StatusBadRequest, "path must be /crm/contacts/{crmSystem}/{crmId}")
		return
	}

	crmSystem := domain.CrmSystem(strings.ToUpper(parts[0]))
	if !domain.ValidCrmSystem(crmSystem) {
		writeError(w, http.StatusBadRequest, "invalid crmSystem")
		return
	}

	contact, err := h.svc.GetContact(crmSystem, parts[1])
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, contact)
}

// listContacts handles GET /crm/contacts?crmSystem=
func (h *Handler) listContacts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	crmSystem := domain.CrmSystem(r.URL.Query().Get("crmSystem"))
	if crmSystem != "" && !domain.ValidCrmSystem(crmSystem) {
		writeError(w, http.StatusBadRequest, "invalid crmSystem filter")
		return
	}

	contacts := h.svc.ListContacts(crmSystem)
	writeJSON(w, http.StatusOK, map[string]interface{}{"contacts": contacts, "count": len(contacts)})
}

// syncHistory handles GET /crm/sync-history?crmSystem=&limit=
func (h *Handler) syncHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	q := r.URL.Query()
	crmSystem := domain.CrmSystem(q.Get("crmSystem"))
	if crmSystem != "" && !domain.ValidCrmSystem(crmSystem) {
		writeError(w, http.StatusBadRequest, "invalid crmSystem filter")
		return
	}

	limit := 20
	if l := q.Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}

	history := h.svc.GetSyncHistory(crmSystem, limit)
	writeJSON(w, http.StatusOK, map[string]interface{}{"history": history, "count": len(history)})
}

// fieldMappings handles GET /crm/field-mappings/{crmSystem}
func (h *Handler) fieldMappings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	cs := domain.CrmSystem(strings.ToUpper(strings.TrimPrefix(r.URL.Path, "/crm/field-mappings/")))
	if !domain.ValidCrmSystem(cs) {
		writeError(w, http.StatusBadRequest, "invalid crmSystem; must be SALESFORCE, HUBSPOT, ZOHO, or PIPEDRIVE")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"crmSystem": cs,
		"mappings":  h.svc.GetFieldMapping(cs),
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
