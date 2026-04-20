package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/shopos/in-app-notification-service/internal/domain"
	"github.com/shopos/in-app-notification-service/internal/service"
)

// Handler holds the HTTP handler dependencies.
type Handler struct {
	svc service.Servicer
}

// New creates a new Handler.
func New(svc service.Servicer) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes wires the handler methods to the given ServeMux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/healthz", h.healthz)
	mux.HandleFunc("/notifications/", h.routeNotifications)
}

// healthz returns a simple liveness response.
func (h *Handler) healthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

// routeNotifications dispatches to the correct sub-handler based on path and method.
//
// Route table:
//   POST   /notifications/{userId}                    → sendNotification
//   GET    /notifications/{userId}                    → getNotifications
//   POST   /notifications/{userId}/read-all           → markAllRead
//   PATCH  /notifications/{userId}/{notifId}/read     → markRead
//   DELETE /notifications/{userId}/{notifId}          → deleteNotification
//   DELETE /notifications/{userId}                    → clearAll
//   GET    /notifications/{userId}/count              → getUnreadCount
func (h *Handler) routeNotifications(w http.ResponseWriter, r *http.Request) {
	// Strip leading "/notifications/" prefix and split segments.
	path := strings.TrimPrefix(r.URL.Path, "/notifications/")
	path = strings.TrimSuffix(path, "/")
	segments := strings.Split(path, "/")

	switch len(segments) {
	case 1:
		userID := segments[0]
		switch r.Method {
		case http.MethodPost:
			h.sendNotification(w, r, userID)
		case http.MethodGet:
			h.getNotifications(w, r, userID)
		case http.MethodDelete:
			h.clearAll(w, r, userID)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}

	case 2:
		userID := segments[0]
		second := segments[1]
		switch {
		case second == "read-all" && r.Method == http.MethodPost:
			h.markAllRead(w, r, userID)
		case second == "count" && r.Method == http.MethodGet:
			h.getUnreadCount(w, r, userID)
		case r.Method == http.MethodDelete:
			// DELETE /notifications/{userId}/{notifId}
			h.deleteNotificationByID(w, r, userID, second)
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}

	case 3:
		userID := segments[0]
		notifID := segments[1]
		action := segments[2]
		if action == "read" && r.Method == http.MethodPatch {
			h.markRead(w, r, userID, notifID)
			return
		}
		http.Error(w, "not found", http.StatusNotFound)

	default:
		http.Error(w, "not found", http.StatusNotFound)
	}
}

// sendNotification handles POST /notifications/{userId}.
func (h *Handler) sendNotification(w http.ResponseWriter, r *http.Request, userID string) {
	var req domain.SendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if msg := req.Validate(); msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}
	notif, err := h.svc.SendNotification(r.Context(), userID, req.Type, req.Title, req.Body, req.Link)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, notif)
}

// getNotifications handles GET /notifications/{userId}.
func (h *Handler) getNotifications(w http.ResponseWriter, r *http.Request, userID string) {
	q := r.URL.Query()
	unreadOnly := strings.EqualFold(q.Get("unreadOnly"), "true")
	limit := parseIntQuery(q.Get("limit"), 20)
	offset := parseIntQuery(q.Get("offset"), 0)

	page, err := h.svc.GetNotifications(r.Context(), userID, unreadOnly, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, page)
}

// markAllRead handles POST /notifications/{userId}/read-all.
func (h *Handler) markAllRead(w http.ResponseWriter, r *http.Request, userID string) {
	if err := h.svc.MarkAllAsRead(r.Context(), userID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// markRead handles PATCH /notifications/{userId}/{notifId}/read.
func (h *Handler) markRead(w http.ResponseWriter, r *http.Request, userID, notifID string) {
	if err := h.svc.MarkAsRead(r.Context(), userID, notifID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// deleteNotification handles DELETE /notifications/{userId}/{notifId}.
// This route is served when path has 2 segments where the second is not "count" or "read-all".
func (h *Handler) deleteNotificationByID(w http.ResponseWriter, r *http.Request, userID, notifID string) {
	if err := h.svc.DeleteNotification(r.Context(), userID, notifID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// clearAll handles DELETE /notifications/{userId}.
func (h *Handler) clearAll(w http.ResponseWriter, r *http.Request, userID string) {
	if err := h.svc.ClearNotifications(r.Context(), userID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// getUnreadCount handles GET /notifications/{userId}/count.
func (h *Handler) getUnreadCount(w http.ResponseWriter, r *http.Request, userID string) {
	count, err := h.svc.GetUnreadCount(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]int{"unread": count})
}

// writeJSON serialises v as JSON and writes it with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError writes a JSON error response.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// parseIntQuery parses an integer query parameter, returning def on failure.
func parseIntQuery(s string, def int) int {
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil || v < 0 {
		return def
	}
	return v
}
