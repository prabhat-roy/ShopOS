package adapter

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopos/logistics-provider-integration/internal/domain"
)

// LogisticsAdapter simulates interactions with external carrier APIs.
// In production each method would call the real carrier REST/SOAP endpoint.
type LogisticsAdapter struct{}

// New returns a ready-to-use LogisticsAdapter.
func New() *LogisticsAdapter {
	return &LogisticsAdapter{}
}

// CreateShipment generates a provider-specific tracking number, a label URL,
// and estimates the shipping cost based on weight and service type.
func (a *LogisticsAdapter) CreateShipment(req domain.ShipmentRequest) (domain.ShipmentResponse, error) {
	tracking := generateTrackingNumber(req.Provider)
	cost := estimateCost(req.Provider, req.WeightKg, req.ServiceType)
	transit := transitDays(req.Provider, req.ServiceType)
	eta := time.Now().Add(time.Duration(transit) * 24 * time.Hour).Format("2006-01-02")
	labelID := strings.ReplaceAll(uuid.New().String(), "-", "")

	resp := domain.ShipmentResponse{
		TrackingNumber:    tracking,
		Provider:          req.Provider,
		ServiceType:       req.ServiceType,
		EstimatedDelivery: eta,
		LabelURL: fmt.Sprintf(
			"https://labels.shopos.internal/%s/%s/%s.pdf",
			strings.ToLower(string(req.Provider)), tracking, labelID,
		),
		Cost:      math.Round(cost*100) / 100,
		Currency:  "USD",
		CreatedAt: time.Now().UTC(),
	}
	return resp, nil
}

// GetTracking returns a realistic synthetic tracking history for the given number.
func (a *LogisticsAdapter) GetTracking(trackingNumber string, provider domain.Provider) (domain.TrackingResponse, error) {
	now := time.Now().UTC()
	events := buildTrackingEvents(provider, trackingNumber, now)
	lastEvent := events[len(events)-1]

	status := lastEvent.Status
	resp := domain.TrackingResponse{
		TrackingNumber:    trackingNumber,
		Provider:          provider,
		Status:            status,
		Events:            events,
		EstimatedDelivery: now.Add(48 * time.Hour).Format("2006-01-02"),
	}
	if status == "DELIVERED" {
		delivered := lastEvent.Timestamp.Format(time.RFC3339)
		resp.ActualDelivery = &delivered
		resp.EstimatedDelivery = lastEvent.Timestamp.Format("2006-01-02")
	}
	return resp, nil
}

// GetRates returns a rate quote from every supported carrier for the given route and weight.
func (a *LogisticsAdapter) GetRates(fromPostal, toPostal string, weightKg float64) ([]domain.RateQuote, error) {
	quotes := make([]domain.RateQuote, 0, 10)
	for _, p := range domain.AllProviders {
		for _, svc := range servicesForProvider(p) {
			base := estimateCost(p, weightKg, svc)
			quotes = append(quotes, domain.RateQuote{
				Provider:    p,
				ServiceType: svc,
				TransitDays: transitDays(p, svc),
				Price:       math.Round(base*100) / 100,
				Currency:    "USD",
			})
		}
	}
	return quotes, nil
}

// FormatAddress returns a provider-specific address map as expected by each carrier's API.
func (a *LogisticsAdapter) FormatAddress(provider domain.Provider, addr domain.Address) map[string]interface{} {
	switch provider {
	case domain.ProviderFedEx:
		return map[string]interface{}{
			"streetLines":         []string{addr.Street1, addr.Street2},
			"city":                addr.City,
			"stateOrProvinceCode": addr.State,
			"postalCode":          addr.PostalCode,
			"countryCode":         addr.Country,
		}
	case domain.ProviderUPS:
		return map[string]interface{}{
			"AddressLine":       addr.Street1,
			"City":              addr.City,
			"StateProvinceCode": addr.State,
			"PostalCode":        addr.PostalCode,
			"CountryCode":       addr.Country,
		}
	case domain.ProviderDHL:
		return map[string]interface{}{
			"addressLine1": addr.Street1,
			"addressLine2": addr.Street2,
			"cityName":     addr.City,
			"countyName":   addr.State,
			"postalCode":   addr.PostalCode,
			"countryCode":  addr.Country,
		}
	case domain.ProviderUSPS:
		return map[string]interface{}{
			"Address1": addr.Street2,
			"Address2": addr.Street1,
			"City":     addr.City,
			"State":    addr.State,
			"Zip5":     addr.PostalCode,
		}
	case domain.ProviderShipBob:
		return map[string]interface{}{
			"address1":   addr.Street1,
			"address2":   addr.Street2,
			"city":       addr.City,
			"state":      addr.State,
			"zip_code":   addr.PostalCode,
			"country":    addr.Country,
			"name":       addr.Name,
		}
	default:
		return map[string]interface{}{
			"street":  addr.Street1,
			"city":    addr.City,
			"state":   addr.State,
			"zip":     addr.PostalCode,
			"country": addr.Country,
		}
	}
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// generateTrackingNumber creates a provider-realistic tracking number.
func generateTrackingNumber(provider domain.Provider) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	switch provider {
	case domain.ProviderUPS:
		// UPS: 1Z + 6 alphanum + 2 digits + 8 digits
		return fmt.Sprintf("1Z%06X%02d%08d", r.Intn(0xFFFFFF), r.Intn(100), r.Intn(100000000))
	case domain.ProviderUSPS:
		// USPS: 9400 + 14 digits (22 total)
		return fmt.Sprintf("9400%018d", r.Int63n(1_000_000_000_000_000_000))
	case domain.ProviderFedEx:
		// FedEx: 7489 + 16 digits
		return fmt.Sprintf("7489%016d", r.Int63n(10_000_000_000_000_000))
	case domain.ProviderDHL:
		// DHL: JD0 + 15 digits
		return fmt.Sprintf("JD0%015d", r.Int63n(1_000_000_000_000_000))
	case domain.ProviderShipBob:
		// ShipBob: SB- + 8 hex chars
		return fmt.Sprintf("SB-%08X", r.Intn(0xFFFFFFFF))
	default:
		return uuid.New().String()
	}
}

