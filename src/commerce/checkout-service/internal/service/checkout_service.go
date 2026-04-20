package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/shopos/checkout-service/internal/domain"
)

// Doer abstracts the HTTP client so it can be mocked in tests.
type Doer interface {
	Get(ctx context.Context, url string) ([]byte, error)
	Post(ctx context.Context, url string, body []byte) ([]byte, error)
}

// Servicer is the interface that the HTTP handler depends on.
type Servicer interface {
	Initiate(ctx context.Context, req domain.InitiateRequest) (*domain.CheckoutSession, error)
	Confirm(ctx context.Context, req domain.ConfirmRequest) (*domain.CheckoutSession, error)
	GetSession(ctx context.Context, id string) (*domain.CheckoutSession, error)
	CancelSession(ctx context.Context, id string) error
}

// CheckoutService implements Servicer.
type CheckoutService struct {
	http     Doer
	cartURL  string
	taxURL   string
	orderURL string

	mu       sync.RWMutex
	sessions map[string]*domain.CheckoutSession
}

// New creates a CheckoutService wired to the given downstream URLs.
func New(doer Doer, cartURL, taxURL, orderURL string) *CheckoutService {
	return &CheckoutService{
		http:     doer,
		cartURL:  cartURL,
		taxURL:   taxURL,
		orderURL: orderURL,
		sessions: make(map[string]*domain.CheckoutSession),
	}
}

// ─── downstream response shapes ───────────────────────────────────────────────

