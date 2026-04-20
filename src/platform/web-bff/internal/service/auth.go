package service

import (
	"context"

	"github.com/shopos/web-bff/internal/domain"
)

type AuthService interface {
	Login(ctx context.Context, email, password string) (*domain.LoginResponse, error)
	Register(ctx context.Context, req *domain.RegisterRequest) (*domain.RegisterResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*domain.TokenPair, error)
}

// GRPCAuthService connects to identity/auth-service (port 50060).
// TODO: Replace stub bodies with generated proto client calls once proto/identity/ is compiled.
type GRPCAuthService struct{ addr string }

func NewGRPCAuthService(addr string) AuthService {
	return &GRPCAuthService{addr: addr}
}

func (s *GRPCAuthService) Login(_ context.Context, _, _ string) (*domain.LoginResponse, error) {
	return nil, ErrNotImplemented
}

func (s *GRPCAuthService) Register(_ context.Context, _ *domain.RegisterRequest) (*domain.RegisterResponse, error) {
	return nil, ErrNotImplemented
}

func (s *GRPCAuthService) RefreshToken(_ context.Context, _ string) (*domain.TokenPair, error) {
	return nil, ErrNotImplemented
}
