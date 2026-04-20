package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/shopos/ab-testing-service/domain"
	"github.com/shopos/ab-testing-service/service"
)

// Handler holds dependencies for HTTP handlers.
type Handler struct {
	svc *service.ExperimentService
}

// New creates a Handler and registers routes on mux.
func New(svc *service.ExperimentService) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes wires all HTTP routes onto the provided mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/healthz", h.Healthz)
	mux.HandleFunc("/experiments", h.experimentsCollection)
	mux.HandleFunc("/experiments/", h.experimentsResource)
}

// Healthz returns service health.
func (h *Handler) Healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// experimentsCollection handles GET /experiments and POST /experiments.
func (h *Handler) experimentsCollection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listExperiments(w, r)
	case http.MethodPost:
		h.createExperiment(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// experimentsResource routes sub-paths under /experiments/{id}/...
func (h *Handler) experimentsResource(w http.ResponseWriter, r *http.Request) {
	// Strip leading "/experiments/"
	path := strings.TrimPrefix(r.URL.Path, "/experiments/")
	parts := strings.SplitN(path, "/", 2)
	id := parts[0]
	if id == "" {
		http.NotFound(w, r)
		return
	}

	if len(parts) == 1 {
		// /experiments/{id}
		if r.Method == http.MethodGet {
			h.getExperiment(w, r, id)
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	action := parts[1]
	switch {
	case action == "assign" && r.Method == http.MethodPost:
		h.assignVariant(w, r, id)
	case action == "convert" && r.Method == http.MethodPost:
		h.recordConversion(w, r, id)
	default:
		http.NotFound(w, r)
	}
}

func (h *Handler) createExperiment(w http.ResponseWriter, r *http.Request) {
	var exp domain.Experiment
	if err := json.NewDecoder(r.Body).Decode(&exp); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	created, err := h.svc.CreateExperiment(r.Context(), &exp)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *Handler) listExperiments(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.ListExperiments(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if list == nil {
		list = []*domain.Experiment{}
	}
	writeJSON(w, http.StatusOK, list)
}

func (h *Handler) getExperiment(w http.ResponseWriter, r *http.Request, id string) {
	exp, err := h.svc.GetExperiment(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, "experiment not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, exp)
}

func (h *Handler) assignVariant(w http.ResponseWriter, r *http.Request, experimentID string) {
	var body struct {
		UserID string `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if body.UserID == "" {
		writeError(w, http.StatusBadRequest, "user_id is required")
		return
	}
	assignment, err := h.svc.Assign(r.Context(), experimentID, body.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, "experiment not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"experiment_id": assignment.ExperimentID,
		"variant":       assignment.Variant,
	})
}

func (h *Handler) recordConversion(w http.ResponseWriter, r *http.Request, experimentID string) {
	var body struct {
		UserID string  `json:"user_id"`
		Metric string  `json:"metric"`
		Value  float64 `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if body.UserID == "" || body.Metric == "" {
		writeError(w, http.StatusBadRequest, "user_id and metric are required")
		return
	}
	if err := h.svc.RecordConversion(r.Context(), experimentID, body.UserID, body.Metric, body.Value); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// helpers ----------------------------------------------------------------

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}

func writeError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}