// estimateCost approximates carrier cost based on weight and service tier.
func estimateCost(provider domain.Provider, weightKg float64, serviceType string) float64 {
	base := map[domain.Provider]float64{
		domain.ProviderFedEx:   8.50,
		domain.ProviderUPS:     7.95,
		domain.ProviderDHL:     9.20,
		domain.ProviderUSPS:    5.75,
		domain.ProviderShipBob: 6.40,
	}[provider]

	multiplier := 1.0
	svc := strings.ToUpper(serviceType)
	switch {
	case strings.Contains(svc, "OVERNIGHT") || strings.Contains(svc, "EXPRESS") || strings.Contains(svc, "PRIORITY"):
		multiplier = 3.5
	case strings.Contains(svc, "2DAY") || strings.Contains(svc, "TWO_DAY") || strings.Contains(svc, "2_DAY"):
		multiplier = 2.2
	case strings.Contains(svc, "3DAY") || strings.Contains(svc, "THREE_DAY"):
		multiplier = 1.6
	default:
		multiplier = 1.0
	}

	weightFee := weightKg * 1.25
	return (base + weightFee) * multiplier
}

// transitDays returns the expected number of business days for a service tier.
func transitDays(provider domain.Provider, serviceType string) int {
	svc := strings.ToUpper(serviceType)
	switch {
	case strings.Contains(svc, "OVERNIGHT") || strings.Contains(svc, "EXPRESS"):
		return 1
	case strings.Contains(svc, "2DAY") || strings.Contains(svc, "TWO_DAY") || strings.Contains(svc, "2_DAY"):
		return 2
	case strings.Contains(svc, "3DAY") || strings.Contains(svc, "THREE_DAY"):
		return 3
	case provider == domain.ProviderUSPS:
		return 5
	case provider == domain.ProviderShipBob:
		return 4
	default:
		return 5
	}
}

// servicesForProvider returns the main service types offered by a carrier.
func servicesForProvider(provider domain.Provider) []string {
	switch provider {
	case domain.ProviderFedEx:
		return []string{"FEDEX_OVERNIGHT", "FEDEX_2DAY", "FEDEX_3DAY", "FEDEX_GROUND"}
	case domain.ProviderUPS:
		return []string{"UPS_NEXT_DAY_AIR", "UPS_2ND_DAY_AIR", "UPS_THREE_DAY", "UPS_GROUND"}
	case domain.ProviderDHL:
		return []string{"DHL_EXPRESS_OVERNIGHT", "DHL_EXPRESS_2DAY", "DHL_ECONOMY"}
	case domain.ProviderUSPS:
		return []string{"PRIORITY_MAIL_EXPRESS", "PRIORITY_MAIL", "FIRST_CLASS", "MEDIA_MAIL"}
	case domain.ProviderShipBob:
		return []string{"SHIPBOB_2_DAY", "SHIPBOB_STANDARD"}
	default:
		return []string{"STANDARD"}
	}
}

// buildTrackingEvents creates a realistic 3-4 event tracking history.
func buildTrackingEvents(provider domain.Provider, trackingNumber string, now time.Time) []domain.TrackingEvent {
	hub := carrierHub(provider)
	events := []domain.TrackingEvent{
		{
			Timestamp:   now.Add(-72 * time.Hour),
			Location:    "Origin Facility, CA US",
			Status:      "PICKED_UP",
			Description: "Shipment picked up by carrier",
		},
		{
			Timestamp:   now.Add(-48 * time.Hour),
			Location:    hub,
			Status:      "IN_TRANSIT",
			Description: "Arrived at sort facility",
		},
		{
			Timestamp:   now.Add(-24 * time.Hour),
			Location:    "Local Delivery Facility, NY US",
			Status:      "OUT_FOR_DELIVERY",
			Description: "Out for delivery",
		},
	}

	// Roughly half the time show the package as delivered.
	if (now.UnixNano() % 2) == 0 {
		delivered := now.Add(-2 * time.Hour)
		events = append(events, domain.TrackingEvent{
			Timestamp:   delivered,
			Location:    "Destination",
			Status:      "DELIVERED",
			Description: "Delivered – left at front door",
		})
	}
	return events
}

// carrierHub returns a plausible hub city for transit event simulation.
func carrierHub(provider domain.Provider) string {
	switch provider {
	case domain.ProviderFedEx:
		return "FedEx Hub, Memphis TN US"
	case domain.ProviderUPS:
		return "UPS Worldport, Louisville KY US"
	case domain.ProviderDHL:
		return "DHL Hub, Cincinnati OH US"
	case domain.ProviderUSPS:
		return "USPS NDC, Kansas City MO US"
	case domain.ProviderShipBob:
		return "ShipBob Fulfillment Center, Chicago IL US"
	default:
		return "Carrier Hub, US"
	}
}
