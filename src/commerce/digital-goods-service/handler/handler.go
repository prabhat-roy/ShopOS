package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/shopos/digital-goods-service/domain"
	"github.com/shopos/digital-goods-service/service"
)

// Handler holds HTTP route logic for the digital-goods-service.
type Handler struct {
	svc *service.AssetService
}

// New creates a Handler.
func New(svc *service.AssetService) *Handler { return &Handler{svc: svc} }

// RegisterRoutes wires all routes onto mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/healthz", h.Healthz)
	mux.HandleFunc("/assets", h.assetsCollection)
	mux.HandleFunc("/assets/", h.assetsResource)
}

func (h *Handler) Healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// assetsCollection handles POST /assets (upload).
func (h *Handler) assetsCollection(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	h.uploadAsset(w, r)
}

// assetsResource routes sub-paths under /assets/...
func (h *Handler) assetsResource(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/assets/")
	parts := strings.SplitN(path, "/", 3)

	// /assets/product/{productID}
	if parts[0] == "product" && len(parts) == 2 {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		h.listByProduct(w, r, parts[1])
		return
	}

	id := parts[0]
	if id == "" {
		http.NotFound(w, r)
		return
	}

	// /assets/{id}
	if len(parts) == 1 {
		switch r.Method {
		case http.MethodGet:
			h.getAsset(w, r, id)
		case http.MethodDelete:
			h.deleteAsset(w, r, id)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	// /assets/{id}/download
	if parts[1] == "download" && r.Method == http.MethodGet {
		h.generateDownload(w, r, id)
		return
	}

	http.NotFound(w, r)
}

// uploadAsset handles multipart file upload: POST /assets
func (h *Handler) uploadAsset(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "failed to parse multipart form: "+err.Error())
		return
	}

	productID := r.FormValue("product_id")
	if productID == "" {
		writeError(w, http.StatusBadRequest, "product_id is required")
		return
	}
	name := r.FormValue("name")
	if name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	downloadLimit := 0
	if dl := r.FormValue("download_limit"); dl != "" {
		var err error
		downloadLimit, err = strconv.Atoi(dl)
		if err != nil {
			writeError(w, http.StatusBadRequest, "download_limit must be an integer")
			return
		}
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "file field is required: "+err.Error())
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Read into buffer to determine size (MinIO needs size for non-chunked uploads)
	data, err := io.ReadAll(file)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read file: "+err.Error())
		return
	}

	asset, err := h.svc.UploadAsset(
		r.Context(),
		productID,
		name,
		header.Filename,
		contentType,
		strings.NewReader(string(data)),
		int64(len(data)),
		downloadLimit,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "upload failed: "+err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, asset)
}

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

func (h *Handler) listByProduct(w http.ResponseWriter, r *http.Request, productID string) {
	assets, err := h.svc.ListByProduct(r.Context(), productID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if assets == nil {
		assets = []*domain.DigitalAsset{}
	}
	writeJSON(w, http.StatusOK, assets)
}

func (h *Handler) generateDownload(w http.ResponseWriter, r *http.Request, id string) {
	orderID := r.URL.Query().Get("order_id")
	if orderID == "" {
		writeError(w, http.StatusBadRequest, "order_id query parameter is required")
		return
	}
	link, err := h.svc.GenerateDownloadLink(r.Context(), id, orderID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			writeError(w, http.StatusNotFound, "asset not found")
		case errors.Is(err, domain.ErrDownloadLimitExceeded):
			writeError(w, http.StatusForbidden, fmt.Sprintf("download limit exceeded for asset %s", id))
		default:
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	writeJSON(w, http.StatusOK, link)
}

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

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}

func writeError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}
