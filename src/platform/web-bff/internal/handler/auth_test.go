package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shopos/web-bff/internal/domain"
	"github.com/shopos/web-bff/internal/handler"
	"github.com/shopos/web-bff/internal/service"
)

func TestLogin_Success(t *testing.T) {
	svc := &mockAuthService{
		loginFn: func(_ context.Context, email, _ string) (*domain.LoginResponse, error) {
			return &domain.LoginResponse{AccessToken: "tok", UserID: "u1"}, nil
		},
	}
	h := handler.NewAuthHandler(svc)
	body, _ := json.Marshal(map[string]string{"email": "a@b.com", "password": "secret"})
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	h.Login(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestLogin_MissingFields_Returns400(t *testing.T) {
	h := handler.NewAuthHandler(&mockAuthService{})
	body, _ := json.Marshal(map[string]string{"email": "a@b.com"})
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	h.Login(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestLogin_ServiceError_Returns501(t *testing.T) {
	svc := &mockAuthService{
		loginFn: func(_ context.Context, _, _ string) (*domain.LoginResponse, error) {
			return nil, service.ErrNotImplemented
		},
	}
	h := handler.NewAuthHandler(svc)
	body, _ := json.Marshal(map[string]string{"email": "a@b.com", "password": "secret"})
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	h.Login(rr, req)

	if rr.Code != http.StatusNotImplemented {
		t.Errorf("expected 501, got %d", rr.Code)
	}
}

func TestRegister_MissingFirstName_Returns400(t *testing.T) {
	h := handler.NewAuthHandler(&mockAuthService{})
	body, _ := json.Marshal(map[string]string{"email": "a@b.com", "password": "secret"})
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	h.Register(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestRefresh_MissingToken_Returns400(t *testing.T) {
	h := handler.NewAuthHandler(&mockAuthService{})
	body, _ := json.Marshal(map[string]string{})
	req := httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	h.Refresh(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}
