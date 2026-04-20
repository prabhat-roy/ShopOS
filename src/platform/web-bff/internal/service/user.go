package service

import (
	"context"

	"github.com/shopos/web-bff/internal/domain"
)

type UserService interface {
	GetProfile(ctx context.Context, userID string) (*domain.UserProfile, error)
	UpdateProfile(ctx context.Context, userID string, req *domain.UpdateProfileRequest) (*domain.UserProfile, error)
}

// GRPCUserService connects to identity/user-service (port 50061).
// TODO: Replace stub bodies with generated proto client calls once proto/identity/ is compiled.
type GRPCUserService struct{ addr string }

func NewGRPCUserService(addr string) UserService {
	return &GRPCUserService{addr: addr}
}

func (s *GRPCUserService) GetProfile(_ context.Context, _ string) (*domain.UserProfile, error) {
	return nil, ErrNotImplemented
}

func (s *GRPCUserService) UpdateProfile(_ context.Context, _ string, _ *domain.UpdateProfileRequest) (*domain.UserProfile, error) {
	return nil, ErrNotImplemented
}
