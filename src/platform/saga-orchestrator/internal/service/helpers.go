package service

import "github.com/shopos/saga-orchestrator/internal/domain"

// StartRequest is a convenience constructor for domain.StartSagaRequest.
func StartRequest(orderID string, payload map[string]string) domain.StartSagaRequest {
	return domain.StartSagaRequest{
		Type:    domain.TypeOrderFulfillment,
		OrderID: orderID,
		Payload: payload,
	}
}
