// Package domain defines the core data structures for push notifications.
package domain

import "time"

// Platform constants for the supported push delivery targets.
const (
	PlatformIOS     = "ios"
	PlatformAndroid = "android"
	PlatformWeb     = "web"
)

// PushMessage represents an inbound event payload consumed from Kafka.
type PushMessage struct {
	MessageID   string            `json:"messageId"`
	DeviceToken string            `json:"deviceToken"`
	UserID      string            `json:"userId"`
	Platform    string            `json:"platform"` // "ios" | "android" | "web"
	Title       string            `json:"title"`
	Body        string            `json:"body"`
	Data        map[string]string `json:"data,omitempty"`
}

// PushRecord is the persisted result of a processed push attempt.
type PushRecord struct {
	MessageID   string    `json:"messageId"`
	DeviceToken string    `json:"deviceToken"`
	Platform    string    `json:"platform"`
	Title       string    `json:"title"`
	Status      string    `json:"status"` // "delivered" | "failed"
	SentAt      time.Time `json:"sentAt"`
	ErrorMsg    string    `json:"errorMsg,omitempty"`
}

// PushStats holds aggregate delivery counters.
type PushStats struct {
	Sent      int `json:"sent"`
	Delivered int `json:"delivered"`
	Failed    int `json:"failed"`
}
