// Package handler provides the HTTP API for push-notification-service.
package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/shopos/push-notification-service/internal/domain"
)

// Storer is the subset of store.Store used by the handler.
// Defined as an interface so tests can inject a mock.
type Storer interface {
	Get(messageID string) (domain.PushRecord, bool)
	List(limit int) []domain.PushRecord
	Stats() domain.PushStats
}

// ConsumerStatus is implemented by the Kafka consumer to report liveness.
type ConsumerStatus interface {
	IsRunning() bool
}

// Handler holds dependencies and registers HTTP routes.
type Handler struct {
	store    Storer
	consumer ConsumerStatus
	mux      *http.ServeMux
}

// New wires up the HTTP handler and registers all routes.
func New(st Storer, cs ConsumerStatus) *Handler {
	h := &Handler{
		store:    st,
		consumer: cs,
		mux:      http.NewServeMux(),
	}
	h.routes()
	return h
}

// ServeHTTP implements http.Handler.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

// -------------------------------------------------------------------
// Route registration
// -------------------------------------------------------------------

func (h *Handler) routes() {
	h.mux.HandleFunc("/healthz", h.handleHealthz)
	h.mux.HandleFunc("/push/stats", h.handleStats) // must be before /push/{id}
	h.mux.HandleFunc("/push/", h.handlePushByID)   // /push/{messageId}
	h.mux.HandleFunc("/push", h.handleListPush)
}

// -------------------------------------------------------------------
// Handlers
// -------------------------------------------------------------------

// GET /healthz
func (h *Handler) handleHealthz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	consumerStatus := "stopped"
	if h.consumer.IsRunning() {
		consumerStatus = "running"
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"status":   "ok",
		"consumer": consumerStatus,
	})
}

// GET /push/stats
func (h *Handler) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	writeJSON(w, http.StatusOK, h.store.Stats())
}

// GET /push/{messageId}
func (h *Handler) handlePushByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Strip the "/push/" prefix
	messageID := strings.TrimPrefix(r.URL.Path, "/push/")
	messageID = strings.TrimSpace(messageID)
	if messageID == "" {
		writeError(w, http.StatusBadRequest, "messageId path parameter is required")
		return
	}

	record, ok := h.store.Get(messageID)
	if !ok {
		writeError(w, http.StatusNotFound, "push record not found: "+messageID)
		return
	}
	writeJSON(w, http.StatusOK, record)
}

// GET /push?limit=N
func (h *Handler) handleListPush(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	limit := 50
	if q := r.URL.Query().Get("limit"); q != "" {
		n, err := strconv.Atoi(q)
		if err != nil || n < 1 || n > 500 {
			writeError(w, http.StatusBadRequest, "limit must be an integer between 1 and 500")
			return
		}
		limit = n
	}

	records := h.store.List(limit)
	if records == nil {
		records = []domain.PushRecord{}
	}
	writeJSON(w, http.StatusOK, records)
}

// -------------------------------------------------------------------
// Helpers
// -------------------------------------------------------------------

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("[handler] JSON encode error: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
