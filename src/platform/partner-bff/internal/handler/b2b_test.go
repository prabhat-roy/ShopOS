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

func TestGetOrganization(t *testing.T) {
	svc := &mockB2BService{
		org: &domain.Organization{ID: "org1", Name: "Acme Corp"},
	}
	h := handler.NewB2BHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/b2b/organization", nil)
	req.Header.Set("X-Partner-ID", "partnerA")
	w := httptest.NewRecorder()

	h.GetOrganization(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var body domain.Organization
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if body.ID != "org1" {
		t.Errorf("expected org1, got %q", body.ID)
	}
}

func TestListContracts(t *testing.T) {
	svc := &mockB2BService{
		contracts: []*domain.Contract{
			{ID: "c1", PartnerID: "partnerA", Status: "active"},
		},
	}
	h := handler.NewB2BHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/b2b/contracts", nil)
	req.Header.Set("X-Partner-ID", "partnerA")
	w := httptest.NewRecorder()

	h.ListContracts(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var body struct {
		Items []*domain.Contract `json:"items"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(body.Items) != 1 || body.Items[0].ID != "c1" {
		t.Errorf("unexpected contracts: %+v", body)
	}
}

func TestListQuotes(t *testing.T) {
	svc := &mockB2BService{
		quotes: []*domain.Quote{
			{ID: "q1", PartnerID: "partnerA", Status: "draft"},
		},
	}
	h := handler.NewB2BHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/b2b/quotes", nil)
	req.Header.Set("X-Partner-ID", "partnerA")
	w := httptest.NewRecorder()

	h.ListQuotes(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var body struct {
		Items []*domain.Quote `json:"items"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(body.Items) != 1 {
		t.Errorf("expected 1 quote, got %d", len(body.Items))
	}
}

func TestCreateQuote(t *testing.T) {
	svc := &mockB2BService{
		quote: &domain.Quote{ID: "q2", PartnerID: "partnerA", Status: "draft"},
	}
	h := handler.NewB2BHandler(svc)

	payload := domain.CreateQuoteRequest{
		Items: []domain.OrderItem{{ProductID: "p1", Quantity: 100}},
	}
	b, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/b2b/quotes", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Partner-ID", "partnerA")
	w := httptest.NewRecorder()

	h.CreateQuote(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
}

func TestCreateQuoteBadBody(t *testing.T) {
	svc := &mockB2BService{}
	h := handler.NewB2BHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/b2b/quotes", bytes.NewReader([]byte("bad")))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Partner-ID", "partnerA")
	w := httptest.NewRecorder()

	h.CreateQuote(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
