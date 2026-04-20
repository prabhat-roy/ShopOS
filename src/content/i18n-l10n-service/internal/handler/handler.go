package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/shopos/i18n-l10n-service/internal/domain"
	"github.com/shopos/i18n-l10n-service/internal/service"
)

// Handler bundles the HTTP router with its dependencies.
type Handler struct {
	svc service.Servicer
	mux *http.ServeMux
}

// New creates a Handler and registers all routes.
func New(svc service.Servicer) *Handler {
	h := &Handler{svc: svc, mux: http.NewServeMux()}
	h.routes()
	return h
}

// ServeHTTP satisfies http.Handler.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func (h *Handler) routes() {
	h.mux.HandleFunc("GET /healthz", h.healthz)
	h.mux.HandleFunc("GET /i18n/locales", h.listLocales)
	h.mux.HandleFunc("GET /i18n/{locale}/namespaces", h.listNamespaces)
	h.mux.HandleFunc("GET /i18n/{locale}/{namespace}/{key}", h.getTranslation)
	h.mux.HandleFunc("GET /i18n/{locale}/{namespace}", h.getNamespace)
	h.mux.HandleFunc("PUT /i18n/{locale}/{namespace}/{key}", h.upsertTranslation)
	h.mux.HandleFunc("PUT /i18n/{locale}/{namespace}", h.bulkUpsert)
	h.mux.HandleFunc("DELETE /i18n/{locale}/{namespace}/{key}", h.deleteTranslation)
}

// ─── Handlers ────────────────────────────────────────────────────────────────

func (h *Handler) healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) listLocales(w http.ResponseWriter, r *http.Request) {
	locales, err := h.svc.ListLocales()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if locales == nil {
		locales = []string{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"locales": locales})
}

func (h *Handler) listNamespaces(w http.ResponseWriter, r *http.Request) {
	locale := r.PathValue("locale")
	namespaces, err := h.svc.ListNamespaces(locale)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if namespaces == nil {
		namespaces = []string{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"namespaces": namespaces})
}

func (h *Handler) getTranslation(w http.ResponseWriter, r *http.Request) {
	locale := r.PathValue("locale")
	namespace := r.PathValue("namespace")
	key := r.PathValue("key")

	value, err := h.svc.GetTranslation(locale, namespace, key)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, "translation not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"locale":    locale,
		"namespace": namespace,
		"key":       key,
		"value":     value,
	})
}

func (h *Handler) getNamespace(w http.ResponseWriter, r *http.Request) {
	locale := r.PathValue("locale")
	namespace := r.PathValue("namespace")

	translations, err := h.svc.GetNamespace(locale, namespace)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"locale":       locale,
		"namespace":    namespace,
		"translations": translations,
	})
}

func (h *Handler) upsertTranslation(w http.ResponseWriter, r *http.Request) {
	locale := r.PathValue("locale")
	namespace := r.PathValue("namespace")
	key := r.PathValue("key")

	var body struct {
		Value string `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if body.Value == "" {
		writeError(w, http.StatusBadRequest, "value is required")
		return
	}

	t := domain.Translation{
		Locale:    locale,
		Namespace: namespace,
		Key:       key,
		Value:     body.Value,
	}

	// Determine if existing (to return 200 vs 201)
	_, errGet := h.svc.GetTranslation(locale, namespace, key)
	isNew := errors.Is(errGet, domain.ErrNotFound)

	if err := h.svc.UpsertTranslation(t); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	statusCode := http.StatusOK
	if isNew {
		statusCode = http.StatusCreated
	}
	writeJSON(w, statusCode, map[string]string{
		"locale":    locale,
		"namespace": namespace,
		"key":       key,
		"value":     body.Value,
	})
}

func (h *Handler) bulkUpsert(w http.ResponseWriter, r *http.Request) {
	locale := r.PathValue("locale")
	namespace := r.PathValue("namespace")

	var translations map[string]string
	if err := json.NewDecoder(r.Body).Decode(&translations); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body — expected map[string]string")
		return
	}
	if len(translations) == 0 {
		writeError(w, http.StatusBadRequest, "translations map must not be empty")
		return
	}

	req := domain.BulkUpsertRequest{
		Locale:       locale,
		Namespace:    namespace,
		Translations: translations,
	}
	if err := h.svc.BulkUpsert(req); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"locale":    locale,
		"namespace": namespace,
		"count":     len(translations),
	})
}

func (h *Handler) deleteTranslation(w http.ResponseWriter, r *http.Request) {
	locale := r.PathValue("locale")
	namespace := r.PathValue("namespace")
	key := r.PathValue("key")

	if err := h.svc.DeleteTranslation(locale, namespace, key); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, "translation not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("[handler] json encode error: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
