package domain

import "time"

// NotifType represents the category of an in-app notification.
type NotifType string

const (
	NotifTypeOrderUpdate NotifType = "ORDER_UPDATE"
	NotifTypePayment     NotifType = "PAYMENT"
	NotifTypePromotion   NotifType = "PROMOTION"
	NotifTypeSystem      NotifType = "SYSTEM"
	NotifTypeAlert       NotifType = "ALERT"
)

// Notification is a single in-app notification belonging to a user.
type Notification struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	Type      NotifType `json:"type"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	Link      string    `json:"link,omitempty"`
	Read      bool      `json:"read"`
	CreatedAt time.Time `json:"createdAt"`
}

// NotifPage is a paginated result set of notifications with aggregate counts.
type NotifPage struct {
	Notifications []Notification `json:"notifications"`
	Total         int            `json:"total"`
	Unread        int            `json:"unread"`
}

// SendRequest carries the fields required to create a new notification.
type SendRequest struct {
	Type  NotifType `json:"type"`
	Title string    `json:"title"`
	Body  string    `json:"body"`
	Link  string    `json:"link,omitempty"`
}

// Validate returns an error string if the request is malformed.
func (r *SendRequest) Validate() string {
	if r.Title == "" {
		return "title is required"
	}
	if r.Body == "" {
		return "body is required"
	}
	switch r.Type {
	case NotifTypeOrderUpdate, NotifTypePayment, NotifTypePromotion, NotifTypeSystem, NotifTypeAlert:
		// valid
	default:
		return "type must be one of ORDER_UPDATE, PAYMENT, PROMOTION, SYSTEM, ALERT"
	}
	return ""
}
