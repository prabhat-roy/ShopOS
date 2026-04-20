package domain

import "time"

// Gateway represents a supported payment gateway.
type Gateway string

const (
	GatewayStripe     Gateway = "STRIPE"
	GatewayPayPal     Gateway = "PAYPAL"
	GatewayBraintree  Gateway = "BRAINTREE"
	GatewayAdyen      Gateway = "ADYEN"
)

// ValidGateway returns true when g is a recognised gateway.
func ValidGateway(g Gateway) bool {
	switch g {
	case GatewayStripe, GatewayPayPal, GatewayBraintree, GatewayAdyen:
		return true
	}
	return false
}

// PaymentIntent represents a payment transaction record in ShopOS.
type PaymentIntent struct {
	ID               string            `json:"id"`
	Gateway          Gateway           `json:"gateway"`
	Amount           float64           `json:"amount"`
	Currency         string            `json:"currency"`
	CustomerID       string            `json:"customerId"`
	OrderID          string            `json:"orderId"`
	Status           string            `json:"status"`
	GatewayPaymentID string            `json:"gatewayPaymentId"`
	Metadata         map[string]string `json:"metadata,omitempty"`
	CreatedAt        time.Time         `json:"createdAt"`
}

// ChargeRequest is the input payload to create a charge.
type ChargeRequest struct {
	Gateway             Gateway           `json:"gateway"`
	Amount              float64           `json:"amount"`
	Currency            string            `json:"currency"`
	CustomerID          string            `json:"customerId"`
	OrderID             string            `json:"orderId"`
	PaymentMethodToken  string            `json:"paymentMethodToken"`
	Metadata            map[string]string `json:"metadata,omitempty"`
}

// ChargeResponse is the result of a charge operation.
type ChargeResponse struct {
	PaymentIntentID  string    `json:"paymentIntentId"`
	GatewayPaymentID string    `json:"gatewayPaymentId"`
	Status           string    `json:"status"`
	Gateway          Gateway   `json:"gateway"`
	Amount           float64   `json:"amount"`
	Currency         string    `json:"currency"`
	CreatedAt        time.Time `json:"createdAt"`
}

// RefundRequest is the input payload to create a refund.
type RefundRequest struct {
	PaymentIntentID string  `json:"paymentIntentId"`
	Amount          float64 `json:"amount"`
	Reason          string  `json:"reason"`
}

// RefundResponse is the result of a refund operation.
type RefundResponse struct {
	RefundID        string    `json:"refundId"`
	PaymentIntentID string    `json:"paymentIntentId"`
	Amount          float64   `json:"amount"`
	Status          string    `json:"status"`
	ProcessedAt     time.Time `json:"processedAt"`
}
