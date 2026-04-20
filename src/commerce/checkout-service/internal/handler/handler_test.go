package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/shopos/checkout-service/internal/domain"
	"github.com/shopos/checkout-service/internal/handler"
)

// ─── mock service ────────────────────────────────────────────────────────────

type mockService struct {
	sessions map[string]*domain.CheckoutSession
}

func newMockService() *mockService {
	return &mockService{sessions: make(map[string]*domain.CheckoutSession)}
}

func (m *mockService) Initiate(_ context.Context, req domain.InitiateRequest) (*domain.CheckoutSession, error) {
	s := &domain.CheckoutSession{
		ID:       "sess-001",
		UserID:   req.UserID,
		CartID:   req.CartID,
		Currency: req.Currency,
		Subtotal: 50.00,
		Tax:      3.63,
		Shipping: 5.99,
		Total:    59.62,
		Status:   domain.StatusPending,
	}
	m.sessions[s.ID] = s
	return s, nil
}

func (m *mockService) Confirm(_ context.Context, req domain.ConfirmRequest) (*domain.CheckoutSession, error) {
	s, ok := m.sessions[req.SessionID]
	if !ok {
		return nil, notFoundErr(req.SessionID)
	}
	s.Status = domain.StatusConfirmed
	s.OrderID = "ord-001"
	return s, nil
}

func (m *mockService) GetSession(_ context.Context, id string) (*domain.CheckoutSession, error) {
	s, ok := m.sessions[id]
	if !ok {
		return nil, notFoundErr(id)
	}
	return s, nil
}

func (m *mockService) CancelSession(_ context.Context, id string) error {
	if _, ok := m.sessions[id]; !ok {
		return notFoundErr(id)
	}
	m.sessions[id].Status = domain.StatusCancelled
	return nil
}

type notFoundError struct{ id string }

func notFoundErr(id string) error    { return &notFoundError{id: id} }
func (e *notFoundError) Error() string { return "session " + e.id + " not found" }

// ─── test helpers ─────────────────────────────────────────────────────────────

func setup(t *testing.T) (*handler.Handler, *mockService) {
	t.Helper()
	svc := newMockService()
	logger := log.New(os.Stderr, "[test] ", 0)
	h := handler.New(svc, logger)
	return h, svc
}

func toJSON(t *testing.T, v any) *bytes.Buffer {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("toJSON: %v", err)
	}
	return bytes.NewBuffer(b)
}

// ─── tests ───────────────────────────────────────────────────────────────────

func TestHealthz(t *testing.T) {
	h, _ := setup(t)
	r := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp map[string]string
	mustDecodeJSON(t, w.Body, &resp)
	if resp["status"] != "ok" {
		t.Errorf("expected status=ok, got %q", resp["status"])
	}
}

func TestHealthz_MethodNotAllowed(t *testing.T) {
	h, _ := setup(t)
	r := httptest.NewRequest(http.MethodPost, "/healthz", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestInitiateSession_Created(t *testing.T) {
	h, _ := setup(t)
	req := domain.InitiateRequest{
		UserID:   "user-1",
		CartID:   "cart-1",
		Currency: "USD",
		ShippingAddr: domain.Address{
			Line1: "123 Main St", City: "Los Angeles", State: "CA",
			PostalCode: "90001", Country: "US",
		},
		BillingAddr: domain.Address{
			Line1: "123 Main St", City: "Los Angeles", State: "CA",
			PostalCode: "90001", Country: "US",
		},
	}

	r := httptest.NewRequest(http.MethodPost, "/checkout/sessions", toJSON(t, req))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d; body: %s", w.Code, w.Body.String())
	}
	var sess domain.CheckoutSession
	mustDecodeJSON(t, w.Body, &sess)
	if sess.ID == "" {
		t.Error("session ID should not be empty")
	}
	if sess.Status != domain.StatusPending {
		t.Errorf("expected status=pending, got %q", sess.Status)
	}
}

func TestInitiateSession_MissingUserID(t *testing.T) {
	h, _ := setup(t)
	req := domain.InitiateRequest{CartID: "cart-1"}
	r := httptest.NewRequest(http.MethodPost, "/checkout/sessions", toJSON(t, req))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestInitiateSession_MissingCartID(t *testing.T) {
	h, _ := setup(t)
	req := domain.InitiateRequest{UserID: "user-1"}
	r := httptest.NewRequest(http.MethodPost, "/checkout/sessions", toJSON(t, req))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestInitiateSession_InvalidJSON(t *testing.T) {
	h, _ := setup(t)
	r := httptest.NewRequest(http.MethodPost, "/checkout/sessions", bytes.NewBufferString("{bad json"))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestConfirmSession_OK(t *testing.T) {
	h, svc := setup(t)

	// Pre-seed a session
	svc.sessions["sess-001"] = &domain.CheckoutSession{
		ID:     "sess-001",
		UserID: "user-1",
		Status: domain.StatusPending,
	}

	req := domain.ConfirmRequest{PaymentMethodID: "pm-stripe-001"}
	r := httptest.NewRequest(http.MethodPost, "/checkout/sessions/sess-001/confirm", toJSON(t, req))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}
	var sess domain.CheckoutSession
	mustDecodeJSON(t, w.Body, &sess)
	if sess.Status != domain.StatusConfirmed {
		t.Errorf("expected confirmed, got %q", sess.Status)
	}
	if sess.OrderID == "" {
		t.Error("order_id should be set after confirmation")
	}
}

func TestConfirmSession_MissingPaymentMethod(t *testing.T) {
	h, svc := setup(t)
	svc.sessions["sess-001"] = &domain.CheckoutSession{ID: "sess-001", Status: domain.StatusPending}

	req := domain.ConfirmRequest{} // no PaymentMethodID
	r := httptest.NewRequest(http.MethodPost, "/checkout/sessions/sess-001/confirm", toJSON(t, req))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestConfirmSession_NotFound(t *testing.T) {
	h, _ := setup(t)
	req := domain.ConfirmRequest{PaymentMethodID: "pm-001"}
	r := httptest.NewRequest(http.MethodPost, "/checkout/sessions/missing/confirm", toJSON(t, req))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestGetSession_OK(t *testing.T) {
	h, svc := setup(t)
	svc.sessions["sess-001"] = &domain.CheckoutSession{
		ID:     "sess-001",
		Status: domain.StatusPending,
	}
	r := httptest.NewRequest(http.MethodGet, "/checkout/sessions/sess-001", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var sess domain.CheckoutSession
	mustDecodeJSON(t, w.Body, &sess)
	if sess.ID != "sess-001" {
		t.Errorf("expected sess-001, got %q", sess.ID)
	}
}

func TestGetSession_NotFound(t *testing.T) {
	h, _ := setup(t)
	r := httptest.NewRequest(http.MethodGet, "/checkout/sessions/does-not-exist", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestCancelSession_OK(t *testing.T) {
	h, svc := setup(t)
	svc.sessions["sess-001"] = &domain.CheckoutSession{ID: "sess-001", Status: domain.StatusPending}

	r := httptest.NewRequest(http.MethodDelete, "/checkout/sessions/sess-001", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}

func TestCancelSession_NotFound(t *testing.T) {
	h, _ := setup(t)
	r := httptest.NewRequest(http.MethodDelete, "/checkout/sessions/ghost", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func mustDecodeJSON(t *testing.T, buf *bytes.Buffer, v any) {
	t.Helper()
	if err := json.NewDecoder(buf).Decode(v); err != nil {
		t.Fatalf("mustDecodeJSON: %v", err)
	}
}
