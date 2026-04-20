package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/shopos/api-key-service/internal/domain"
)

// Servicer is the interface the handler depends on. The service package satisfies it;
// tests inject a mock.
type Servicer interface {
	Create(ctx context.Context, req *domain.CreateKeyRequest) (*domain.APIKey, string, error)
	Validate(ctx context.Context, rawKey string) (*domain.ValidateResponse, error)
	List(ctx context.Context, ownerID string) ([]*domain.APIKey, error)
	GetByID(ctx context.Context, id string) (*domain.APIKey, error)
	Deactivate(ctx context.Context, id string) error
	Delete(ctx context.Context, id string) error
}

// Handler groups all HTTP handler methods.
type Handler struct {
	svc Servicer
}

// New creates a Handler and registers all routes on mux.
func New(svc Servicer) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes attaches every route to the provided ServeMux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/healthz", h.healthz)
	mux.HandleFunc("/keys/validate", h.validateKey)  // must come before /keys/{id}
	mux.HandleFunc("/keys/", h.keysWithID)            // /keys/{id}
	mux.HandleFunc("/keys", h.keysRoot)               // /keys  (no trailing slash)
}

// ---- route handlers ---------------------------------------------------------

func (h *Handler) healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// keysRoot handles:
//
//	POST /keys  → create
//	GET  /keys  → list (query: owner_id)
func (h *Handler) keysRoot(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.createKey(w, r)
	case http.MethodGet:
		h.listKeys(w, r)
	default:
		methodNotAllowed(w)
	}
}

// keysWithID handles:
//
//	GET    /keys/{id} → get by ID
//	DELETE /keys/{id} → deactivate
func (h *Handler) keysWithID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/keys/")
	if id == "" {
		http.NotFound(w, r)
		return
	}
	switch r.Method {
	case http.MethodGet:
		h.getKey(w, r, id)
	case http.MethodDelete:
		h.deactivateKey(w, r, id)
	default:
		methodNotAllowed(w)
	}
}

// POST /keys
func (h *Handler) createKey(w http.ResponseWriter, r *http.Request) {
	var body struct {
		OwnerID   string    `json:"owner_id"`
		OwnerType string    `json:"owner_type"`
		Name      string    `json:"name"`
		Scopes    []string  `json:"scopes"`
		ExpiresAt *time.Time `json:"expires_at"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeBadRequest(w, "invalid JSON: "+err.Error())
		return
	}

	req := &domain.CreateKeyRequest{
		OwnerID:   body.OwnerID,
		OwnerType: body.OwnerType,
		Name:      body.Name,
		Scopes:    body.Scopes,
		ExpiresAt: body.ExpiresAt,
	}

	key, rawKey, err := h.svc.Create(r.Context(), req)
	if err != nil {
		log.Printf("create key error: %v", err)
		writeBadRequest(w, err.Error())
		return
	}

	// Return the raw key exactly once in the response together with key metadata.
	writeJSON(w, http.StatusCreated, map[string]any{
		"id":          key.ID,
		"owner_id":    key.OwnerID,
		"owner_type":  key.OwnerType,
		"name":        key.Name,
		"key_prefix":  key.KeyPrefix,
		"key":         rawKey, // shown once, never stored
		"scopes":      key.Scopes,
		"active":      key.Active,
		"expires_at":  key.ExpiresAt,
		"created_at":  key.CreatedAt,
		"updated_at":  key.UpdatedAt,
	})
}

// GET /keys?owner_id=xxx
func (h *Handler) listKeys(w http.ResponseWriter, r *http.Request) {
	ownerID := r.URL.Query().Get("owner_id")
	if ownerID == "" {
		writeBadRequest(w, "owner_id query parameter is required")
		return
	}

	keys, err := h.svc.List(r.Context(), ownerID)
	if err != nil {
		log.Printf("list keys error: %v", err)
		writeInternalError(w)
		return
	}
	if keys == nil {
		keys = []*domain.APIKey{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"keys": keys})
}

// GET /keys/{id}
func (h *Handler) getKey(w http.ResponseWriter, r *http.Request, id string) {
	key, err := h.svc.GetByID(r.Context(), id)
	if errors.Is(err, domain.ErrNotFound) {
		writeNotFound(w)
		return
	}
	if err != nil {
		log.Printf("get key error: %v", err)
		writeInternalError(w)
		return
	}
	writeJSON(w, http.StatusOK, key)
}

// DELETE /keys/{id}  → deactivates (soft delete)
func (h *Handler) deactivateKey(w http.ResponseWriter, r *http.Request, id string) {
	err := h.svc.Deactivate(r.Context(), id)
	if errors.Is(err, domain.ErrNotFound) {
		writeNotFound(w)
		return
	}
	if err != nil {
		log.Printf("deactivate key error: %v", err)
		writeInternalError(w)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// POST /keys/validate
func (h *Handler) validateKey(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}

	var body struct {
		Key string `json:"key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeBadRequest(w, "invalid JSON: "+err.Error())
		return
	}

	resp, err := h.svc.Validate(r.Context(), body.Key)
	if err != nil {
		log.Printf("validate key error: %v", err)
		writeInternalError(w)
		return
	}

	status := http.StatusOK
	if !resp.Valid {
		status = http.StatusUnauthorized
	}
	writeJSON(w, status, resp)
}

// ---- JSON helpers -----------------------------------------------------------

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("writeJSON encode error: %v", err)
	}
}

func writeBadRequest(w http.ResponseWriter, msg string) {
	writeJSON(w, http.StatusBadRequest, map[string]string{"error": msg})
}

func writeNotFound(w http.ResponseWriter) {
	writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
}

func writeInternalError(w http.ResponseWriter) {
	writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
}

func methodNotAllowed(w http.ResponseWriter) {
	writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
}
