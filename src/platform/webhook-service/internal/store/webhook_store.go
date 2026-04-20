package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/shopos/webhook-service/internal/domain"
)

type WebhookStore struct {
	db *sql.DB
}

func New(db *sql.DB) *WebhookStore {
	return &WebhookStore{db: db}
}

func (s *WebhookStore) Create(ctx context.Context, req *domain.CreateWebhookRequest) (*domain.Webhook, error) {
	w := &domain.Webhook{
		ID:        uuid.NewString(),
		OwnerID:   req.OwnerID,
		URL:       req.URL,
		Events:    req.Events,
		Secret:    req.Secret,
		Active:    true,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO webhooks (id, owner_id, url, events, secret, active, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		w.ID, w.OwnerID, w.URL, pq.Array(w.Events), w.Secret, w.Active, w.CreatedAt, w.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("store.Create: %w", err)
	}
	return w, nil
}

func (s *WebhookStore) Get(ctx context.Context, id string) (*domain.Webhook, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, owner_id, url, events, secret, active, created_at, updated_at
		 FROM webhooks WHERE id = $1`, id)
	return scanWebhook(row)
}

func (s *WebhookStore) List(ctx context.Context, ownerID string) ([]*domain.Webhook, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, owner_id, url, events, secret, active, created_at, updated_at
		 FROM webhooks WHERE owner_id = $1 ORDER BY created_at DESC`, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*domain.Webhook
	for rows.Next() {
		w, err := scanWebhookRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	return out, rows.Err()
}

func (s *WebhookStore) ListByEvent(ctx context.Context, eventTopic string) ([]*domain.Webhook, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, owner_id, url, events, secret, active, created_at, updated_at
		 FROM webhooks WHERE active = true AND $1 = ANY(events)`, eventTopic)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*domain.Webhook
	for rows.Next() {
		w, err := scanWebhookRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	return out, rows.Err()
}

func (s *WebhookStore) Update(ctx context.Context, id string, req *domain.UpdateWebhookRequest) (*domain.Webhook, error) {
	w, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.URL != nil {
		w.URL = *req.URL
	}
	if req.Events != nil {
		w.Events = req.Events
	}
	if req.Active != nil {
		w.Active = *req.Active
	}
	w.UpdatedAt = time.Now().UTC()

	_, err = s.db.ExecContext(ctx,
		`UPDATE webhooks SET url=$1, events=$2, active=$3, updated_at=$4 WHERE id=$5`,
		w.URL, pq.Array(w.Events), w.Active, w.UpdatedAt, id,
	)
	if err != nil {
		return nil, fmt.Errorf("store.Update: %w", err)
	}
	return w, nil
}

func (s *WebhookStore) Delete(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM webhooks WHERE id = $1`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (s *WebhookStore) SaveDelivery(ctx context.Context, d *domain.Delivery) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO webhook_deliveries (id, webhook_id, event_topic, payload, status_code, attempt, success, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		d.ID, d.WebhookID, d.EventTopic, d.Payload, d.StatusCode, d.Attempt, d.Success, d.CreatedAt,
	)
	return err
}

type scanner interface {
	Scan(dest ...any) error
}

func scanWebhook(row *sql.Row) (*domain.Webhook, error) {
	var w domain.Webhook
	err := row.Scan(&w.ID, &w.OwnerID, &w.URL, pq.Array(&w.Events), &w.Secret, &w.Active, &w.CreatedAt, &w.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return &w, err
}

func scanWebhookRow(rows *sql.Rows) (*domain.Webhook, error) {
	var w domain.Webhook
	err := rows.Scan(&w.ID, &w.OwnerID, &w.URL, pq.Array(&w.Events), &w.Secret, &w.Active, &w.CreatedAt, &w.UpdatedAt)
	return &w, err
}
