package domain

import "errors"

// Location holds all geolocation data resolved for an IP address.
type Location struct {
	IP          string  `json:"ip"`
	CountryCode string  `json:"country_code"`
	CountryName string  `json:"country_name"`
	Region      string  `json:"region"`
	City        string  `json:"city"`
	PostalCode  string  `json:"postal_code"`
	Timezone    string  `json:"timezone"`
	ISP         string  `json:"isp"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
}

// Sentinel errors returned by the lookup layer.
var (
	ErrInvalidIP = errors.New("invalid IP address")
	ErrNotFound  = errors.New("location not found")
)
