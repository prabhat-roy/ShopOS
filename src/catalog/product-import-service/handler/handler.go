package handler

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/shopos/product-import-service/domain"
)

// Servicer is the interface the handler depends on.
type Servicer interface {
	CreateJob(fileName string, format domain.ImportFormat) (*domain.ImportJob, error)
	GetJob(id string) (*domain.ImportJob, error)
	ListJobs() ([]*domain.ImportJob, error)
	ProcessCSV(jobID string, data []byte) error
	ProcessJSON(jobID string, data []byte) error
}

// Handler holds HTTP handler methods and a reference to the service layer.
type Handler struct {
	svc Servicer
}

// New returns an initialised Handler.
func New(svc Servicer) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes wires all routes onto mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/healthz", h.Health)
	mux.HandleFunc("/imports", h.importsRouter)
	mux.HandleFunc("/imports/", h.importsWithIDRouter)
}

// Health responds with {"status":"ok"}.
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// importsRouter handles /imports (no trailing segment).
func (h *Handler) importsRouter(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.ListJobs(w, r)
	case http.MethodPost:
		h.CreateJob(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// importsWithIDRouter handles /imports/{id} and /imports/{id}/process.
func (h *Handler) importsWithIDRouter(w http.ResponseWriter, r *http.Request) {
	// Strip the "/imports/" prefix
	path := strings.TrimPrefix(r.URL.Path, "/imports/")
	parts := strings.SplitN(path, "/", 2)
	id := parts[0]
	if id == "" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	if len(parts) == 2 && parts[1] == "process" {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		h.ProcessJob(w, r, id)
		return
	}

	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	h.GetJob(w, r, id)
}

// ---- handlers ---------------------------------------------------------------

// CreateJob creates a new import job.
// POST /imports
// Body: { "file_name": "products.csv", "format": "CSV" }
func (h *Handler) CreateJob(w http.ResponseWriter, r *http.Request) {
	var req struct {
		FileName string             `json:"file_name"`
		Format   domain.ImportFormat `json:"format"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if strings.TrimSpace(req.FileName) == "" {
		writeError(w, http.StatusBadRequest, "file_name is required")
		return
	}
	if req.Format != domain.FormatCSV && req.Format != domain.FormatJSON {
		writeError(w, http.StatusBadRequest, "format must be CSV or JSON")
		return
	}

	job, err := h.svc.CreateJob(req.FileName, req.Format)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create import job")
		return
	}
	writeJSON(w, http.StatusCreated, job)
}

// ListJobs returns all import jobs.
// GET /imports
func (h *Handler) ListJobs(w http.ResponseWriter, r *http.Request) {
	jobs, err := h.svc.ListJobs()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list import jobs")
		return
	}
	if jobs == nil {
		jobs = []*domain.ImportJob{}
	}
	writeJSON(w, http.StatusOK, jobs)
}

// GetJob returns a single import job with its error list.
// GET /imports/{id}
func (h *Handler) GetJob(w http.ResponseWriter, r *http.Request, id string) {
	job, err := h.svc.GetJob(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "import job not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get import job")
		return
	}
	writeJSON(w, http.StatusOK, job)
}

// ProcessJob submits file content to be processed asynchronously.
// POST /imports/{id}/process
// Body: { "format": "CSV", "data_base64": "<base64-encoded file content>" }
func (h *Handler) ProcessJob(w http.ResponseWriter, r *http.Request, id string) {
	var req struct {
		Format     domain.ImportFormat `json:"format"`
		DataBase64 string             `json:"data_base64"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.DataBase64 == "" {
		writeError(w, http.StatusBadRequest, "data_base64 is required")
		return
	}

	data, err := base64.StdEncoding.DecodeString(req.DataBase64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "data_base64 is not valid base64")
		return
	}

	// Verify the job exists before dispatching processing.
	if _, err := h.svc.GetJob(id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "import job not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to find import job")
		return
	}

	// Process in the background so the caller receives 202 immediately.
	go func() {
		switch req.Format {
		case domain.FormatCSV:
			_ = h.svc.ProcessCSV(id, data)
		case domain.FormatJSON:
			_ = h.svc.ProcessJSON(id, data)
		}
	}()

	writeJSON(w, http.StatusAccepted, map[string]string{
		"message": "import processing started",
		"job_id":  id,
	})
}

// ---- helpers ----------------------------------------------------------------

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