// cartItem mirrors the response shape from cart-service.
type cartItem struct {
	ProductID string  `json:"product_id"`
	SKU       string  `json:"sku"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Quantity  int     `json:"quantity"`
}

// cartResponse mirrors the cart-service GET /carts/{id} response.
type cartResponse struct {
	ID       string     `json:"id"`
	UserID   string     `json:"user_id"`
	Items    []cartItem `json:"items"`
	Currency string     `json:"currency"`
}

// taxRequest mirrors the tax-service POST /tax/calculate request.
type taxRequest struct {
	Items        []taxLineItem `json:"items"`
	ShipTo       taxAddress    `json:"ship_to"`
	Currency     string        `json:"currency"`
	CustomerType string        `json:"customer_type"`
}

type taxLineItem struct {
	ProductID string  `json:"product_id"`
	Category  string  `json:"category"`
	Amount    float64 `json:"amount"`
	Quantity  int     `json:"quantity"`
}

type taxAddress struct {
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
}

// taxResponse mirrors the tax-service POST /tax/calculate response.
type taxResponse struct {
	TaxAmount float64 `json:"tax_amount"`
	Total     float64 `json:"total"`
}

// orderRequest mirrors the order-service POST /orders request.
type orderRequest struct {
	UserID          string              `json:"user_id"`
	CartID          string              `json:"cart_id"`
	SessionID       string              `json:"session_id"`
	Items           []domain.CheckoutItem `json:"items"`
	ShippingAddr    domain.Address      `json:"shipping_address"`
	BillingAddr     domain.Address      `json:"billing_address"`
	PaymentMethodID string              `json:"payment_method_id"`
	Subtotal        float64             `json:"subtotal"`
	Tax             float64             `json:"tax"`
	Shipping        float64             `json:"shipping"`
	Total           float64             `json:"total"`
	Currency        string              `json:"currency"`
}

// orderResponse mirrors the order-service POST /orders response.
type orderResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// ─── Initiate ─────────────────────────────────────────────────────────────────

// Initiate opens a new checkout session. It:
//  1. Fetches the cart from cart-service
//  2. Calls tax-service to compute taxes
//  3. Applies a fixed shipping fee
//  4. Persists the session in-memory and returns it
func (s *CheckoutService) Initiate(ctx context.Context, req domain.InitiateRequest) (*domain.CheckoutSession, error) {
	// 1. Fetch cart
	cartData, err := s.http.Get(ctx, fmt.Sprintf("%s/carts/%s", s.cartURL, req.CartID))
	if err != nil {
		return nil, fmt.Errorf("checkout_service: fetch cart: %w", err)
	}

	var cart cartResponse
	if err := json.Unmarshal(cartData, &cart); err != nil {
		return nil, fmt.Errorf("checkout_service: decode cart: %w", err)
	}
	if len(cart.Items) == 0 {
		return nil, fmt.Errorf("checkout_service: cart %s is empty", req.CartID)
	}

	// Build domain items and subtotal
	items := make([]domain.CheckoutItem, 0, len(cart.Items))
	var subtotal float64
	for _, ci := range cart.Items {
		items = append(items, domain.CheckoutItem{
			ProductID: ci.ProductID,
			SKU:       ci.SKU,
			Name:      ci.Name,
			Price:     ci.Price,
			Quantity:  ci.Quantity,
		})
		subtotal += ci.Price * float64(ci.Quantity)
	}

	currency := req.Currency
	if currency == "" {
		currency = cart.Currency
	}
	if currency == "" {
		currency = "USD"
	}

	// 2. Calculate tax
	taxLineItems := make([]taxLineItem, 0, len(items))
	for _, it := range items {
		taxLineItems = append(taxLineItems, taxLineItem{
			ProductID: it.ProductID,
			Category:  "general",
			Amount:    it.Price,
			Quantity:  it.Quantity,
		})
	}
	taxReqBody, err := json.Marshal(taxRequest{
		Items: taxLineItems,
		ShipTo: taxAddress{
			City:       req.ShippingAddr.City,
			State:      req.ShippingAddr.State,
			PostalCode: req.ShippingAddr.PostalCode,
			Country:    req.ShippingAddr.Country,
		},
		Currency:     currency,
		CustomerType: "b2c",
	})
	if err != nil {
		return nil, fmt.Errorf("checkout_service: marshal tax request: %w", err)
	}

	taxData, err := s.http.Post(ctx, fmt.Sprintf("%s/tax/calculate", s.taxURL), taxReqBody)
	if err != nil {
		return nil, fmt.Errorf("checkout_service: call tax-service: %w", err)
	}

	var taxResp taxResponse
	if err := json.Unmarshal(taxData, &taxResp); err != nil {
		return nil, fmt.Errorf("checkout_service: decode tax response: %w", err)
	}

	// 3. Apply shipping fee (flat rate based on country)
	shippingFee := shippingFee(req.ShippingAddr.Country)

	now := time.Now().UTC()
	session := &domain.CheckoutSession{
		ID:           uuid.New().String(),
		UserID:       req.UserID,
		CartID:       req.CartID,
		Items:        items,
		ShippingAddr: req.ShippingAddr,
		BillingAddr:  req.BillingAddr,
		Subtotal:     round2(subtotal),
		Tax:          round2(taxResp.TaxAmount),
		Shipping:     round2(shippingFee),
		Total:        round2(subtotal + taxResp.TaxAmount + shippingFee),
		Currency:     currency,
		Status:       domain.StatusPending,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	s.mu.Lock()
	s.sessions[session.ID] = session
	s.mu.Unlock()

	return session, nil
}

// ─── Confirm ──────────────────────────────────────────────────────────────────

// Confirm submits the pending session to order-service, creating an order.
func (s *CheckoutService) Confirm(ctx context.Context, req domain.ConfirmRequest) (*domain.CheckoutSession, error) {
	s.mu.RLock()
	session, ok := s.sessions[req.SessionID]
	s.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("checkout_service: session %s not found", req.SessionID)
	}
	if session.Status != domain.StatusPending {
		return nil, fmt.Errorf("checkout_service: session %s is not pending (status=%s)", req.SessionID, session.Status)
	}

	orderReqBody, err := json.Marshal(orderRequest{
		UserID:          session.UserID,
		CartID:          session.CartID,
		SessionID:       session.ID,
		Items:           session.Items,
		ShippingAddr:    session.ShippingAddr,
		BillingAddr:     session.BillingAddr,
		PaymentMethodID: req.PaymentMethodID,
		Subtotal:        session.Subtotal,
		Tax:             session.Tax,
		Shipping:        session.Shipping,
		Total:           session.Total,
		Currency:        session.Currency,
	})
	if err != nil {
		return nil, fmt.Errorf("checkout_service: marshal order request: %w", err)
	}

	orderData, err := s.http.Post(ctx, fmt.Sprintf("%s/orders", s.orderURL), orderReqBody)
	if err != nil {
		s.mu.Lock()
		session.Status = domain.StatusFailed
		session.UpdatedAt = time.Now().UTC()
		s.mu.Unlock()
		return nil, fmt.Errorf("checkout_service: call order-service: %w", err)
	}

	var orderResp orderResponse
	if err := json.Unmarshal(orderData, &orderResp); err != nil {
		return nil, fmt.Errorf("checkout_service: decode order response: %w", err)
	}

	s.mu.Lock()
	session.Status = domain.StatusConfirmed
	session.OrderID = orderResp.ID
	session.UpdatedAt = time.Now().UTC()
	// Take a value copy while still holding the lock, then return it.
	confirmed := *session
	s.mu.Unlock()

	return &confirmed, nil
}

// ─── GetSession ───────────────────────────────────────────────────────────────

// GetSession returns the session with the given ID, or an error if not found.
func (s *CheckoutService) GetSession(_ context.Context, id string) (*domain.CheckoutSession, error) {
	s.mu.RLock()
	session, ok := s.sessions[id]
	s.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("checkout_service: session %s not found", id)
	}
	copy := *session
	return &copy, nil
}

// ─── CancelSession ────────────────────────────────────────────────────────────

// CancelSession marks the session as cancelled. Only pending sessions may be cancelled.
func (s *CheckoutService) CancelSession(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, ok := s.sessions[id]
	if !ok {
		return fmt.Errorf("checkout_service: session %s not found", id)
	}
	if session.Status == domain.StatusConfirmed {
		return fmt.Errorf("checkout_service: cannot cancel a confirmed session (%s)", id)
	}
	session.Status = domain.StatusCancelled
	session.UpdatedAt = time.Now().UTC()
	return nil
}

// ─── helpers ─────────────────────────────────────────────────────────────────

// shippingFee returns a flat shipping fee (USD) by destination country.
func shippingFee(country string) float64 {
	fees := map[string]float64{
		"US": 5.99,
		"CA": 8.99,
		"GB": 7.99,
		"AU": 12.99,
		"DE": 9.99,
		"FR": 9.99,
	}
	if fee, ok := fees[country]; ok {
		return fee
	}
	return 14.99 // international default
}

// round2 rounds f to two decimal places.
func round2(f float64) float64 {
	// Use integer arithmetic to avoid floating-point oddities.
	v := int64(f*100 + 0.5)
	return float64(v) / 100
}

