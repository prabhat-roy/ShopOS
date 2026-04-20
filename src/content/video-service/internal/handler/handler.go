package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/shopos/video-service/internal/domain"
	"github.com/shopos/video-service/internal/service"
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
	h.mux.HandleFunc("/videos", h.videosRoot)
	h.mux.HandleFunc("/videos/", h.videosWithID)
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

// videoID extracts the first path segment after the given prefix.
func videoID(path, prefix string) string {
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

// videosRoot handles /videos (no trailing ID).
// POST /videos — multipart upload
// GET  /videos?ownerId=xxx — list videos by owner
func (h *Handler) videosRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/videos" && r.URL.Path != "/videos/" {
		http.NotFound(w, r)
		return
	}
	switch r.Method {
	case http.MethodPost:
		h.uploadVideo(w, r)
	case http.MethodGet:
		h.listVideos(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// videosWithID dispatches sub-routes: /videos/{id}, /videos/{id}/stream, /videos/{id}/status.
func (h *Handler) videosWithID(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	switch {
	case strings.HasSuffix(path, "/stream"):
		id := videoID(strings.TrimSuffix(path, "/stream"), "/videos/")
		if id == "" {
			writeError(w, http.StatusBadRequest, "missing video id")
			return
		}
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		h.streamVideo(w, r, id)

	case strings.HasSuffix(path, "/status"):
		id := videoID(strings.TrimSuffix(path, "/status"), "/videos/")
		if id == "" {
			writeError(w, http.StatusBadRequest, "missing video id")
			return
		}
		if r.Method != http.MethodPatch {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		h.updateStatus(w, r, id)

	default:
		id := videoID(path, "/videos/")
		if id == "" {
			writeError(w, http.StatusBadRequest, "missing video id")
			return
		}
		switch r.Method {
		case http.MethodGet:
			h.getVideo(w, r, id)
		case http.MethodDelete:
			h.deleteVideo(w, r, id)
		default:
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	}
}

// uploadVideo handles POST /videos.
// Multipart form fields: owner_id, title, description (optional).
// File field: file.
func (h *Handler) uploadVideo(w http.ResponseWriter, r *http.Request) {
	const maxMemory = 64 << 20 // 64 MB in-memory buffer
	if err := r.ParseMultipartForm(maxMemory); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("parse multipart form: %s", err))
		return
	}

	ownerID := r.FormValue("owner_id")
	if ownerID == "" {
		writeError(w, http.StatusBadRequest, "owner_id is required")
		return
	}

	title := r.FormValue("title")
	if title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}

	description := r.FormValue("description")

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("file field: %s", err))
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "video/mp4"
	}

	video, err := h.svc.UploadVideo(r.Context(), ownerID, title, description, header.Filename, contentType, file, header.Size)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("upload: %s", err))
		return
	}

	writeJSON(w, http.StatusCreated, video)
}

// listVideos handles GET /videos?ownerId=xxx.
func (h *Handler) listVideos(w http.ResponseWriter, r *http.Request) {
	ownerID := r.URL.Query().Get("ownerId")
	if ownerID == "" {
		writeError(w, http.StatusBadRequest, "ownerId query parameter is required")
		return
	}

	videos, err := h.svc.ListOwnerVideos(r.Context(), ownerID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, videos)
}

// getVideo handles GET /videos/{id}.
func (h *Handler) getVideo(w http.ResponseWriter, r *http.Request, id string) {
	video, err := h.svc.GetVideo(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, "video not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, video)
}

// streamVideo handles GET /videos/{id}/stream — returns the presigned URL as JSON.
func (h *Handler) streamVideo(w http.ResponseWriter, r *http.Request, id string) {
	url, err := h.svc.GetStreamURL(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, "video not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"streamUrl": url})
}

// updateStatus handles PATCH /videos/{id}/status.
// JSON body: {"status":"READY"}
func (h *Handler) updateStatus(w http.ResponseWriter, r *http.Request, id string) {
	var body struct {
		Status domain.VideoStatus `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("decode body: %s", err))
		return
	}

	switch body.Status {
	case domain.VideoStatusUploading, domain.VideoStatusProcessing, domain.VideoStatusReady, domain.VideoStatusFailed:
		// valid
	default:
		writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid status: %q", body.Status))
		return
	}

	if err := h.svc.UpdateVideoStatus(r.Context(), id, body.Status); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, "video not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// deleteVideo handles DELETE /videos/{id}.
func (h *Handler) deleteVideo(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.svc.DeleteVideo(r.Context(), id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, "video not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
