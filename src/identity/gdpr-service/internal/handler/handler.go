package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/shopos/gdpr-service/internal/domain"
)

// Servicer is the business-logic interface consumed by the HTTP handler.
type Servicer interface {
	SubmitRequest(ctx context.Context, userID string, reqType domain.RequestType, reason string) (*domain.DataRequest, error)
	GetRequest(ctx context.Context, id string) (*domain.DataRequest, error)
	ListRequests(ctx context.Context, userID string) ([]*domain.DataRequest, error)
	ProcessRequest(ctx context.Context, id string, notes string) error
	CompleteRequest(ctx context.Context, id string, notes string) error
	RejectRequest(ctx context.Context, id string, notes string) error
	UpdateConsent(ctx context.Context, userID string, consentType domain.ConsentType, granted bool, ip string) error
	GetConsents(ctx context.Context, userID string) ([]*domain.Consent, error)
	CheckConsent(ctx context.Context, userID string, consentType domain.ConsentType) (bool, error)
}

// Handler wires all HTTP routes for the gdpr-service.
type Handler struct {
	svc Servicer
	mux *http.ServeMux
}

// New creates a Handler and registers all routes on a fresh ServeMux.
func New(svc Servicer) *Handler {
	h := &Handler{svc: svc, mux: http.NewServeMux()}
	h.registerRoutes()
	return h
}

// ServeHTTP satisfies http.Handler so the Handler can be passed directly to
// http.ListenAndServe.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func (h *Handler) registerRoutes() {
	h.mux.HandleFunc("/healthz", h.healthz)
	h.mux.HandleFunc("/gdpr/requests", h.requestsCollection)
	h.mux.HandleFunc("/gdpr/requests/", h.requestsItem)
	h.mux.HandleFunc("/gdpr/users/", h.usersRouter)
}

// ---------- route dispatchers ----------

// requestsCollection handles POST /gdpr/requests
func (h *Handler) requestsCollection(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	h.submitRequest(w, r)
}

// requestsItem dispatches on /gdpr/requests/{id}[/action]
func (h *Handler) requestsItem(w http.ResponseWriter, r *http.Request) {
	// strip leading "/gdpr/requests/"
	path := strings.TrimPrefix(r.URL.Path, "/gdpr/requests/")
	parts := strings.SplitN(path, "/", 2)
	id := parts[0]
	if id == "" {
		http.NotFound(w, r)
		return
	}

	if len(parts) == 1 {
		// /gdpr/requests/{id}
		if r.Method == http.MethodGet {
			h.getRequest(w, r, id)
			return
		}
		methodNotAllowed(w)
		return
	}

	// /gdpr/requests/{id}/{action}
	action := parts[1]
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	switch action {
	case "process":
		h.processRequest(w, r, id)
	case "complete":
		h.completeRequest(w, r, id)
	case "reject":
		h.rejectRequest(w, r, id)
	default:
		http.NotFound(w, r)
	}
}

// usersRouter dispatches on /gdpr/users/{userID}/...
func (h *Handler) usersRouter(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/gdpr/users/")
	// path = "{userID}/{sub}"
	parts := strings.SplitN(path, "/", 2)
	if len(parts) < 2 || parts[0] == "" {
		http.NotFound(w, r)
		return
	}
	userID := parts[0]
	sub := parts[1]

	switch sub {
	case "requests":
		if r.Method != http.MethodGet {
			methodNotAllowed(w)
			return
		}
		h.listRequests(w, r, userID)
	case "consent":
		switch r.Method {
		case http.MethodPut:
			h.updateConsent(w, r, userID)
		case http.MethodGet:
			h.getConsents(w, r, userID)
		default:
			methodNotAllowed(w)
		}
	default:
		http.NotFound(w, r)
	}
}

// ---------- handler methods ----------

func (h *Handler) healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// submitRequest handles POST /gdpr/requests
func (h *Handler) submitRequest(w http.ResponseWriter, r *http.Request) {
	var body struct {
		UserID string             `json:"user_id"`
		Type   domain.RequestType `json:"type"`
		Reason string             `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if body.UserID == "" || body.Type == "" {
		writeError(w, http.StatusBadRequest, "user_id and type are required")
		return
	}

	req, err := h.svc.SubmitRequest(r.Context(), body.UserID, body.Type, body.Reason)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, req)
}

// getRequest handles GET /gdpr/requests/{id}
func (h *Handler) getRequest(w http.ResponseWriter, r *http.Request, id string) {
	req, err := h.svc.GetRequest(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, "request not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, req)
}

// listRequests handles GET /gdpr/users/{userID}/requests
func (h *Handler) listRequests(w http.ResponseWriter, r *http.Request, userID string) {
	reqs, err := h.svc.ListRequests(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if reqs == nil {
		reqs = []*domain.DataRequest{}
	}
	writeJSON(w, http.StatusOK, reqs)
}

// processRequest handles POST /gdpr/requests/{id}/process
func (h *Handler) processRequest(w http.ResponseWriter, r *http.Request, id string) {
	notes := parseNotes(r)
	if err := h.svc.ProcessRequest(r.Context(), id, notes); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, "request not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// completeRequest handles POST /gdpr/requests/{id}/complete
func (h *Handler) completeRequest(w http.ResponseWriter, r *http.Request, id string) {
	notes := parseNotes(r)
	if err := h.svc.CompleteRequest(r.Context(), id, notes); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, "request not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// rejectRequest handles POST /gdpr/requests/{id}/reject
func (h *Handler) rejectRequest(w http.ResponseWriter, r *http.Request, id string) {
	notes := parseNotes(r)
	if err := h.svc.RejectRequest(r.Context(), id, notes); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, "request not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// updateConsent handles PUT /gdpr/users/{userID}/consent
func (h *Handler) updateConsent(w http.ResponseWriter, r *http.Request, userID string) {
	var body struct {
		Type      domain.ConsentType `json:"type"`
		Granted   bool               `json:"granted"`
		IPAddress string             `json:"ip_address"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if body.Type == "" {
		writeError(w, http.StatusBadRequest, "type is required")
		return
	}

	if err := h.svc.UpdateConsent(r.Context(), userID, body.Type, body.Granted, body.IPAddress); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// getConsents handles GET /gdpr/users/{userID}/consent
func (h *Handler) getConsents(w http.ResponseWriter, r *http.Request, userID string) {
	consents, err := h.svc.GetConsents(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if consents == nil {
		consents = []*domain.Consent{}
	}
	writeJSON(w, http.StatusOK, consents)
}

// ---------- helpers ----------

// parseNotes decodes an optional {"notes":"..."} JSON body; ignores errors.
func parseNotes(r *http.Request) string {
	var body struct {
		Notes string `json:"notes"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	return body.Notes
}

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
