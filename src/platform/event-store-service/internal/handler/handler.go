package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/shopos/event-store-service/internal/domain"
)

// Servicer is the interface the handlers depend on.
type Servicer interface {
	Append(ctx context.Context, req *domain.AppendRequest) ([]*domain.Event, error)
	ReadStream(ctx context.Context, req domain.ReadRequest) ([]*domain.Event, error)
	ReadAll(ctx context.Context, req domain.ReadAllRequest) ([]*domain.Event, error)
	GetEvent(ctx context.Context, id string) (*domain.Event, error)
	SaveSnapshot(ctx context.Context, streamID, streamType string, version int64, state []byte) error
	GetSnapshot(ctx context.Context, streamID string) (*domain.Snapshot, error)
}

// Handler holds all HTTP handlers.
type Handler struct{ svc Servicer }

func New(svc Servicer) *Handler { return &Handler{svc: svc} }

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /healthz", h.health)

	// Events
	mux.HandleFunc("POST /streams/{streamID}/events", h.appendEvents)
	mux.HandleFunc("GET /streams/{streamID}/events", h.readStream)
	mux.HandleFunc("GET /events/{id}", h.getEvent)
	mux.HandleFunc("GET /events", h.readAll)

	// Snapshots
	mux.HandleFunc("PUT /streams/{streamID}/snapshot", h.saveSnapshot)
	mux.HandleFunc("GET /streams/{streamID}/snapshot", h.getSnapshot)
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) appendEvents(w http.ResponseWriter, r *http.Request) {
	var req domain.AppendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errMsg("invalid request body"))
		return
	}
	req.StreamID = r.PathValue("streamID")

	events, err := h.svc.Append(r.Context(), &req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"events": events, "count": len(events)})
}

func (h *Handler) readStream(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	req := domain.ReadRequest{
		StreamID:    r.PathValue("streamID"),
		FromVersion: queryInt64(q.Get("from_version"), 0),
		ToVersion:   queryInt64(q.Get("to_version"), 0),
		MaxCount:    queryInt(q.Get("max_count"), 0),
	}
	events, err := h.svc.ReadStream(r.Context(), req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"events": events, "count": len(events)})
}

func (h *Handler) getEvent(w http.ResponseWriter, r *http.Request) {
	ev, err := h.svc.GetEvent(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, ev)
}

func (h *Handler) readAll(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	req := domain.ReadAllRequest{
		FromGlobalSeq: queryInt64(q.Get("from_seq"), 0),
		MaxCount:      queryInt(q.Get("max_count"), 0),
		StreamType:    q.Get("stream_type"),
		EventType:     q.Get("event_type"),
	}
	events, err := h.svc.ReadAll(r.Context(), req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"events": events, "count": len(events)})
}

func (h *Handler) saveSnapshot(w http.ResponseWriter, r *http.Request) {
	var body struct {
		StreamType string `json:"stream_type"`
		Version    int64  `json:"version"`
		State      []byte `json:"state"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, errMsg("invalid request body"))
		return
	}
	err := h.svc.SaveSnapshot(r.Context(), r.PathValue("streamID"), body.StreamType, body.Version, body.State)
	if err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) getSnapshot(w http.ResponseWriter, r *http.Request) {
	snap, err := h.svc.GetSnapshot(r.Context(), r.PathValue("streamID"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, snap)
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		writeJSON(w, http.StatusNotFound, errMsg(err.Error()))
	case errors.Is(err, domain.ErrInvalidInput):
		writeJSON(w, http.StatusBadRequest, errMsg(err.Error()))
	case errors.Is(err, domain.ErrVersionConflict):
		writeJSON(w, http.StatusConflict, errMsg(err.Error()))
	default:
		writeJSON(w, http.StatusInternalServerError, errMsg("internal server error"))
	}
}

func errMsg(msg string) map[string]string { return map[string]string{"error": msg} }

func queryInt(s string, fallback int) int {
	if n, err := strconv.Atoi(s); err == nil && n > 0 {
		return n
	}
	return fallback
}

func queryInt64(s string, fallback int64) int64 {
	if n, err := strconv.ParseInt(s, 10, 64); err == nil && n >= 0 {
		return n
	}
	return fallback
}
