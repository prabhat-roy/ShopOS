package handler

import (
	"encoding/json"
	"net/http"

	"github.com/shopos/partner-bff/internal/domain"
	"github.com/shopos/partner-bff/internal/service"
)

type WebhookHandler struct{ svc service.WebhookService }

func NewWebhookHandler(svc service.WebhookService) *WebhookHandler {
	return &WebhookHandler{svc: svc}
}

func (h *WebhookHandler) ListWebhooks(w http.ResponseWriter, r *http.Request) {
	hooks, err := h.svc.ListWebhooks(r.Context(), partnerID(r))
	if err != nil {
		Error(w, err)
		return
	}
	JSON(w, http.StatusOK, map[string]any{"items": hooks})
}

func (h *WebhookHandler) CreateWebhook(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if req.URL == "" || len(req.Events) == 0 {
		JSON(w, http.StatusBadRequest, map[string]string{"error": "url and events are required"})
		return
	}
	hook, err := h.svc.CreateWebhook(r.Context(), partnerID(r), &req)
	if err != nil {
		Error(w, err)
		return
	}
	JSON(w, http.StatusCreated, hook)
}

func (h *WebhookHandler) DeleteWebhook(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.DeleteWebhook(r.Context(), partnerID(r), r.PathValue("id")); err != nil {
		Error(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
