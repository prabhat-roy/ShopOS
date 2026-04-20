package handler

import (
	"encoding/json"
	"net/http"

	"github.com/shopos/partner-bff/internal/domain"
	"github.com/shopos/partner-bff/internal/service"
)

type B2BHandler struct{ svc service.B2BService }

func NewB2BHandler(svc service.B2BService) *B2BHandler { return &B2BHandler{svc: svc} }

func (h *B2BHandler) GetOrganization(w http.ResponseWriter, r *http.Request) {
	org, err := h.svc.GetOrganization(r.Context(), partnerID(r))
	if err != nil {
		Error(w, err)
		return
	}
	JSON(w, http.StatusOK, org)
}

func (h *B2BHandler) ListContracts(w http.ResponseWriter, r *http.Request) {
	contracts, err := h.svc.ListContracts(r.Context(), partnerID(r))
	if err != nil {
		Error(w, err)
		return
	}
	JSON(w, http.StatusOK, map[string]any{"items": contracts})
}

func (h *B2BHandler) ListQuotes(w http.ResponseWriter, r *http.Request) {
	quotes, err := h.svc.ListQuotes(r.Context(), partnerID(r))
	if err != nil {
		Error(w, err)
		return
	}
	JSON(w, http.StatusOK, map[string]any{"items": quotes})
}

func (h *B2BHandler) CreateQuote(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateQuoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if len(req.Items) == 0 {
		JSON(w, http.StatusBadRequest, map[string]string{"error": "items are required"})
		return
	}
	quote, err := h.svc.CreateQuote(r.Context(), partnerID(r), &req)
	if err != nil {
		Error(w, err)
		return
	}
	JSON(w, http.StatusCreated, quote)
}
