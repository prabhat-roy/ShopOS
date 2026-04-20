package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shopos/partner-bff/internal/domain"
	"github.com/shopos/partner-bff/internal/handler"
)

func TestListWebhooks(t *testing.T) {
	svc := &mockWebhookService{
		webhooks: []*domain.Webhook{
			{ID: "wh1", URL: "https://example.com/hook", Events: []string{"order.placed"}},
		},
	}
	h := handler.NewWebhookHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/webhooks", nil)
	req.Header.Set("X-Partner-ID", "partnerA")
	w := httptest.NewRecorder()

	h.ListWebhooks(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var body struct {
		Items []*domain.Webhook `json:"items"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(body.Items) != 1 || body.Items[0].ID != "wh1" {
		t.Errorf("unexpected webhooks: %+v", body)
	}
}

func TestCreateWebhook(t *testing.T) {
	svc := &mockWebhookService{
		webhook: &domain.Webhook{ID: "wh2", URL: "https://example.com/hook2"},
	}
	h := handler.NewWebhookHandler(svc)

	payload := domain.CreateWebhookRequest{
		URL:    "https://example.com/hook2",
		Events: []string{"order.placed", "payment.processed"},
	}
	b, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/webhooks", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Partner-ID", "partnerA")
	w := httptest.NewRecorder()

	h.CreateWebhook(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
}

func TestDeleteWebhook(t *testing.T) {
	svc := &mockWebhookService{}
	h := handler.NewWebhookHandler(svc)

	req := httptest.NewRequest(http.MethodDelete, "/webhooks/wh1", nil)
	req.SetPathValue("id", "wh1")
	req.Header.Set("X-Partner-ID", "partnerA")
	w := httptest.NewRecorder()

	h.DeleteWebhook(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}

func TestCreateWebhookMissingFields(t *testing.T) {
	svc := &mockWebhookService{}
	h := handler.NewWebhookHandler(svc)

	payload := domain.CreateWebhookRequest{URL: "https://example.com/hook"}
	b, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/webhooks", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Partner-ID", "partnerA")
	w := httptest.NewRecorder()

	h.CreateWebhook(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
