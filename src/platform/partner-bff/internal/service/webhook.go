package service

import (
	"context"
	"github.com/shopos/partner-bff/internal/domain"
)

type WebhookService interface {
	ListWebhooks(ctx context.Context, partnerID string) ([]*domain.Webhook, error)
	CreateWebhook(ctx context.Context, partnerID string, req *domain.CreateWebhookRequest) (*domain.Webhook, error)
	DeleteWebhook(ctx context.Context, partnerID, webhookID string) error
}

type GRPCWebhookService struct{ addr string }

func NewGRPCWebhookService(addr string) WebhookService { return &GRPCWebhookService{addr: addr} }

func (s *GRPCWebhookService) ListWebhooks(_ context.Context, _ string) ([]*domain.Webhook, error) {
	return nil, ErrNotImplemented
}
func (s *GRPCWebhookService) CreateWebhook(_ context.Context, _ string, _ *domain.CreateWebhookRequest) (*domain.Webhook, error) {
	return nil, ErrNotImplemented
}
func (s *GRPCWebhookService) DeleteWebhook(_ context.Context, _, _ string) error {
	return ErrNotImplemented
}
