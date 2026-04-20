package domain

import (
	"errors"
	"time"
)

// DeviceAttributes holds the raw signals collected from a client device.
// All fields must be provided by the caller; optional fields are omitted when
// computing the canonical hash so they do not break existing fingerprints.
type DeviceAttributes struct {
	UserAgent  string `json:"user_agent"`
	AcceptLang string `json:"accept_language"`
	Timezone   string `json:"timezone"`
	ScreenRes  string `json:"screen_resolution"`
	Platform   string `json:"platform"`
	IPAddress  string `json:"ip_address"`
	// Optional — collected only when available in the browser environment.
	Plugins string `json:"plugins,omitempty"`
	Canvas  string `json:"canvas,omitempty"`
}

// Fingerprint is the persisted record for a unique device signature.
type Fingerprint struct {
	ID         string           `json:"id"`          // UUID v4
	Hash       string           `json:"hash"`        // SHA-256 of normalised attributes
	UserID     string           `json:"user_id,omitempty"`
	Attributes DeviceAttributes `json:"attributes"`
	TrustScore int              `json:"trust_score"` // 0-100
	SeenCount  int              `json:"seen_count"`
	FirstSeen  time.Time        `json:"first_seen"`
	LastSeen   time.Time        `json:"last_seen"`
}

// FingerprintRequest is the inbound payload for the identify endpoint.
type FingerprintRequest struct {
	UserID     string           `json:"user_id"`
	Attributes DeviceAttributes `json:"attributes"`
}

// FingerprintResponse is the outbound payload returned after identification.
type FingerprintResponse struct {
	FingerprintID string `json:"fingerprint_id"`
	Hash          string `json:"hash"`
	IsKnown       bool   `json:"is_known"`    // true if this device was seen before for this user
	TrustScore    int    `json:"trust_score"` // 0-100
}

// ErrNotFound is returned when a fingerprint cannot be located in the store.
var ErrNotFound = errors.New("fingerprint not found")
