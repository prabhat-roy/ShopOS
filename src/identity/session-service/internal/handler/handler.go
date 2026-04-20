package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/shopos/session-service/internal/domain"
)

// Servicer is the business-logic contract required by the HTTP handler layer.
type Servicer interface {
	Create(ctx context.Context, req *domain.CreateSessionRequest) (*domain.Session, error)
	Validate(ctx context.Context, id string) (*domain.Session, error)
	Get(ctx context.Context, id string) (*domain.Session, error)
	Delete(ctx context.Context, id string) error
	ListByUser(ctx context.Context, userID string) ([]*domain.Session, error)
	DeleteAllByUser(ctx context.Context, userID string) error
}

// Handler aggregates the HTTP mux and service layer.
type Handler struct {
	mux     *http.ServeMux
	service Servicer
}

// New builds and wires up all routes, returning an http.Handler ready to serve.
func New(svc Servicer) http.Handler {
	h := &Handler{
		mux:     http.NewServeMux(),
		service: svc,
	}
	h.registerRoutes()
	return h.mux
}

func (h *Handler) registerRoutes() {
	h.mux.HandleFunc("/healthz", h.handleHealth)
	h.mux.HandleFunc("/sessions", h.handleSessions)
	h.mux.HandleFunc("/sessions/", h.handleSessionByID)
}

// --------------------------------------------------------------------------
// Route: GET /healthz
// --------------------------------------------------------------------------

func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// --------------------------------------------------------------------------
// Route: POST /sessions
// --------------------------------------------------------------------------

func (h *Handler) handleSessions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.createSession(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) createSession(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	session, err := h.service.Create(r.Context(), &req)
	if err != nil {
		log.Printf("createSession error: %v", err)
		writeError(w, http.StatusInternalServerError, "could not create session")
		return
	}

	writeJSON(w, http.StatusCreated, session)
}

// --------------------------------------------------------------------------
// Route: /sessions/{id} and /sessions/user/{userID}
// --------------------------------------------------------------------------

// handleSessionByID dispatches to the correct sub-handler based on the URL
// path structure:
//
//	/sessions/{id}            — GET (validate) or DELETE
//	/sessions/user/{userID}   — GET (list) or DELETE (delete all)
func (h *Handler) handleSessionByID(w http.ResponseWriter, r *http.Request) {
	// Strip leading /sessions/ prefix.
	path := strings.TrimPrefix(r.URL.Path, "/sessions/")
	if path == "" {
		writeError(w, http.StatusBadRequest, "missing path segment")
		return
	}

	// /sessions/user/{userID}
	if strings.HasPrefix(path, "user/") {
		userID := strings.TrimPrefix(path, "user/")
		if userID == "" {
			writeError(w, http.StatusBadRequest, "missing user ID")
			return
		}
		switch r.Method {
		case http.MethodGet:
			h.listByUser(w, r, userID)
		case http.MethodDelete:
			h.deleteAllByUser(w, r, userID)
		default:
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	// /sessions/{id}
	id := path
	switch r.Method {
	case http.MethodGet:
		h.getSession(w, r, id)
	case http.MethodDelete:
		h.deleteSession(w, r, id)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// GET /sessions/{id} — validates the session (and extends TTL on success)
func (h *Handler) getSession(w http.ResponseWriter, r *http.Request, id string) {
	session, err := h.service.Validate(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, "session not found")
			return
		}
		if errors.Is(err, domain.ErrExpired) {
			writeError(w, http.StatusUnauthorized, "session expired")
			return
		}
		log.Printf("getSession error: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, session)
}

// DELETE /sessions/{id}
func (h *Handler) deleteSession(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.service.Delete(r.Context(), id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, "session not found")
			return
		}
		log.Printf("deleteSession error: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GET /sessions/user/{userID}
func (h *Handler) listByUser(w http.ResponseWriter, r *http.Request, userID string) {
	sessions, err := h.service.ListByUser(r.Context(), userID)
	if err != nil {
		log.Printf("listByUser error: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, sessions)
}

// DELETE /sessions/user/{userID}
func (h *Handler) deleteAllByUser(w http.ResponseWriter, r *http.Request, userID string) {
	if err := h.service.DeleteAllByUser(r.Context(), userID); err != nil {
		log.Printf("deleteAllByUser error: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --------------------------------------------------------------------------
// Helpers
// --------------------------------------------------------------------------

type errorResponse struct {
	Error     string    `json:"error"`
	Timestamp time.Time `json:"timestamp"`
}

func writeJSON(w http.ResponseWriter, status int, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Printf("writeJSON encode error: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorResponse{
		Error:     msg,
		Timestamp: time.Now().UTC(),
	})
}
