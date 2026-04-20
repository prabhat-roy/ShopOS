package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopos/quote-rfq-service/internal/domain"
	"github.com/shopos/quote-rfq-service/internal/handler"
	"github.com/shopos/quote-rfq-service/internal/service"
)

// --- mock service ---

type mockService struct {
	createFn       func(req service.CreateRFQRequest) (*domain.Quote, error)
	getFn          func(id uuid.UUID) (*domain.Quote, error)
	listFn         func(orgID *uuid.UUID, status *domain.QuoteStatus) ([]*domain.Quote, error)
	submitFn       func(id uuid.UUID) error
	reviewFn       func(id uuid.UUID) error
	provideFn      func(id uuid.UUID, req service.ProvideQuoteRequest) (*domain.Quote, error)
	acceptFn       func(id uuid.UUID) error
	rejectFn       func(id uuid.UUID) error
	expireFn       func(id uuid.UUID) error
	cancelFn       func(id uuid.UUID) error
	updateNotesFn  func(id uuid.UUID, notes string) error
}

func (m *mockService) CreateRFQ(req service.CreateRFQRequest) (*domain.Quote, error) {
	return m.createFn(req)
}
func (m *mockService) GetQuote(id uuid.UUID) (*domain.Quote, error) { return m.getFn(id) }
func (m *mockService) ListQuotes(orgID *uuid.UUID, status *domain.QuoteStatus) ([]*domain.Quote, error) {
	return m.listFn(orgID, status)
}
func (m *mockService) SubmitRFQ(id uuid.UUID) error { return m.submitFn(id) }
func (m *mockService) ReviewQuote(id uuid.UUID) error { return m.reviewFn(id) }
func (m *mockService) ProvideQuote(id uuid.UUID, req service.ProvideQuoteRequest) (*domain.Quote, error) {
	return m.provideFn(id, req)
}
func (m *mockService) AcceptQuote(id uuid.UUID) error       { return m.acceptFn(id) }
func (m *mockService) RejectQuote(id uuid.UUID) error       { return m.rejectFn(id) }
func (m *mockService) ExpireQuote(id uuid.UUID) error       { return m.expireFn(id) }
func (m *mockService) CancelQuote(id uuid.UUID) error       { return m.cancelFn(id) }
func (m *mockService) UpdateNotes(id uuid.UUID, notes string) error { return m.updateNotesFn(id, notes) }

// --- helpers ---

func sampleQuote() *domain.Quote {
	return &domain.Quote{
		ID:        uuid.New(),
		OrgID:     uuid.New(),
		Title:     "Office Supplies Q4",
		Status:    domain.QuoteStatusDraft,
		Currency:  "USD",
		CreatedBy: "user-1",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Items: domain.QuoteItems{
			{ProductID: "p1", SKU: "SKU-001", Quantity: 10},
		},
	}
}

func doRequest(h http.Handler, method, path string, body interface{}) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	return rr
}

// --- tests ---

func TestHealthz(t *testing.T) {
	h := handler.New(&mockService{})
	rr := doRequest(h, http.MethodGet, "/healthz", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var resp map[string]string
	_ = json.NewDecoder(rr.Body).Decode(&resp)
	if resp["status"] != "ok" {
		t.Errorf("expected status=ok, got %q", resp["status"])
	}
}

func TestCreateQuote_Success(t *testing.T) {
	q := sampleQuote()
	svc := &mockService{
		createFn: func(req service.CreateRFQRequest) (*domain.Quote, error) { return q, nil },
	}
	h := handler.New(svc)
	body := map[string]interface{}{
		"org_id":     q.OrgID.String(),
		"title":      q.Title,
		"currency":   "USD",
		"created_by": "user-1",
		"items":      []map[string]interface{}{{"product_id": "p1", "sku": "SKU-001", "quantity": 10}},
	}
	rr := doRequest(h, http.MethodPost, "/quotes", body)
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestCreateQuote_BadBody(t *testing.T) {
	h := handler.New(&mockService{
		createFn: func(req service.CreateRFQRequest) (*domain.Quote, error) {
			return nil, fmt.Errorf("title is required")
		},
	})
	rr := doRequest(h, http.MethodPost, "/quotes", map[string]interface{}{"title": ""})
	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", rr.Code)
	}
}

func TestGetQuote_Found(t *testing.T) {
	q := sampleQuote()
	svc := &mockService{
		getFn: func(id uuid.UUID) (*domain.Quote, error) { return q, nil },
	}
	h := handler.New(svc)
	rr := doRequest(h, http.MethodGet, "/quotes/"+q.ID.String(), nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestGetQuote_NotFound(t *testing.T) {
	svc := &mockService{
		getFn: func(id uuid.UUID) (*domain.Quote, error) { return nil, domain.ErrNotFound },
	}
	h := handler.New(svc)
	rr := doRequest(h, http.MethodGet, "/quotes/"+uuid.New().String(), nil)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestListQuotes_NoFilter(t *testing.T) {
	q := sampleQuote()
	svc := &mockService{
		listFn: func(*uuid.UUID, *domain.QuoteStatus) ([]*domain.Quote, error) {
			return []*domain.Quote{q}, nil
		},
	}
	h := handler.New(svc)
	rr := doRequest(h, http.MethodGet, "/quotes", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestSubmitRFQ_Success(t *testing.T) {
	svc := &mockService{
		submitFn: func(id uuid.UUID) error { return nil },
	}
	h := handler.New(svc)
	rr := doRequest(h, http.MethodPost, "/quotes/"+uuid.New().String()+"/submit", nil)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}
}

func TestSubmitRFQ_InvalidTransition(t *testing.T) {
	svc := &mockService{
		submitFn: func(id uuid.UUID) error { return domain.ErrInvalidTransition },
	}
	h := handler.New(svc)
	rr := doRequest(h, http.MethodPost, "/quotes/"+uuid.New().String()+"/submit", nil)
	if rr.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rr.Code)
	}
}

func TestProvideQuote_Success(t *testing.T) {
	q := sampleQuote()
	q.Status = domain.QuoteStatusQuoted
	svc := &mockService{
		provideFn: func(id uuid.UUID, req service.ProvideQuoteRequest) (*domain.Quote, error) {
			return q, nil
		},
	}
	h := handler.New(svc)
	body := map[string]interface{}{
		"total_amount": 500.0,
		"items":        []map[string]interface{}{{"product_id": "p1", "sku": "SKU-001", "quantity": 10}},
	}
	rr := doRequest(h, http.MethodPost, "/quotes/"+uuid.New().String()+"/provide", body)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestAcceptQuote_Success(t *testing.T) {
	svc := &mockService{
		acceptFn: func(id uuid.UUID) error { return nil },
	}
	h := handler.New(svc)
	rr := doRequest(h, http.MethodPost, "/quotes/"+uuid.New().String()+"/accept", nil)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}
}

func TestCancelQuote_Success(t *testing.T) {
	svc := &mockService{
		cancelFn: func(id uuid.UUID) error { return nil },
	}
	h := handler.New(svc)
	rr := doRequest(h, http.MethodDelete, "/quotes/"+uuid.New().String(), nil)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}
}
