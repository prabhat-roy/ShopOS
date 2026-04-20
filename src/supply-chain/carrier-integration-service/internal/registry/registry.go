package registry

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/shopos/carrier-integration-service/internal/domain"
)

// carrierRate holds the per-carrier pricing configuration used to simulate quotes.
type carrierRate struct {
	baseRate       float64 // base cost in USD regardless of weight
	weightRate     float64 // additional cost per kg
	internationalM float64 // multiplier applied for international shipments
	daysMin        int     // minimum estimated transit days
	daysMax        int     // maximum estimated transit days
}

// serviceConfig maps a service name to its pricing and transit parameters.
var serviceConfigs = map[string]carrierRate{
	// FedEx
	"fedex_ground":    {baseRate: 8.50, weightRate: 0.75, internationalM: 2.5, daysMin: 3, daysMax: 7},
	"fedex_express":   {baseRate: 22.00, weightRate: 1.50, internationalM: 2.0, daysMin: 1, daysMax: 3},
	"fedex_overnight": {baseRate: 45.00, weightRate: 2.50, internationalM: 1.8, daysMin: 1, daysMax: 1},
	// UPS
	"ups_ground":   {baseRate: 9.00, weightRate: 0.70, internationalM: 2.5, daysMin: 3, daysMax: 7},
	"ups_2day":     {baseRate: 28.00, weightRate: 1.80, internationalM: 2.0, daysMin: 2, daysMax: 2},
	"ups_nextday":  {baseRate: 48.00, weightRate: 2.75, internationalM: 1.8, daysMin: 1, daysMax: 1},
	// DHL
	"dhl_express": {baseRate: 35.00, weightRate: 2.20, internationalM: 1.5, daysMin: 1, daysMax: 3},
	"dhl_economy": {baseRate: 18.00, weightRate: 1.10, internationalM: 1.6, daysMin: 5, daysMax: 10},
	// USPS
	"usps_priority": {baseRate: 7.90, weightRate: 0.60, internationalM: 2.2, daysMin: 1, daysMax: 3},
	"usps_ground":   {baseRate: 5.50, weightRate: 0.45, internationalM: 2.8, daysMin: 2, daysMax: 8},
}

// CarrierRegistry is an in-memory registry of carriers and their capabilities.
type CarrierRegistry struct {
	mu       sync.RWMutex
	carriers map[string]*domain.Carrier
	// shipments stores created shipments keyed by tracking number
	shipments map[string]*domain.ShipmentResponse
}

// New returns a fully initialised CarrierRegistry populated with the four
// supported carriers: FedEx, UPS, DHL, and USPS.
func New() *CarrierRegistry {
	r := &CarrierRegistry{
		carriers:  make(map[string]*domain.Carrier),
		shipments: make(map[string]*domain.ShipmentResponse),
	}

	r.carriers["fedex"] = &domain.Carrier{
		ID:                "fedex",
		Name:              "FedEx",
		Code:              "FEDEX",
		SupportedServices: []string{"Ground", "Express", "Overnight"},
		Active:            true,
	}
	r.carriers["ups"] = &domain.Carrier{
		ID:                "ups",
		Name:              "UPS",
		Code:              "UPS",
		SupportedServices: []string{"Ground", "2Day", "NextDay"},
		Active:            true,
	}
	r.carriers["dhl"] = &domain.Carrier{
		ID:                "dhl",
		Name:              "DHL",
		Code:              "DHL",
		SupportedServices: []string{"Express", "Economy"},
		Active:            true,
	}
	r.carriers["usps"] = &domain.Carrier{
		ID:                "usps",
		Name:              "USPS",
		Code:              "USPS",
		SupportedServices: []string{"Priority", "Ground"},
		Active:            true,
	}

	return r
}

// GetCarrier returns the carrier for the given id or ErrCarrierNotFound.
func (r *CarrierRegistry) GetCarrier(id string) (*domain.Carrier, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.carriers[strings.ToLower(id)]
	if !ok {
		return nil, domain.ErrCarrierNotFound
	}
	return c, nil
}

// ListCarriers returns all carriers (active and inactive).
func (r *CarrierRegistry) ListCarriers() []*domain.Carrier {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*domain.Carrier, 0, len(r.carriers))
	for _, c := range r.carriers {
		out = append(out, c)
	}
	return out
}

// ListActive returns only active carriers.
func (r *CarrierRegistry) ListActive() []*domain.Carrier {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*domain.Carrier, 0, len(r.carriers))
	for _, c := range r.carriers {
		if c.Active {
			out = append(out, c)
		}
	}
	return out
}

// GetRate returns a single rate quote for the specified carrier and service.
// Rates are simulated using fixed base rates plus a weight multiplier; an
// international multiplier is applied when countryFrom != countryTo.
func (r *CarrierRegistry) GetRate(carrierID, service string, req domain.RateQuoteRequest) (*domain.RateQuoteResponse, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	carrier, ok := r.carriers[strings.ToLower(carrierID)]
	if !ok {
		return nil, domain.ErrCarrierNotFound
	}

	serviceKey := buildServiceKey(carrierID, service)
	cfg, ok := serviceConfigs[serviceKey]
	if !ok {
		return nil, domain.ErrServiceNotFound
	}

	// Validate service is supported by the carrier
	if !serviceSupported(carrier, service) {
		return nil, domain.ErrServiceNotFound
	}

	price := cfg.baseRate + cfg.weightRate*math.Max(req.WeightKg, 0.1)
	if !strings.EqualFold(req.CountryFrom, req.CountryTo) {
		price *= cfg.internationalM
	}
	price = math.Round(price*100) / 100

	days := cfg.daysMin
	if req.CountryFrom != req.CountryTo {
		days = cfg.daysMax
	}

	return &domain.RateQuoteResponse{
		Carrier:       carrier.Name,
		Service:       service,
		EstimatedDays: days,
		Price:         price,
		Currency:      "USD",
	}, nil
}

