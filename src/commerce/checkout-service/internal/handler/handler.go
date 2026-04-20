package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/shopos/checkout-service/internal/domain"
	"github.com/shopos/checkout-service/internal/service"
)

// Handler holds a reference to the business-logic layer and exposes HTTP routes.
type Handler struct {
	svc    service.Servicer
	mux    *http.ServeMux
	logger *log.Logger
}

// New creates a Handler and registers all routes on a new ServeMux.
func New(svc service.Servicer, logger *log.Logger) *Handler {
	h := &Handler{
		svc:    svc,
		mux:    http.NewServeMux(),
		logger: logger,
	}
	h.registerRoutes()
	return h
}

// ServeHTTP satisfies the http.Handler interface.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

// registerRoutes wires URL patterns to handler methods.
func (h *Handler) registerRoutes() {
	// Health
	h.mux.HandleFunc("/healthz", h.healthz)

	// Session collection
	h.mux.HandleFunc("/checkout/sessions", h.sessions)

	// Individual session — dispatches on trailing path segment and method
	h.mux.HandleFunc("/checkout/sessions/", h.sessionByID)
}

// ─── /healthz ────────────────────────────────────────────────────────────────

func (h *Handler) healthz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ─── /checkout/sessions ──────────────────────────────────────────────────────

func (h *Handler) sessions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.initiateSession(w, r)
	default:
		methodNotAllowed(w)
	}
}

// POST /checkout/sessions — creates a new checkout session.
func (h *Handler) initiateSession(w http.ResponseWriter, r *http.Request) {
	var req domain.InitiateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}
	if req.UserID == "" {
		writeError(w, http.StatusBadRequest, "user_id is required")
		return
	}
	if req.CartID == "" {
		writeError(w, http.StatusBadRequest, "cart_id is required")
		return
	}

	session, err := h.svc.Initiate(r.Context(), req)
	if err != nil {
		h.logger.Printf("initiate session error: %v", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, session)
}

// ─── /checkout/sessions/{id}[/confirm] ───────────────────────────────────────

// sessionByID dispatches sub-routes under /checkout/sessions/{id}.
func (h *Handler) sessionByID(w http.ResponseWriter, r *http.Request) {
	// Strip the prefix to get "/{id}" or "/{id}/confirm"
	suffix := strings.TrimPrefix(r.URL.Path, "/checkout/sessions/")
	parts := strings.SplitN(suffix, "/", 2)
	sessionID := parts[0]

	if sessionID == "" {
		writeError(w, http.StatusNotFound, "session id is required")
		return
	}

	// Sub-resource: /checkout/sessions/{id}/confirm
	if len(parts) == 2 && parts[1] == "confirm" {
		if r.Method != http.MethodPost {
			methodNotAllowed(w)
			return
		}
		h.confirmSession(w, r, sessionID)
		return
	}

	// Plain resource: /checkout/sessions/{id}
	switch r.Method {
	case http.MethodGet:
		h.getSession(w, r, sessionID)
	case http.MethodDelete:
		h.cancelSession(w, r, sessionID)
	default:
		methodNotAllowed(w)
	}
}

// POST /checkout/sessions/{id}/confirm — confirms a pending checkout session.
func (h *Handler) confirmSession(w http.ResponseWriter, r *http.Request, sessionID string) {
	var req domain.ConfirmRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}
	req.SessionID = sessionID

	if req.PaymentMethodID == "" {
		writeError(w, http.StatusBadRequest, "payment_method_id is required")
		return
	}

	session, err := h.svc.Confirm(r.Context(), req)
	if err != nil {
		h.logAndRespond(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, session)
}

// GET /checkout/sessions/{id} — fetches an existing session.
func (h *Handler) getSession(w http.ResponseWriter, r *http.Request, sessionID string) {
	session, err := h.svc.GetSession(r.Context(), sessionID)
	if err != nil {
		h.logAndRespond(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, session)
}

// DELETE /checkout/sessions/{id} — cancels a pending session.
func (h *Handler) cancelSession(w http.ResponseWriter, r *http.Request, sessionID string) {
	if err := h.svc.CancelSession(r.Context(), sessionID); err != nil {
		h.logAndRespond(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ─── helpers ─────────────────────────────────────────────────────────────────

// logAndRespond logs the error and writes an appropriate HTTP response.
func (h *Handler) logAndRespond(w http.ResponseWriter, _ *http.Request, err error) {
	h.logger.Printf("handler error: %v", err)
	msg := err.Error()
	if strings.Contains(msg, "not found") {
		writeError(w, http.StatusNotFound, msg)
		return
	}
	writeError(w, http.StatusInternalServerError, msg)
}

// writeJSON serialises v as JSON and writes it with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("writeJSON encode error: %v", err)
	}
}

// writeError writes a JSON error response.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, domain.ErrorResponse{Error: msg})
}

// methodNotAllowed writes a 405 response.
func methodNotAllowed(w http.ResponseWriter) {
	writeError(w, http.StatusMethodNotAllowed, "method not allowed")
}
