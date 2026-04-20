package domain

import "time"

// Provider represents a supported logistics/shipping carrier.
type Provider string

const (
	ProviderFedEx   Provider = "FEDEX"
	ProviderUPS     Provider = "UPS"
	ProviderDHL     Provider = "DHL"
	ProviderUSPS    Provider = "USPS"
	ProviderShipBob Provider = "SHIPBOB"
)

// AllProviders lists every supported carrier.
var AllProviders = []Provider{
	ProviderFedEx,
	ProviderUPS,
	ProviderDHL,
	ProviderUSPS,
	ProviderShipBob,
}

// Address represents a physical mailing address.
type Address struct {
	Name       string `json:"name"`
	Street1    string `json:"street1"`
	Street2    string `json:"street2,omitempty"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postalCode"`
	Country    string `json:"country"`
}

// Dimensions holds the physical dimensions of a package.
type Dimensions struct {
	LengthCm float64 `json:"lengthCm"`
	WidthCm  float64 `json:"widthCm"`
	HeightCm float64 `json:"heightCm"`
}

// ShipmentRequest is the unified input for creating a shipment with any carrier.
type ShipmentRequest struct {
	Provider    Provider   `json:"provider"`
	FromAddress Address    `json:"fromAddress"`
	ToAddress   Address    `json:"toAddress"`
	WeightKg    float64    `json:"weightKg"`
	Dimensions  Dimensions `json:"dimensions"`
	ServiceType string     `json:"serviceType"`
	Reference   string     `json:"reference"`
}

// ShipmentResponse is the unified response after a shipment is created.
type ShipmentResponse struct {
	TrackingNumber    string    `json:"trackingNumber"`
	Provider          Provider  `json:"provider"`
	ServiceType       string    `json:"serviceType"`
	EstimatedDelivery string    `json:"estimatedDelivery"`
	LabelURL          string    `json:"labelUrl"`
	Cost              float64   `json:"cost"`
	Currency          string    `json:"currency"`
	CreatedAt         time.Time `json:"createdAt"`
}

// TrackingEvent is a single scan/update on a shipment's journey.
type TrackingEvent struct {
	Timestamp   time.Time `json:"timestamp"`
	Location    string    `json:"location"`
	Status      string    `json:"status"`
	Description string    `json:"description"`
}

// TrackingResponse contains the current state and full history of a shipment.
type TrackingResponse struct {
	TrackingNumber    string          `json:"trackingNumber"`
	Provider          Provider        `json:"provider"`
	Status            string          `json:"status"`
	Events            []TrackingEvent `json:"events"`
	EstimatedDelivery string          `json:"estimatedDelivery"`
	ActualDelivery    *string         `json:"actualDelivery,omitempty"`
}

// RateQuote is a carrier-specific price/transit estimate for a given route and weight.
type RateQuote struct {
	Provider    Provider `json:"provider"`
	ServiceType string   `json:"serviceType"`
	TransitDays int      `json:"transitDays"`
	Price       float64  `json:"price"`
	Currency    string   `json:"currency"`
}
