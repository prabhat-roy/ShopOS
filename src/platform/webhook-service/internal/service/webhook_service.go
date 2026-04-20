package service

import (
	"context"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/shopos/webhook-service/internal/delivery"
	"github.com/shopos/webhook-service/internal/domain"
)

// Storer is the persistence interface.
type Storer interface {
	Create(ctx context.Context, req *domain.CreateWebhookRequest) (*domain.Webhook, error)
	Get(ctx context.Context, id string) (*domain.Webhook, error)
	List(ctx context.Context, ownerID string) ([]*domain.Webhook, error)
	ListByEvent(ctx context.Context, eventTopic string) ([]*domain.Webhook, error)
	Update(ctx context.Context, id string, req *domain.UpdateWebhookRequest) (*domain.Webhook, error)
	Delete(ctx context.Context, id string) error
	SaveDelivery(ctx context.Context, d *domain.Delivery) error
}

type WebhookService struct {
	store      Storer
	dispatcher *delivery.Dispatcher
}

func New(store Storer, dispatcher *delivery.Dispatcher) *WebhookService {
	return &WebhookService{store: store, dispatcher: dispatcher}
}

func (s *WebhookService) Create(ctx context.Context, req *domain.CreateWebhookRequest) (*domain.Webhook, error) {
	if _, err := url.ParseRequestURI(req.URL); err != nil || req.URL == "" {
		return nil, domain.ErrInvalidURL
	}
	return s.store.Create(ctx, req)
}

func (s *WebhookService) Get(ctx context.Context, id string) (*domain.Webhook, error) {
	return s.store.Get(ctx, id)
}

func (s *WebhookService) List(ctx context.Context, ownerID string) ([]*domain.Webhook, error) {
	return s.store.List(ctx, ownerID)
}

func (s *WebhookService) Update(ctx context.Context, id string, req *domain.UpdateWebhookRequest) (*domain.Webhook, error) {
	if req.URL != nil {
		if _, err := url.ParseRequestURI(*req.URL); err != nil {
			return nil, domain.ErrInvalidURL
		}
	}
	return s.store.Update(ctx, id, req)
}

func (s *WebhookService) Delete(ctx context.Context, id string) error {
	return s.store.Delete(ctx, id)
}

// Dispatch looks up webhooks subscribed to eventTopic, delivers the payload,
// and persists a delivery record for each.
func (s *WebhookService) Dispatch(ctx context.Context, eventTopic string, payload []byte) {
	hooks, err := s.store.ListByEvent(ctx, eventTopic)
	if err != nil || len(hooks) == 0 {
		return
	}

	for _, hook := range hooks {
		result := s.dispatcher.Send(ctx, hook.URL, hook.Secret, payload)
		d := &domain.Delivery{
			ID:         uuid.NewString(),
			WebhookID:  hook.ID,
			EventTopic: eventTopic,
			Payload:    payload,
			StatusCode: result.StatusCode,
			Attempt:    result.Attempt,
			Success:    result.Success,
			CreatedAt:  time.Now().UTC(),
		}
		_ = s.store.SaveDelivery(ctx, d)
	}
}
