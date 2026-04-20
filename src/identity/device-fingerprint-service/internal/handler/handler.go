package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/shopos/device-fingerprint-service/internal/domain"
	"github.com/shopos/device-fingerprint-service/internal/service"
)

// Handler holds the HTTP handler state and routes.
type Handler struct {
	svc    service.Servicer
	logger *slog.Logger
	mux    *http.ServeMux
}

// New constructs a Handler and registers all routes onto an internal ServeMux.
func New(svc service.Servicer, logger *slog.Logger) *Handler {
	h := &Handler{
		svc:    svc,
		logger: logger,
		mux:    http.NewServeMux(),
	}
	h.routes()
	return h
}

// ServeHTTP satisfies http.Handler so the Handler can be passed directly to
// http.ListenAndServe.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

// routes registers all HTTP routes.
func (h *Handler) routes() {
	h.mux.HandleFunc("GET /healthz", h.handleHealthz)
	h.mux.HandleFunc("POST /fingerprints/identify", h.handleIdentify)

	// GET /fingerprints/{id} and GET /fingerprints/user/{userID} share the same
	// path prefix, so we use a single catch-all and dispatch manually.
	h.mux.HandleFunc("/fingerprints/", h.handleFingerprintRouter)
}

// handleHealthz returns a simple liveness response.
func (h *Handler) handleHealthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handleIdentify identifies a device and returns its fingerprint information.
//
// POST /fingerprints/identify
func (h *Handler) handleIdentify(w http.ResponseWriter, r *http.Request) {
	var req domain.FingerprintRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	if req.Attributes.UserAgent == "" {
		writeError(w, http.StatusBadRequest, "attributes.user_agent is required")
		return
	}

	resp, err := h.svc.Identify(r.Context(), &req)
	if err != nil {
		h.logger.Error("identify error", "error", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// handleFingerprintRouter dispatches GET /fingerprints/{id} and
// GET /fingerprints/user/{userID}.
func (h *Handler) handleFingerprintRouter(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Strip leading /fingerprints/ to get the remainder.
	remainder := strings.TrimPrefix(r.URL.Path, "/fingerprints/")
	remainder = strings.TrimSuffix(remainder, "/")

	if strings.HasPrefix(remainder, "user/") {
		// GET /fingerprints/user/{userID}
		userID := strings.TrimPrefix(remainder, "user/")
		if userID == "" {
			writeError(w, http.StatusBadRequest, "userID is required")
			return
		}
		h.handleGetUserFingerprints(w, r, userID)
		return
	}

	// GET /fingerprints/{id}
	id := remainder
	if id == "" {
		writeError(w, http.StatusBadRequest, "fingerprint id is required")
		return
	}
	h.handleGetByID(w, r, id)
}

// handleGetByID returns the full Fingerprint record for the given UUID.
//
// GET /fingerprints/{id}
func (h *Handler) handleGetByID(w http.ResponseWriter, r *http.Request, id string) {
	fp, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, "fingerprint not found")
			return
		}
		h.logger.Error("get by id error", "id", id, "error", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, fp)
}

// handleGetUserFingerprints returns all known devices for a user.
//
// GET /fingerprints/user/{userID}
func (h *Handler) handleGetUserFingerprints(w http.ResponseWriter, r *http.Request, userID string) {
	fps, err := h.svc.GetUserFingerprints(r.Context(), userID)
	if err != nil {
		h.logger.Error("get user fingerprints error", "userID", userID, "error", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, fps)
}

// ---- helpers ----------------------------------------------------------------

type errorResponse struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Default().Error("writeJSON encode error", "error", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorResponse{Error: msg})
}
