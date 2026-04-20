package domain

import (
	"errors"
	"time"
)

// ReturnAuthStatus enumerates every state a return authorisation can be in.
type ReturnAuthStatus string

const (
	StatusPending    ReturnAuthStatus = "PENDING"
	StatusApproved   ReturnAuthStatus = "APPROVED"
	StatusRejected   ReturnAuthStatus = "REJECTED"
	StatusLabelIssued ReturnAuthStatus = "LABEL_ISSUED"
	StatusInTransit  ReturnAuthStatus = "IN_TRANSIT"
	StatusReceived   ReturnAuthStatus = "RECEIVED"
	StatusInspecting ReturnAuthStatus = "INSPECTING"
	StatusCompleted  ReturnAuthStatus = "COMPLETED"
	StatusCancelled  ReturnAuthStatus = "CANCELLED"
)

// ReturnItem represents a single line item within a return authorisation.
type ReturnItem struct {
	ProductID string `json:"productId"`
	SKU       string `json:"sku"`
	Quantity  int    `json:"quantity"`
	Condition string `json:"condition"` // e.g. "NEW", "USED", "DAMAGED"
}

// ReturnAuth is the aggregate root for a return authorisation.
type ReturnAuth struct {
	ID               string           `json:"id"`
	OrderID          string           `json:"orderId"`
	CustomerID       string           `json:"customerId"`
	Items            []ReturnItem     `json:"items"`
	Reason           string           `json:"reason"`
	Status           ReturnAuthStatus `json:"status"`
	ReturnLabel      string           `json:"returnLabel,omitempty"`
	TrackingNumber   string           `json:"trackingNumber,omitempty"`
	WarehouseID      string           `json:"warehouseId,omitempty"`
	InspectionNotes  string           `json:"inspectionNotes,omitempty"`
	RejectionReason  string           `json:"rejectionReason,omitempty"`
	CreatedAt        time.Time        `json:"createdAt"`
	UpdatedAt        time.Time        `json:"updatedAt"`
}

// validTransitions maps the current status to the set of allowed next statuses.
var validTransitions = map[ReturnAuthStatus][]ReturnAuthStatus{
	StatusPending:     {StatusApproved, StatusRejected, StatusCancelled},
	StatusApproved:    {StatusLabelIssued, StatusCancelled},
	StatusRejected:    {},
	StatusLabelIssued: {StatusInTransit, StatusCancelled},
	StatusInTransit:   {StatusReceived},
	StatusReceived:    {StatusInspecting},
	StatusInspecting:  {StatusCompleted},
	StatusCompleted:   {},
	StatusCancelled:   {},
}

// CanTransition returns true when moving from current to next is allowed.
func CanTransition(current, next ReturnAuthStatus) bool {
	allowed, ok := validTransitions[current]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == next {
			return true
		}
	}
	return false
}

// Sentinel errors for the returns domain.
var (
	ErrNotFound          = errors.New("return authorisation not found")
	ErrInvalidTransition = errors.New("invalid status transition")
	ErrInvalidRequest    = errors.New("invalid request")
)
