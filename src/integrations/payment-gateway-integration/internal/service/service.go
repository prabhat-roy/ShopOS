package service

import (
	"fmt"
	"time"

	"github.com/shopos/payment-gateway-integration/internal/domain"
	"github.com/shopos/payment-gateway-integration/internal/gateway"
	"github.com/shopos/payment-gateway-integration/internal/store"
)

// Servicer encapsulates payment gateway integration business logic.
type Servicer struct {
	store   *store.PaymentStore
	adapter *gateway.GatewayAdapter
}

// New constructs a Servicer.
func New(st *store.PaymentStore, ad *gateway.GatewayAdapter) *Servicer {
	return &Servicer{store: st, adapter: ad}
}

// CreateCharge submits a charge request through the appropriate gateway.
func (s *Servicer) CreateCharge(req domain.ChargeRequest) (domain.ChargeResponse, error) {
	if !domain.ValidGateway(req.Gateway) {
		return domain.ChargeResponse{}, fmt.Errorf("unsupported gateway: %s", req.Gateway)
	}
	if req.Amount <= 0 {
		return domain.ChargeResponse{}, fmt.Errorf("amount must be greater than zero")
	}
	if req.Currency == "" {
		return domain.ChargeResponse{}, fmt.Errorf("currency is required")
	}
	if req.OrderID == "" {
		return domain.ChargeResponse{}, fmt.Errorf("orderId is required")
	}
	if req.CustomerID == "" {
		return domain.ChargeResponse{}, fmt.Errorf("customerId is required")
	}
	if req.PaymentMethodToken == "" {
		return domain.ChargeResponse{}, fmt.Errorf("paymentMethodToken is required")
	}

	resp, err := s.adapter.Charge(req)
	if err != nil {
		return domain.ChargeResponse{}, err
	}

	// Persist intent.
	pi := &domain.PaymentIntent{
		ID:               resp.PaymentIntentID,
		Gateway:          req.Gateway,
		Amount:           req.Amount,
		Currency:         resp.Currency,
		CustomerID:       req.CustomerID,
		OrderID:          req.OrderID,
		Status:           resp.Status,
		GatewayPaymentID: resp.GatewayPaymentID,
		Metadata:         req.Metadata,
		CreatedAt:        resp.CreatedAt,
	}
	s.store.SaveIntent(pi)

	return resp, nil
}

// GetPayment retrieves a PaymentIntent by its ShopOS ID.
func (s *Servicer) GetPayment(id string) (*domain.PaymentIntent, error) {
	return s.store.GetIntent(id)
}

// ListOrderPayments returns all payment intents for a given order.
func (s *Servicer) ListOrderPayments(orderID string) []*domain.PaymentIntent {
	return s.store.ListByOrderId(orderID)
}

// ListCustomerPayments returns all payment intents for a given customer.
func (s *Servicer) ListCustomerPayments(customerID string) []*domain.PaymentIntent {
	return s.store.ListByCustomerId(customerID)
}

// CreateRefund processes a refund against an existing payment intent.
func (s *Servicer) CreateRefund(req domain.RefundRequest) (domain.RefundResponse, error) {
	if req.PaymentIntentID == "" {
		return domain.RefundResponse{}, fmt.Errorf("paymentIntentId is required")
	}
	if req.Amount <= 0 {
		return domain.RefundResponse{}, fmt.Errorf("refund amount must be greater than zero")
	}

	// Verify the intent exists.
	pi, err := s.store.GetIntent(req.PaymentIntentID)
	if err != nil {
		return domain.RefundResponse{}, fmt.Errorf("payment intent not found: %w", err)
	}
	if req.Amount > pi.Amount {
		return domain.RefundResponse{}, fmt.Errorf("refund amount (%.2f) exceeds original charge (%.2f)", req.Amount, pi.Amount)
	}

	resp, err := s.adapter.Refund(req)
	if err != nil {
		return domain.RefundResponse{}, err
	}

	s.store.SaveRefund(&resp)

	// Update intent status.
	if req.Amount >= pi.Amount {
		pi.Status = "refunded"
	} else {
		pi.Status = "partially_refunded"
	}
	if pi.Metadata == nil {
		pi.Metadata = make(map[string]string)
	}
	pi.Metadata["refund_id"] = resp.RefundID
	pi.Metadata["refund_at"] = resp.ProcessedAt.Format(time.RFC3339)
	s.store.SaveIntent(pi)

	return resp, nil
}

// GetRefund retrieves a RefundResponse by its ID.
func (s *Servicer) GetRefund(refundID string) (*domain.RefundResponse, error) {
	return s.store.GetRefund(refundID)
}

// GetSupportedGateways returns the list of supported gateways.
func (s *Servicer) GetSupportedGateways() []domain.Gateway {
	return s.adapter.GetSupportedGateways()
}
