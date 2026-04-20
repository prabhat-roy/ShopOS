package handler

import (
	"strings"
	"time"
)

// RFCTime is a time.Time that marshals/unmarshals as RFC 3339.
type RFCTime struct {
	time.Time
}

// UnmarshalJSON parses a JSON string in RFC 3339 format.
func (t *RFCTime) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	if s == "null" || s == "" {
		return nil
	}
	parsed, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return err
	}
	t.Time = parsed
	return nil
}

// timePtr extracts the underlying *time.Time from an *RFCTime, returning nil when the input is nil.
func timePtr(r *RFCTime) *time.Time {
	if r == nil {
		return nil
	}
	t := r.Time.UTC()
	return &t
}
