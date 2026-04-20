package handler

import (
	"encoding/json"
	"net/http"

	"github.com/shopos/web-bff/internal/domain"
	"github.com/shopos/web-bff/internal/service"
)

type UserHandler struct {
	svc service.UserService
}

func NewUserHandler(svc service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	profile, err := h.svc.GetProfile(r.Context(), userID(r))
	if err != nil {
		Error(w, err)
		return
	}
	JSON(w, http.StatusOK, profile)
}

func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	var req domain.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	profile, err := h.svc.UpdateProfile(r.Context(), userID(r), &req)
	if err != nil {
		Error(w, err)
		return
	}
	JSON(w, http.StatusOK, profile)
}
