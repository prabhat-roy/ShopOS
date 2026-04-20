package handler

import (
	"encoding/json"
	"net/http"

	"github.com/shopos/web-bff/internal/domain"
	"github.com/shopos/web-bff/internal/service"
)

type AuthHandler struct {
	svc service.AuthService
}

func NewAuthHandler(svc service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req domain.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if req.Email == "" || req.Password == "" {
		JSON(w, http.StatusBadRequest, map[string]string{"error": "email and password are required"})
		return
	}
	resp, err := h.svc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		Error(w, err)
		return
	}
	JSON(w, http.StatusOK, resp)
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req domain.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if req.Email == "" || req.Password == "" || req.FirstName == "" {
		JSON(w, http.StatusBadRequest, map[string]string{"error": "email, password and first_name are required"})
		return
	}
	resp, err := h.svc.Register(r.Context(), &req)
	if err != nil {
		Error(w, err)
		return
	}
	JSON(w, http.StatusCreated, resp)
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req domain.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if req.RefreshToken == "" {
		JSON(w, http.StatusBadRequest, map[string]string{"error": "refresh_token is required"})
		return
	}
	resp, err := h.svc.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		Error(w, err)
		return
	}
	JSON(w, http.StatusOK, resp)
}