// GetAllRates returns rate quotes from every active carrier for every service,
// enabling the caller to perform "rate shopping".
func (r *CarrierRegistry) GetAllRates(req domain.RateQuoteRequest) ([]domain.RateQuoteResponse, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var quotes []domain.RateQuoteResponse
	for id, carrier := range r.carriers {
		if !carrier.Active {
			continue
		}
		for _, svc := range carrier.SupportedServices {
			key := buildServiceKey(id, svc)
			cfg, ok := serviceConfigs[key]
			if !ok {
				continue
			}
			price := cfg.baseRate + cfg.weightRate*math.Max(req.WeightKg, 0.1)
			if !strings.EqualFold(req.CountryFrom, req.CountryTo) {
				price *= cfg.internationalM
			}
			price = math.Round(price*100) / 100

			days := cfg.daysMin
			if req.CountryFrom != req.CountryTo {
				days = cfg.daysMax
			}

			quotes = append(quotes, domain.RateQuoteResponse{
				Carrier:       carrier.Name,
				Service:       svc,
				EstimatedDays: days,
				Price:         price,
				Currency:      "USD",
			})
		}
	}
	return quotes, nil
}

// CreateShipment creates a new shipment record in the in-memory store and
// returns the shipment details including a synthetically generated tracking number.
func (r *CarrierRegistry) CreateShipment(req domain.ShipmentRequest) (*domain.ShipmentResponse, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	carrier, ok := r.carriers[strings.ToLower(req.CarrierID)]
	if !ok {
		return nil, domain.ErrCarrierNotFound
	}
	if !serviceSupported(carrier, req.Service) {
		return nil, domain.ErrServiceNotFound
	}

	serviceKey := buildServiceKey(req.CarrierID, req.Service)
	cfg, ok := serviceConfigs[serviceKey]
	if !ok {
		return nil, domain.ErrServiceNotFound
	}

	trackingNumber := generateTrackingNumber(carrier.Code)
	cost := math.Round((cfg.baseRate+cfg.weightRate*math.Max(req.WeightKg, 0.1))*100) / 100
	labelURL := fmt.Sprintf("https://labels.shopos.internal/%s/%s.pdf", strings.ToLower(carrier.Code), trackingNumber)

	resp := &domain.ShipmentResponse{
		TrackingNumber: trackingNumber,
		LabelURL:       labelURL,
		Carrier:        carrier.Name,
		Service:        req.Service,
		Cost:           cost,
		Currency:       "USD",
	}

	r.shipments[trackingNumber] = resp
	return resp, nil
}

// GetTracking returns a synthetic 3-event tracking history for a previously
// created shipment.  Returns ErrTrackingNotFound if the tracking number is
// unknown.
func (r *CarrierRegistry) GetTracking(trackingNumber string) (*domain.TrackResponse, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	shipment, ok := r.shipments[trackingNumber]
	if !ok {
		return nil, domain.ErrTrackingNotFound
	}

	now := time.Now().UTC()
	events := []domain.TrackEvent{
		{
			Timestamp:   now.Add(-48 * time.Hour).Format(time.RFC3339),
			Location:    "Origin Facility, Sorting Hub",
			Description: "Shipment picked up and scanned at origin facility",
			Status:      "PICKED_UP",
		},
		{
			Timestamp:   now.Add(-24 * time.Hour).Format(time.RFC3339),
			Location:    "Regional Distribution Center",
			Description: "Package arrived at regional distribution center",
			Status:      "IN_TRANSIT",
		},
		{
			Timestamp:   now.Format(time.RFC3339),
			Location:    "Local Delivery Hub",
			Description: "Package out for delivery",
			Status:      "OUT_FOR_DELIVERY",
		},
	}

	return &domain.TrackResponse{
		TrackingNumber: trackingNumber,
		Carrier:        shipment.Carrier,
		Status:         "OUT_FOR_DELIVERY",
		Events:         events,
	}, nil
}

// ─── helpers ─────────────────────────────────────────────────────────────────

// buildServiceKey constructs the map key used in serviceConfigs.
// e.g. ("fedex", "Ground") → "fedex_ground"
func buildServiceKey(carrierID, service string) string {
	return strings.ToLower(carrierID) + "_" + strings.ToLower(service)
}

// serviceSupported checks whether the carrier supports the given service name.
func serviceSupported(carrier *domain.Carrier, service string) bool {
	for _, s := range carrier.SupportedServices {
		if strings.EqualFold(s, service) {
			return true
		}
	}
	return false
}

// generateTrackingNumber creates a tracking number in the format CARRIER-XXXXXXXX.
func generateTrackingNumber(carrierCode string) string {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		// fallback: use timestamp-based suffix
		return fmt.Sprintf("%s-%d", carrierCode, time.Now().UnixNano())
	}
	return fmt.Sprintf("%s-%s", carrierCode, strings.ToUpper(hex.EncodeToString(b)))
}
