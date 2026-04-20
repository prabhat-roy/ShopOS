package domain

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// QuoteStatus represents the lifecycle state of a quote/RFQ.
type QuoteStatus string

const (
	QuoteStatusDraft       QuoteStatus = "DRAFT"
	QuoteStatusSubmitted   QuoteStatus = "SUBMITTED"
	QuoteStatusUnderReview QuoteStatus = "UNDER_REVIEW"
	QuoteStatusQuoted      QuoteStatus = "QUOTED"
	QuoteStatusAccepted    QuoteStatus = "ACCEPTED"
	QuoteStatusRejected    QuoteStatus = "REJECTED"
	QuoteStatusExpired     QuoteStatus = "EXPIRED"
	QuoteStatusCancelled   QuoteStatus = "CANCELLED"
)

// Sentinel errors.
var (
	ErrNotFound          = errors.New("quote not found")
	ErrInvalidTransition = errors.New("invalid status transition")
)

// validTransitions defines allowed from→to status transitions.
var validTransitions = map[QuoteStatus][]QuoteStatus{
	QuoteStatusDraft:       {QuoteStatusSubmitted, QuoteStatusCancelled},
	QuoteStatusSubmitted:   {QuoteStatusUnderReview, QuoteStatusRejected, QuoteStatusCancelled},
	QuoteStatusUnderReview: {QuoteStatusQuoted, QuoteStatusRejected, QuoteStatusCancelled},
	QuoteStatusQuoted:      {QuoteStatusAccepted, QuoteStatusRejected, QuoteStatusExpired, QuoteStatusCancelled},
	QuoteStatusAccepted:    {},
	QuoteStatusRejected:    {},
	QuoteStatusExpired:     {},
	QuoteStatusCancelled:   {},
}

// CanTransition returns true if the transition from current to next is allowed.
func CanTransition(current, next QuoteStatus) bool {
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

// QuoteItem represents a single line item in a quote request.
type QuoteItem struct {
	ProductID      string   `json:"product_id"`
	SKU            string   `json:"sku"`
	Quantity       int      `json:"quantity"`
	RequestedPrice *float64 `json:"requested_price,omitempty"`
	OfferedPrice   *float64 `json:"offered_price,omitempty"`
	Notes          string   `json:"notes,omitempty"`
}

// QuoteItems is a slice of QuoteItem that implements the sql.Scanner / driver.Valuer
// interfaces so it can be stored as JSONB in PostgreSQL.
type QuoteItems []QuoteItem

// Value implements driver.Valuer for database/sql.
func (qi QuoteItems) Value() (driver.Value, error) {
	if qi == nil {
		return "[]", nil
	}
	b, err := json.Marshal(qi)
	if err != nil {
		return nil, fmt.Errorf("marshal QuoteItems: %w", err)
	}
	return string(b), nil
}

// Scan implements sql.Scanner for database/sql.
func (qi *QuoteItems) Scan(src interface{}) error {
	var data []byte
	switch v := src.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	case nil:
		*qi = QuoteItems{}
		return nil
	default:
		return fmt.Errorf("unsupported type for QuoteItems: %T", src)
	}
	return json.Unmarshal(data, qi)
}

// Quote is the aggregate root for an RFQ/quote lifecycle.
type Quote struct {
	ID                uuid.UUID  `json:"id"`
	OrgID             uuid.UUID  `json:"org_id"`
	Title             string     `json:"title"`
	Description       string     `json:"description"`
	Items             QuoteItems `json:"items"`
	RequestedDelivery time.Time  `json:"requested_delivery"`
	Status            QuoteStatus `json:"status"`
	TotalAmount       float64    `json:"total_amount"`
	Currency          string     `json:"currency"`
	ValidUntil        *time.Time `json:"valid_until,omitempty"`
	Notes             string     `json:"notes,omitempty"`
	CreatedBy         string     `json:"created_by"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// Transition moves the quote to the next status, returning ErrInvalidTransition if not allowed.
func (q *Quote) Transition(next QuoteStatus) error {
	if !CanTransition(q.Status, next) {
		return fmt.Errorf("%w: %s → %s", ErrInvalidTransition, q.Status, next)
	}
	q.Status = next
	q.UpdatedAt = time.Now().UTC()
	return nil
}
