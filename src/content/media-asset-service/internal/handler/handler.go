package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/shopos/media-asset-service/internal/domain"
	"github.com/shopos/media-asset-service/internal/service"
)

// Handler holds the HTTP mux and its dependencies.
type Handler struct {
	svc service.Servicer
	mux *http.ServeMux
}

// New wires up all routes and returns an http.Handler.
func New(svc service.Servicer) http.Handler {
	h := &Handler{svc: svc, mux: http.NewServeMux()}
	h.mux.HandleFunc("/healthz", h.healthz)
	h.mux.HandleFunc("/assets", h.assetsRoot)
	h.mux.HandleFunc("/assets/", h.assetsWithID)
	return h.mux
}

// ServeHTTP satisfies http.Handler.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

// ---- helpers ----------------------------------------------------------------

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}

func assetID(path, prefix string) string {
	// path: /assets/{id} or /assets/{id}/download
	trimmed := strings.TrimPrefix(path, prefix)
	parts := strings.SplitN(trimmed, "/", 2)
	if len(parts) == 0 {
		return ""
	}
	return parts[0]
}

// ---- route handlers ---------------------------------------------------------

func (h *Handler) healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// assetsRoot handles POST /assets (no trailing ID).
func (h *Handler) assetsRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/assets" && r.URL.Path != "/assets/" {
		http.NotFound(w, r)
		return
	}
	switch r.Method {
	case http.MethodPost:
		h.uploadAsset(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// assetsWithID dispatches /assets/{id} and /assets/{id}/download.
func (h *Handler) assetsWithID(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if strings.HasSuffix(path, "/download") {
		id := assetID(strings.TrimSuffix(path, "/download"), "/assets/")
		if id == "" {
			writeError(w, http.StatusBadRequest, "missing asset id")
			return
		}
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		h.downloadAsset(w, r, id)
		return
	}

	id := assetID(path, "/assets/")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing asset id")
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getAsset(w, r, id)
	case http.MethodDelete:
		h.deleteAsset(w, r, id)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// uploadAsset handles POST /assets as a multipart/form-data upload.
// Form fields: owner_id (string), tags (comma-separated, optional).
// File field: file.
func (h *Handler) uploadAsset(w http.ResponseWriter, r *http.Request) {
	const maxMemory = 32 << 20 // 32 MB in-memory buffer
	if err := r.ParseMultipartForm(maxMemory); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("parse multipart form: %s", err))
		return
	}

	ownerID := r.FormValue("owner_id")
	if ownerID == "" {
		writeError(w, http.StatusBadRequest, "owner_id is required")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("file field: %s", err))
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	asset, err := h.svc.UploadAsset(r.Context(), ownerID, header.Filename, contentType, file, header.Size)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("upload: %s", err))
		return
	}

	writeJSON(w, http.StatusCreated, asset)
}

// getAsset handles GET /assets/{id} — returns asset metadata as JSON.
func (h *Handler) getAsset(w http.ResponseWriter, r *http.Request, id string) {
	asset, err := h.svc.GetAsset(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, "asset not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, asset)
}

// downloadAsset handles GET /assets/{id}/download — redirects to a presigned URL.
func (h *Handler) downloadAsset(w http.ResponseWriter, r *http.Request, id string) {
	url, err := h.svc.GetDownloadURL(r.Context(), id, time.Hour)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, "asset not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	http.Redirect(w, r, url, http.StatusFound)
}

// deleteAsset handles DELETE /assets/{id} — returns 204 on success.
func (h *Handler) deleteAsset(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.svc.DeleteAsset(r.Context(), id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, "asset not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

