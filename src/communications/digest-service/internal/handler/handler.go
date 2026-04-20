package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopos/digest-service/internal/domain"
	"github.com/shopos/digest-service/internal/scheduler"
	"github.com/shopos/digest-service/internal/store"
)

// Handler holds the HTTP handler dependencies.
type Handler struct {
	store store.Storer
}

// New creates a new Handler.
func New(s store.Storer) *Handler {
	return &Handler{store: s}
}

// RegisterRoutes wires all routes to the given ServeMux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/healthz", h.healthz)
	mux.HandleFunc("/digests", h.routeDigests)
	mux.HandleFunc("/digests/", h.routeDigestsWithID)
}

// healthz returns a simple liveness response.
func (h *Handler) healthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

// routeDigests handles /digests (no trailing path).
//
//	POST /digests   → create config
//	GET  /digests   → list configs (optional ?status= &frequency=)
func (h *Handler) routeDigests(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.createConfig(w, r)
	case http.MethodGet:
		h.listConfigs(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// routeDigestsWithID dispatches sub-routes under /digests/{segment...}.
func (h *Handler) routeDigestsWithID(w http.ResponseWriter, r *http.Request) {
	// Strip "/digests/" prefix.
	path := strings.TrimPrefix(r.URL.Path, "/digests/")
	path = strings.TrimSuffix(path, "/")
	segments := strings.Split(path, "/")

	switch len(segments) {
	case 1:
		// /digests/{id}
		h.handleSingleConfig(w, r, segments[0])

	case 2:
		first := segments[0]
		second := segments[1]

		if first == "user" {
			// /digests/user/{userId}
			h.getByUser(w, r, second)
			return
		}
		// /digests/{id}/pause | resume | runs
		switch second {
		case "pause":
			if r.Method != http.MethodPost {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			h.pauseConfig(w, r, first)
		case "resume":
			if r.Method != http.MethodPost {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			h.resumeConfig(w, r, first)
		case "runs":
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			h.listRuns(w, r, first)
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}

	default:
		http.Error(w, "not found", http.StatusNotFound)
	}
}

// handleSingleConfig handles GET and DELETE on /digests/{id}.
func (h *Handler) handleSingleConfig(w http.ResponseWriter, r *http.Request, rawID string) {
	switch r.Method {
	case http.MethodGet:
		h.getConfig(w, r, rawID)
	case http.MethodDelete:
		h.deleteConfig(w, r, rawID)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// createConfig handles POST /digests.
func (h *Handler) createConfig(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if msg := req.Validate(); msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}

	userID, _ := uuid.Parse(req.UserID)
	now := time.Now().UTC()
	cfg := domain.DigestConfig{
		ID:         uuid.New(),
		UserID:     userID,
		Email:      req.Email,
		Frequency:  req.Frequency,
		Status:     domain.StatusActive,
		NextSendAt: scheduler.ComputeNextSend(req.Frequency, now),
		Timezone:   req.Timezone,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := h.store.CreateConfig(r.Context(), cfg); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, cfg)
}

// getConfig handles GET /digests/{id}.
func (h *Handler) getConfig(w http.ResponseWriter, r *http.Request, rawID string) {
	id, err := uuid.Parse(rawID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	cfg, err := h.store.GetConfig(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || strings.Contains(err.Error(), "no rows") {
			writeError(w, http.StatusNotFound, "config not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, cfg)
}

// getByUser handles GET /digests/user/{userId}.
func (h *Handler) getByUser(w http.ResponseWriter, r *http.Request, rawUserID string) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID, err := uuid.Parse(rawUserID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid userId")
		return
	}
	cfgs, err := h.store.GetByUserID(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, cfgs)
}

// listConfigs handles GET /digests?status=&frequency=.
func (h *Handler) listConfigs(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	status := domain.DigestStatus(q.Get("status"))
	frequency := domain.DigestFrequency(q.Get("frequency"))
	cfgs, err := h.store.ListConfigs(r.Context(), status, frequency)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, cfgs)
}

// pauseConfig handles POST /digests/{id}/pause.
func (h *Handler) pauseConfig(w http.ResponseWriter, r *http.Request, rawID string) {
	id, err := uuid.Parse(rawID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.store.PauseConfig(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// resumeConfig handles POST /digests/{id}/resume.
func (h *Handler) resumeConfig(w http.ResponseWriter, r *http.Request, rawID string) {
	id, err := uuid.Parse(rawID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.store.ResumeConfig(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// deleteConfig handles DELETE /digests/{id}.
func (h *Handler) deleteConfig(w http.ResponseWriter, r *http.Request, rawID string) {
	id, err := uuid.Parse(rawID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.store.DeleteConfig(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// listRuns handles GET /digests/{id}/runs.
func (h *Handler) listRuns(w http.ResponseWriter, r *http.Request, rawID string) {
	id, err := uuid.Parse(rawID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	runs, err := h.store.ListRuns(r.Context(), id, 20)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, runs)
}

// writeJSON serialises v as JSON with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError writes a JSON error body.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
