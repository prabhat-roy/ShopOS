package domain

import "time"

// Status constants for a CheckoutSession.
const (
	StatusPending   = "pending"
	StatusConfirmed = "confirmed"
	StatusFailed    = "failed"
	StatusCancelled = "cancelled"
)

// CheckoutSession is the aggregate root for a single checkout flow.
type CheckoutSession struct {
	ID           string         `json:"id"`
	UserID       string         `json:"user_id"`
	CartID       string         `json:"cart_id"`
	Items        []CheckoutItem `json:"items"`
	ShippingAddr Address        `json:"shipping_address"`
	BillingAddr  Address        `json:"billing_address"`
	Subtotal     float64        `json:"subtotal"`
	Tax          float64        `json:"tax"`
	Shipping     float64        `json:"shipping"`
	Total        float64        `json:"total"`
	Currency     string         `json:"currency"`
	Status       string         `json:"status"`
	OrderID      string         `json:"order_id,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

// CheckoutItem represents a single line item within a checkout session.
type CheckoutItem struct {
	ProductID string  `json:"product_id"`
	SKU       string  `json:"sku"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Quantity  int     `json:"quantity"`
}

// Address is a generic postal address used for shipping and billing.
type Address struct {
	Line1      string `json:"line1"`
	Line2      string `json:"line2,omitempty"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
}

// InitiateRequest carries the data needed to open a new checkout session.
type InitiateRequest struct {
	UserID       string  `json:"user_id"`
	CartID       string  `json:"cart_id"`
	ShippingAddr Address `json:"shipping_address"`
	BillingAddr  Address `json:"billing_address"`
	Currency     string  `json:"currency"`
}

// ConfirmRequest carries the data needed to confirm (pay for) a checkout session.
type ConfirmRequest struct {
	SessionID       string `json:"session_id"`
	PaymentMethodID string `json:"payment_method_id"`
}

// ErrorResponse is the JSON shape returned on error.
type ErrorResponse struct {
	Error string `json:"error"`
}
