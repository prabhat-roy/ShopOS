package gateway

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopos/payment-gateway-integration/internal/domain"
)

// GatewayAdapter simulates calls to external payment gateways.
type GatewayAdapter struct {
	rng *rand.Rand
}

// New returns a new GatewayAdapter.
func New() *GatewayAdapter {
	return &GatewayAdapter{
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Charge simulates a gateway charge call and returns a ChargeResponse.
// It uses a 95 % success rate to simulate occasional gateway failures.
func (a *GatewayAdapter) Charge(req domain.ChargeRequest) (domain.ChargeResponse, error) {
	// Simulate 5 % failure rate.
	if a.rng.Intn(100) < 5 {
		return domain.ChargeResponse{}, fmt.Errorf("gateway %s: payment declined (simulated)", req.Gateway)
	}

	gatewayPaymentID := a.generateGatewayPaymentID(req.Gateway)

	return domain.ChargeResponse{
		PaymentIntentID:  uuid.New().String(),
		GatewayPaymentID: gatewayPaymentID,
		Status:           "succeeded",
		Gateway:          req.Gateway,
		Amount:           req.Amount,
		Currency:         strings.ToUpper(req.Currency),
		CreatedAt:        time.Now().UTC(),
	}, nil
}

// Refund simulates a gateway refund call and returns a RefundResponse.
func (a *GatewayAdapter) Refund(req domain.RefundRequest) (domain.RefundResponse, error) {
	if req.PaymentIntentID == "" {
		return domain.RefundResponse{}, fmt.Errorf("paymentIntentId is required")
	}
	if req.Amount <= 0 {
		return domain.RefundResponse{}, fmt.Errorf("refund amount must be greater than zero")
	}

	return domain.RefundResponse{
		RefundID:        uuid.New().String(),
		PaymentIntentID: req.PaymentIntentID,
		Amount:          req.Amount,
		Status:          "succeeded",
		ProcessedAt:     time.Now().UTC(),
	}, nil
}

// FormatRequest converts a ChargeRequest into gateway-specific field names.
func (a *GatewayAdapter) FormatRequest(gw domain.Gateway, req domain.ChargeRequest) map[string]interface{} {
	switch gw {
	case domain.GatewayStripe:
		return map[string]interface{}{
			"amount":               int(req.Amount * 100), // Stripe uses smallest currency unit
			"currency":             strings.ToLower(req.Currency),
			"payment_method":       req.PaymentMethodToken,
			"customer":             req.CustomerID,
			"metadata":             req.Metadata,
			"confirm":              true,
			"return_url":           "https://shopos.io/return",
		}
	case domain.GatewayPayPal:
		return map[string]interface{}{
			"intent": "CAPTURE",
			"purchase_units": []map[string]interface{}{
				{
					"amount": map[string]interface{}{
						"currency_code": strings.ToUpper(req.Currency),
						"value":         fmt.Sprintf("%.2f", req.Amount),
					},
					"custom_id": req.OrderID,
				},
			},
			"payment_source": map[string]interface{}{
				"token": map[string]interface{}{
					"id":   req.PaymentMethodToken,
					"type": "BILLING_AGREEMENT",
				},
			},
		}
	case domain.GatewayBraintree:
		return map[string]interface{}{
			"amount":             fmt.Sprintf("%.2f", req.Amount),
			"paymentMethodNonce": req.PaymentMethodToken,
			"orderId":            req.OrderID,
			"customerId":         req.CustomerID,
			"options": map[string]interface{}{
				"submitForSettlement": true,
			},
		}
	case domain.GatewayAdyen:
		return map[string]interface{}{
			"merchantAccount": "ShopOsMerchant",
			"amount": map[string]interface{}{
				"currency": strings.ToUpper(req.Currency),
				"value":    int(req.Amount * 100),
			},
			"reference":       req.OrderID,
			"paymentMethod": map[string]interface{}{
				"type":  "scheme",
				"token": req.PaymentMethodToken,
			},
			"shopperReference": req.CustomerID,
		}
	}
	return map[string]interface{}{}
}

// GetSupportedGateways returns the list of gateways this adapter supports.
func (a *GatewayAdapter) GetSupportedGateways() []domain.Gateway {
	return []domain.Gateway{
		domain.GatewayStripe,
		domain.GatewayPayPal,
		domain.GatewayBraintree,
		domain.GatewayAdyen,
	}
}

// generateGatewayPaymentID returns a realistic-looking payment ID for each gateway.
func (a *GatewayAdapter) generateGatewayPaymentID(gw domain.Gateway) string {
	short := strings.ReplaceAll(uuid.New().String(), "-", "")[:16]
	switch gw {
	case domain.GatewayStripe:
		return "pi_" + short
	case domain.GatewayPayPal:
		return "PAYPAL-" + strings.ToUpper(short)
	case domain.GatewayBraintree:
		return "bt_" + short
	case domain.GatewayAdyen:
		return "psp_" + short
	}
	return "pay_" + short
}
