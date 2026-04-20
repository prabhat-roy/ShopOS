package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/shopos/live-chat-service/internal/domain"
)

// Servicer defines the business-logic interface used by HTTP handlers.
type Servicer interface {
	StartSession(ctx context.Context, customerID string) (domain.ChatSession, error)
	AssignAgent(ctx context.Context, sessionID, agentID string) (domain.ChatSession, error)
	SendMessage(ctx context.Context, sessionID, senderID, senderType, body string) (domain.ChatMessage, error)
	GetSession(ctx context.Context, sessionID string) (domain.ChatSession, error)
	GetMessages(ctx context.Context, sessionID string) ([]domain.ChatMessage, error)
	CloseSession(ctx context.Context, sessionID string) (domain.ChatSession, error)
	ListWaitingSessions(ctx context.Context) ([]domain.ChatSession, error)
}

// Handler bundles the HTTP mux and service dependency.
type Handler struct {
	svc Servicer
	mux *http.ServeMux
}

// New builds and registers all routes.
func New(svc Servicer) *Handler {
	h := &Handler{svc: svc, mux: http.NewServeMux()}
	h.mux.HandleFunc("/healthz", h.healthz)
	h.mux.HandleFunc("/chat/sessions/waiting", h.listWaiting)
	h.mux.HandleFunc("/chat/sessions/", h.routeSession)
	h.mux.HandleFunc("/chat/sessions", h.handleSessions)
	return h
}

// ServeHTTP satisfies http.Handler.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func (h *Handler) healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handleSessions handles POST /chat/sessions.
func (h *Handler) handleSessions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req struct {
		CustomerID string `json:"customerId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.CustomerID == "" {
		writeError(w, http.StatusBadRequest, "customerId is required")
		return
	}
	session, err := h.svc.StartSession(r.Context(), req.CustomerID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, session)
}

// listWaiting handles GET /chat/sessions/waiting.
func (h *Handler) listWaiting(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	sessions, err := h.svc.ListWaitingSessions(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, sessions)
}

// routeSession dispatches sub-routes under /chat/sessions/{id}/...
func (h *Handler) routeSession(w http.ResponseWriter, r *http.Request) {
	// Strip /chat/sessions/ prefix
	path := strings.TrimPrefix(r.URL.Path, "/chat/sessions/")
	parts := strings.SplitN(path, "/", 2)
	sessionID := parts[0]
	if sessionID == "" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	sub := ""
	if len(parts) == 2 {
		sub = parts[1]
	}

	switch {
	case sub == "" && r.Method == http.MethodGet:
		h.getSession(w, r, sessionID)
	case sub == "assign" && r.Method == http.MethodPost:
		h.assignAgent(w, r, sessionID)
	case sub == "messages" && r.Method == http.MethodPost:
		h.sendMessage(w, r, sessionID)
	case sub == "messages" && r.Method == http.MethodGet:
		h.getMessages(w, r, sessionID)
	case sub == "close" && r.Method == http.MethodPost:
		h.closeSession(w, r, sessionID)
	default:
		writeError(w, http.StatusNotFound, "not found")
	}
}

func (h *Handler) getSession(w http.ResponseWriter, r *http.Request, id string) {
	session, err := h.svc.GetSession(r.Context(), id)
	if err == domain.ErrNotFound {
		writeError(w, http.StatusNotFound, "session not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, session)
}

func (h *Handler) assignAgent(w http.ResponseWriter, r *http.Request, id string) {
	var req struct {
		AgentID string `json:"agentId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.AgentID == "" {
		writeError(w, http.StatusBadRequest, "agentId is required")
		return
	}
	_, err := h.svc.AssignAgent(r.Context(), id, req.AgentID)
	if err == domain.ErrNotFound {
		writeError(w, http.StatusNotFound, "session not found")
		return
	}
	if err == domain.ErrSessionClosed {
		writeError(w, http.StatusConflict, "session is already closed")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) sendMessage(w http.ResponseWriter, r *http.Request, id string) {
	var req struct {
		SenderID   string `json:"senderId"`
		SenderType string `json:"senderType"`
		Body       string `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.SenderID == "" || req.Body == "" {
		writeError(w, http.StatusBadRequest, "senderId and body are required")
		return
	}
	if req.SenderType == "" {
		req.SenderType = domain.SenderCustomer
	}
	msg, err := h.svc.SendMessage(r.Context(), id, req.SenderID, req.SenderType, req.Body)
	if err == domain.ErrNotFound {
		writeError(w, http.StatusNotFound, "session not found")
		return
	}
	if err == domain.ErrSessionClosed {
		writeError(w, http.StatusConflict, "session is closed")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, msg)
}

func (h *Handler) getMessages(w http.ResponseWriter, r *http.Request, id string) {
	messages, err := h.svc.GetMessages(r.Context(), id)
	if err == domain.ErrNotFound {
		writeError(w, http.StatusNotFound, "session not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, messages)
}

func (h *Handler) closeSession(w http.ResponseWriter, r *http.Request, id string) {
	_, err := h.svc.CloseSession(r.Context(), id)
	if err == domain.ErrNotFound {
		writeError(w, http.StatusNotFound, "session not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
