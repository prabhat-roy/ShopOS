package service

import (
	"context"
	"github.com/shopos/partner-bff/internal/domain"
)

type B2BService interface {
	GetOrganization(ctx context.Context, partnerID string) (*domain.Organization, error)
	ListContracts(ctx context.Context, partnerID string) ([]*domain.Contract, error)
	ListQuotes(ctx context.Context, partnerID string) ([]*domain.Quote, error)
	CreateQuote(ctx context.Context, partnerID string, req *domain.CreateQuoteRequest) (*domain.Quote, error)
}

type GRPCB2BService struct{ addr string }

func NewGRPCB2BService(addr string) B2BService { return &GRPCB2BService{addr: addr} }

func (s *GRPCB2BService) GetOrganization(_ context.Context, _ string) (*domain.Organization, error) {
	return nil, ErrNotImplemented
}
func (s *GRPCB2BService) ListContracts(_ context.Context, _ string) ([]*domain.Contract, error) {
	return nil, ErrNotImplemented
}
func (s *GRPCB2BService) ListQuotes(_ context.Context, _ string) ([]*domain.Quote, error) {
	return nil, ErrNotImplemented
}
func (s *GRPCB2BService) CreateQuote(_ context.Context, _ string, _ *domain.CreateQuoteRequest) (*domain.Quote, error) {
	return nil, ErrNotImplemented
}
