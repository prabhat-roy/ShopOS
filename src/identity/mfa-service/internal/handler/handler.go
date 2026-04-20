package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/shopos/mfa-service/internal/domain"
)

// Servicer is the service interface consumed by the HTTP handler layer.
type Servicer interface {
	Enroll(ctx context.Context, userID string) (*domain.MFASetup, error)
	Confirm(ctx context.Context, userID, code string) error
	Verify(ctx context.Context, req *domain.VerifyRequest) (*domain.VerifyResponse, error)
	Disable(ctx context.Context, userID string) error
	GetStatus(ctx context.Context, userID string) (domain.MFAStatus, error)
}

// Handler wires all HTTP routes for the mfa-service.
type Handler struct {
	svc    Servicer
	mux    *http.ServeMux
	logger *log.Logger
}

// New constructs a Handler and registers all routes on an internal ServeMux.
func New(svc Servicer, logger *log.Logger) *Handler {
	h := &Handler{
		svc:    svc,
		mux:    http.NewServeMux(),
		logger: logger,
	}
	h.registerRoutes()
	return h
}

// ServeHTTP implements http.Handler so Handler can be passed directly to
// http.ListenAndServe.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func (h *Handler) registerRoutes() {
	h.mux.HandleFunc("/healthz", h.handleHealthz)
	h.mux.HandleFunc("/mfa/", h.routeMFA)
}

// handleHealthz returns a simple liveness response.
func (h *Handler) handleHealthz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// routeMFA dispatches /mfa/{userID}[/sub-resource] requests.
//
// Route table:
//
//	POST   /mfa/{userID}/enroll   → enroll
//	POST   /mfa/{userID}/confirm  → confirm
//	POST   /mfa/{userID}/verify   → verify
//	DELETE /mfa/{userID}          → disable
//	GET    /mfa/{userID}/status   → getStatus
func (h *Handler) routeMFA(w http.ResponseWriter, r *http.Request) {
	// Strip leading "/mfa/" and split remaining path segments.
	path := r.URL.Path[len("/mfa/"):]
	userID, sub := splitPath(path)

	if userID == "" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	switch {
	case sub == "enroll" && r.Method == http.MethodPost:
		h.handleEnroll(w, r, userID)
	case sub == "confirm" && r.Method == http.MethodPost:
		h.handleConfirm(w, r, userID)
	case sub == "verify" && r.Method == http.MethodPost:
		h.handleVerify(w, r, userID)
	case sub == "" && r.Method == http.MethodDelete:
		h.handleDisable(w, r, userID)
	case sub == "status" && r.Method == http.MethodGet:
		h.handleGetStatus(w, r, userID)
	default:
		writeError(w, http.StatusNotFound, "not found")
	}
}

// handleEnroll generates a new TOTP secret for the user.
// POST /mfa/{userID}/enroll
// Response 201: { user_id, qr_code_url, backup_codes, status }
func (h *Handler) handleEnroll(w http.ResponseWriter, r *http.Request, userID string) {
	setup, err := h.svc.Enroll(r.Context(), userID)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"user_id":      setup.UserID,
		"qr_code_url":  setup.QRCodeURL,
		"backup_codes": setup.BackupCodes,
		"status":       setup.Status,
	})
}

// handleConfirm activates MFA after the user scans the QR code and submits a
// live TOTP code.
// POST /mfa/{userID}/confirm  body: { "code": "123456" }
// Response 200: { "message": "mfa enabled" }
func (h *Handler) handleConfirm(w http.ResponseWriter, r *http.Request, userID string) {
	var body struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Code == "" {
		writeError(w, http.StatusBadRequest, "field 'code' is required")
		return
	}

	if err := h.svc.Confirm(r.Context(), userID, body.Code); err != nil {
		h.handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "mfa enabled"})
}

// handleVerify validates a code during login — accepts TOTP or backup code.
// POST /mfa/{userID}/verify  body: { "code": "..." }
// Response 200: { "valid": true, "is_backup_code": false }
func (h *Handler) handleVerify(w http.ResponseWriter, r *http.Request, userID string) {
	var body struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Code == "" {
		writeError(w, http.StatusBadRequest, "field 'code' is required")
		return
	}

	resp, err := h.svc.Verify(r.Context(), &domain.VerifyRequest{
		UserID: userID,
		Code:   body.Code,
	})
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"valid":          resp.Valid,
		"is_backup_code": resp.IsBackupCode,
	})
}

// handleDisable turns off MFA for the user.
// DELETE /mfa/{userID}
// Response 204: no body
func (h *Handler) handleDisable(w http.ResponseWriter, r *http.Request, userID string) {
	if err := h.svc.Disable(r.Context(), userID); err != nil {
		h.handleServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleGetStatus returns the current MFA enrollment status.
// GET /mfa/{userID}/status
// Response 200: { "user_id": "...", "status": "enabled" }
func (h *Handler) handleGetStatus(w http.ResponseWriter, r *http.Request, userID string) {
	status, err := h.svc.GetStatus(r.Context(), userID)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"user_id": userID,
		"status":  string(status),
	})
}

// handleServiceError maps domain sentinel errors to appropriate HTTP responses.
func (h *Handler) handleServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		writeError(w, http.StatusNotFound, "mfa not configured for this user")
	case errors.Is(err, domain.ErrAlreadyEnabled):
		writeError(w, http.StatusConflict, "mfa is already enabled")
	case errors.Is(err, domain.ErrInvalidCode):
		writeError(w, http.StatusBadRequest, "invalid or expired code")
	default:
		h.logger.Printf("ERROR: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}

// writeJSON encodes v as JSON and writes it with the given status code.
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		// Headers already sent; nothing useful we can do.
		return
	}
}

// writeError writes a JSON error body.
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

// splitPath splits "userID/subResource" or "userID" into its two parts.
func splitPath(path string) (userID, sub string) {
	for i, c := range path {
		if c == '/' {
			return path[:i], path[i+1:]
		}
	}
	return path, ""
}
