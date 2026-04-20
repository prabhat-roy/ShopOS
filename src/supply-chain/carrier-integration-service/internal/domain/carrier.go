package domain

import "errors"

// Carrier represents a shipping carrier and its capabilities.
type Carrier struct {
	ID                string   `json:"id"`
	Name              string   `json:"name"`
	Code              string   `json:"code"`
	SupportedServices []string `json:"supportedServices"`
	Active            bool     `json:"active"`
}

// Address holds a shipping address.
type Address struct {
	Name       string `json:"name"`
	Line1      string `json:"line1"`
	Line2      string `json:"line2,omitempty"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postalCode"`
	Country    string `json:"country"`
	Phone      string `json:"phone,omitempty"`
}

// RateQuoteRequest contains the details needed to retrieve shipping rate quotes.
type RateQuoteRequest struct {
	FromPostal  string  `json:"fromPostal"`
	ToPostal    string  `json:"toPostal"`
	CountryFrom string  `json:"countryFrom"`
	CountryTo   string  `json:"countryTo"`
	WeightKg    float64 `json:"weightKg"`
	LengthCm    float64 `json:"lengthCm"`
	WidthCm     float64 `json:"widthCm"`
	HeightCm    float64 `json:"heightCm"`
}

// RateQuoteResponse contains a single carrier/service rate quote.
type RateQuoteResponse struct {
	Carrier       string  `json:"carrier"`
	Service       string  `json:"service"`
	EstimatedDays int     `json:"estimatedDays"`
	Price         float64 `json:"price"`
	Currency      string  `json:"currency"`
}

// ShipmentRequest contains the details needed to create a shipment.
type ShipmentRequest struct {
	CarrierID   string  `json:"carrierId"`
	Service     string  `json:"service"`
	FromAddress Address `json:"fromAddress"`
	ToAddress   Address `json:"toAddress"`
	WeightKg    float64 `json:"weightKg"`
	Reference   string  `json:"reference,omitempty"`
}

// ShipmentResponse contains the result of a created shipment.
type ShipmentResponse struct {
	TrackingNumber string  `json:"trackingNumber"`
	LabelURL       string  `json:"labelUrl"`
	Carrier        string  `json:"carrier"`
	Service        string  `json:"service"`
	Cost           float64 `json:"cost"`
	Currency       string  `json:"currency"`
}

// TrackEvent represents a single tracking event in a shipment's history.
type TrackEvent struct {
	Timestamp   string `json:"timestamp"`
	Location    string `json:"location"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

// TrackResponse contains the tracking history for a shipment.
type TrackResponse struct {
	TrackingNumber string       `json:"trackingNumber"`
	Carrier        string       `json:"carrier"`
	Status         string       `json:"status"`
	Events         []TrackEvent `json:"events"`
}

// Sentinel errors for the carrier domain.
var (
	ErrCarrierNotFound  = errors.New("carrier not found")
	ErrServiceNotFound  = errors.New("service not supported by this carrier")
	ErrTrackingNotFound = errors.New("tracking number not found")
	ErrInvalidRequest   = errors.New("invalid request")
)
