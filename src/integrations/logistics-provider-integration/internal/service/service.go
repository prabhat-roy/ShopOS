package service

import (
	"fmt"
	"strings"

	"github.com/shopos/logistics-provider-integration/internal/adapter"
	"github.com/shopos/logistics-provider-integration/internal/domain"
	"github.com/shopos/logistics-provider-integration/internal/store"
)

// Servicer is the interface that defines the logistics business operations.
// It is exposed as an interface so handlers can be tested with a mock.
type Servicer interface {
	CreateShipment(req domain.ShipmentRequest) (domain.ShipmentResponse, error)
	GetShipment(trackingNumber string) (domain.ShipmentResponse, error)
	TrackShipment(trackingNumber string, provider domain.Provider) (domain.TrackingResponse, error)
	GetRates(fromPostal, toPostal string, weightKg float64) ([]domain.RateQuote, error)
	ListShipments(provider string, limit int) []domain.ShipmentResponse
}

// Service implements Servicer by coordinating the adapter and the in-memory store.
type Service struct {
	adapter *adapter.LogisticsAdapter
	store   *store.ShipmentStore
}

// New constructs a Service with the given adapter and store.
func New(a *adapter.LogisticsAdapter, s *store.ShipmentStore) *Service {
	return &Service{adapter: a, store: s}
}

// CreateShipment validates the request, delegates to the adapter, and persists the result.
func (svc *Service) CreateShipment(req domain.ShipmentRequest) (domain.ShipmentResponse, error) {
	if err := validateShipmentRequest(req); err != nil {
		return domain.ShipmentResponse{}, err
	}

	resp, err := svc.adapter.CreateShipment(req)
	if err != nil {
		return domain.ShipmentResponse{}, fmt.Errorf("adapter.CreateShipment: %w", err)
	}
	svc.store.SaveShipment(resp)
	return resp, nil
}

// GetShipment retrieves a previously created shipment by its tracking number.
func (svc *Service) GetShipment(trackingNumber string) (domain.ShipmentResponse, error) {
	if trackingNumber == "" {
		return domain.ShipmentResponse{}, fmt.Errorf("trackingNumber is required")
	}
	resp, ok := svc.store.GetShipment(trackingNumber)
	if !ok {
		return domain.ShipmentResponse{}, fmt.Errorf("shipment not found: %s", trackingNumber)
	}
	return resp, nil
}

// TrackShipment fetches live tracking data from the carrier adapter and caches it.
func (svc *Service) TrackShipment(trackingNumber string, provider domain.Provider) (domain.TrackingResponse, error) {
	if trackingNumber == "" {
		return domain.TrackingResponse{}, fmt.Errorf("trackingNumber is required")
	}
	if !isValidProvider(provider) {
		return domain.TrackingResponse{}, fmt.Errorf("unsupported provider: %s", provider)
	}

	resp, err := svc.adapter.GetTracking(trackingNumber, provider)
	if err != nil {
		return domain.TrackingResponse{}, fmt.Errorf("adapter.GetTracking: %w", err)
	}
	svc.store.SaveTracking(resp)
	return resp, nil
}

// GetRates returns rate quotes from all carriers for the given route and weight.
func (svc *Service) GetRates(fromPostal, toPostal string, weightKg float64) ([]domain.RateQuote, error) {
	if fromPostal == "" || toPostal == "" {
		return nil, fmt.Errorf("fromPostal and toPostal are required")
	}
	if weightKg <= 0 {
		return nil, fmt.Errorf("weightKg must be greater than 0")
	}
	return svc.adapter.GetRates(fromPostal, toPostal, weightKg)
}

// ListShipments returns stored shipments, optionally filtered by provider, up to limit.
func (svc *Service) ListShipments(provider string, limit int) []domain.ShipmentResponse {
	return svc.store.ListShipments(provider, limit)
}

// ---------------------------------------------------------------------------
// validation helpers
// ---------------------------------------------------------------------------

func validateShipmentRequest(req domain.ShipmentRequest) error {
	if !isValidProvider(req.Provider) {
		return fmt.Errorf("unsupported provider: %q", req.Provider)
	}
	if req.WeightKg <= 0 {
		return fmt.Errorf("weightKg must be greater than 0")
	}
	if req.ServiceType == "" {
		return fmt.Errorf("serviceType is required")
	}
	if err := validateAddress(req.FromAddress, "fromAddress"); err != nil {
		return err
	}
	if err := validateAddress(req.ToAddress, "toAddress"); err != nil {
		return err
	}
	return nil
}

func validateAddress(addr domain.Address, field string) error {
	if strings.TrimSpace(addr.Street1) == "" {
		return fmt.Errorf("%s.street1 is required", field)
	}
	if strings.TrimSpace(addr.City) == "" {
		return fmt.Errorf("%s.city is required", field)
	}
	if strings.TrimSpace(addr.PostalCode) == "" {
		return fmt.Errorf("%s.postalCode is required", field)
	}
	if strings.TrimSpace(addr.Country) == "" {
		return fmt.Errorf("%s.country is required", field)
	}
	return nil
}

func isValidProvider(p domain.Provider) bool {
	for _, v := range domain.AllProviders {
		if v == p {
			return true
		}
	}
	return false
}
