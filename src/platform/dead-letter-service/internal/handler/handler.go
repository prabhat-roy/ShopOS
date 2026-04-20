package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/shopos/dead-letter-service/internal/domain"
)

// Servicer is the business-logic interface required by Handler.
type Servicer interface {
	Save(msg *domain.DeadMessage) error
	Get(id string) (*domain.DeadMessage, error)
	List(topic string, status domain.MessageStatus, limit, offset int) ([]*domain.DeadMessage, error)
	Retry(id string) error
	Discard(id string) error
	Stats() (map[string]int64, error)
}

// Handler wires all HTTP routes and holds a reference to the service layer.
type Handler struct {
	svc    Servicer
	logger *slog.Logger
}

// New creates a Handler and registers all routes on mux.
func New(svc Servicer, logger *slog.Logger) *Handler {
	return &Handler{svc: svc, logger: logger}
}

// RegisterRoutes mounts all routes onto the provided ServeMux using the
// Go 1.22+ METHOD /path pattern.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /healthz", h.healthz)
	mux.HandleFunc("GET /messages", h.listMessages)
	mux.HandleFunc("GET /messages/{id}", h.getMessage)
	mux.HandleFunc("POST /messages/{id}/retry", h.retryMessage)
	mux.HandleFunc("POST /messages/{id}/discard", h.discardMessage)
	mux.HandleFunc("GET /stats", h.stats)
}

// healthz returns a simple liveness response.
func (h *Handler) healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// listMessages handles GET /messages with optional query params:
//
//	topic, status, limit, offset
func (h *Handler) listMessages(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	topic := q.Get("topic")
	statusStr := q.Get("status")
	limit := queryInt(q.Get("limit"), 50)
	offset := queryInt(q.Get("offset"), 0)

	var status domain.MessageStatus
	if statusStr != "" {
		status = domain.MessageStatus(statusStr)
	}

	msgs, err := h.svc.List(topic, status, limit, offset)
	if err != nil {
		h.logger.Error("list messages", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to list messages")
		return
	}
	// Return an empty array rather than null.
	if msgs == nil {
		msgs = []*domain.DeadMessage{}
	}
	writeJSON(w, http.StatusOK, msgs)
}

// getMessage handles GET /messages/{id}.
func (h *Handler) getMessage(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	msg, err := h.svc.Get(id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, "message not found")
			return
		}
		h.logger.Error("get message", "id", id, "error", err)
		writeError(w, http.StatusInternalServerError, "failed to get message")
		return
	}
	writeJSON(w, http.StatusOK, msg)
}

// retryMessage handles POST /messages/{id}/retry.
func (h *Handler) retryMessage(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.svc.Retry(id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, "message not found")
			return
		}
		h.logger.Error("retry message", "id", id, "error", err)
		writeError(w, http.StatusInternalServerError, "failed to retry message")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// discardMessage handles POST /messages/{id}/discard.
func (h *Handler) discardMessage(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.svc.Discard(id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, "message not found")
			return
		}
		h.logger.Error("discard message", "id", id, "error", err)
		writeError(w, http.StatusInternalServerError, "failed to discard message")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// stats handles GET /stats.
func (h *Handler) stats(w http.ResponseWriter, r *http.Request) {
	counts, err := h.svc.Stats()
	if err != nil {
		h.logger.Error("stats", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to fetch stats")
		return
	}
	writeJSON(w, http.StatusOK, counts)
}

// writeJSON encodes v as JSON and writes it with the given status code.
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		// At this point the status code is already sent; just log.
		_ = err
	}
}

// writeError writes a standard JSON error response.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// queryInt parses a query parameter as int, returning defaultVal on failure.
func queryInt(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 0 {
		return defaultVal
	}
	return n
}
