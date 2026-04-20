package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/shopos/permission-service/internal/domain"
)

// Servicer is the interface the HTTP handler depends on.
type Servicer interface {
	CreateRole(ctx context.Context, name, description string, permissions []string) (*domain.Role, error)
	GetRole(ctx context.Context, id string) (*domain.Role, error)
	ListRoles(ctx context.Context) ([]*domain.Role, error)
	DeleteRole(ctx context.Context, id string) error

	AssignRole(ctx context.Context, userID, roleID string) error
	RevokeRole(ctx context.Context, userID, roleID string) error
	GetUserRoles(ctx context.Context, userID string) ([]*domain.UserRole, error)

	Check(ctx context.Context, req domain.CheckRequest) domain.CheckResponse
}

// Handler holds the HTTP mux and service dependency.
type Handler struct {
	mux *http.ServeMux
	svc Servicer
}

// New wires all routes and returns the Handler ready to serve.
func New(svc Servicer) *Handler {
	h := &Handler{mux: http.NewServeMux(), svc: svc}
	h.registerRoutes()
	return h
}

// ServeHTTP implements http.Handler so Handler can be passed directly to http.Server.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func (h *Handler) registerRoutes() {
	h.mux.HandleFunc("/healthz", h.handleHealthz)
	h.mux.HandleFunc("/roles", h.handleRoles)
	h.mux.HandleFunc("/roles/", h.handleRoleByID)
	h.mux.HandleFunc("/users/", h.handleUsers)
	h.mux.HandleFunc("/check", h.handleCheck)
}

// GET /healthz
func (h *Handler) handleHealthz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// POST /roles  →  create role (201)
// GET  /roles  →  list roles  (200)
func (h *Handler) handleRoles(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.createRole(w, r)
	case http.MethodGet:
		h.listRoles(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// GET    /roles/{id}  →  get role  (200)
// DELETE /roles/{id}  →  delete    (204)
func (h *Handler) handleRoleByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/roles/")
	if id == "" {
		writeError(w, http.StatusBadRequest, "role id is required")
		return
	}
	switch r.Method {
	case http.MethodGet:
		h.getRole(w, r, id)
	case http.MethodDelete:
		h.deleteRole(w, r, id)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// Routes under /users/:
//
//	POST   /users/{userID}/roles              → assign role      (201)
//	DELETE /users/{userID}/roles/{roleID}     → revoke role      (204)
//	GET    /users/{userID}/roles              → list user roles  (200)
func (h *Handler) handleUsers(w http.ResponseWriter, r *http.Request) {
	// Strip leading "/users/"
	path := strings.TrimPrefix(r.URL.Path, "/users/")
	// path is now  "{userID}/roles"  or  "{userID}/roles/{roleID}"
	parts := strings.SplitN(path, "/", 3)
	// parts[0] = userID, parts[1] = "roles", parts[2] (optional) = roleID

	if len(parts) < 2 || parts[1] != "roles" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	userID := parts[0]
	if userID == "" {
		writeError(w, http.StatusBadRequest, "user_id is required")
		return
	}

	if len(parts) == 3 && parts[2] != "" {
		// /users/{userID}/roles/{roleID}
		roleID := parts[2]
		if r.Method != http.MethodDelete {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		h.revokeRole(w, r, userID, roleID)
		return
	}

	// /users/{userID}/roles
	switch r.Method {
	case http.MethodPost:
		h.assignRole(w, r, userID)
	case http.MethodGet:
		h.getUserRoles(w, r, userID)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// POST /check
func (h *Handler) handleCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req domain.CheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	resp := h.svc.Check(r.Context(), req)
	writeJSON(w, http.StatusOK, resp)
}

// ---- role handlers ----------------------------------------------------------

func (h *Handler) createRole(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Permissions []string `json:"permissions"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if body.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	role, err := h.svc.CreateRole(r.Context(), body.Name, body.Description, body.Permissions)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, role)
}

func (h *Handler) listRoles(w http.ResponseWriter, r *http.Request) {
	roles, err := h.svc.ListRoles(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if roles == nil {
		roles = []*domain.Role{}
	}
	writeJSON(w, http.StatusOK, roles)
}

func (h *Handler) getRole(w http.ResponseWriter, r *http.Request, id string) {
	role, err := h.svc.GetRole(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, "role not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, role)
}

func (h *Handler) deleteRole(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.svc.DeleteRole(r.Context(), id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, "role not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ---- user-role handlers -----------------------------------------------------

func (h *Handler) assignRole(w http.ResponseWriter, r *http.Request, userID string) {
	var body struct {
		RoleID string `json:"role_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if body.RoleID == "" {
		writeError(w, http.StatusBadRequest, "role_id is required")
		return
	}
	if err := h.svc.AssignRole(r.Context(), userID, body.RoleID); err != nil {
		switch {
		case errors.Is(err, domain.ErrAlreadyAssigned):
			writeError(w, http.StatusConflict, "role already assigned to user")
		case errors.Is(err, domain.ErrNotFound):
			writeError(w, http.StatusNotFound, "role not found")
		default:
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) revokeRole(w http.ResponseWriter, r *http.Request, userID, roleID string) {
	if err := h.svc.RevokeRole(r.Context(), userID, roleID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, "user-role binding not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) getUserRoles(w http.ResponseWriter, r *http.Request, userID string) {
	roles, err := h.svc.GetUserRoles(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if roles == nil {
		roles = []*domain.UserRole{}
	}
	writeJSON(w, http.StatusOK, roles)
}

// ---- helpers ----------------------------------------------------------------

type errorResponse struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorResponse{Error: msg})
}
