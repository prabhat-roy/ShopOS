package store

import (
	"fmt"
	"strings"
	"sync"

	"github.com/shopos/logistics-provider-integration/internal/domain"
)

// ShipmentStore is a thread-safe in-memory store for shipments and tracking data.
type ShipmentStore struct {
	mu        sync.RWMutex
	shipments map[string]domain.ShipmentResponse  // keyed by trackingNumber
	tracking  map[string]domain.TrackingResponse   // keyed by trackingNumber
}

// New creates and returns an initialised ShipmentStore.
func New() *ShipmentStore {
	return &ShipmentStore{
		shipments: make(map[string]domain.ShipmentResponse),
		tracking:  make(map[string]domain.TrackingResponse),
	}
}

// SaveShipment persists a ShipmentResponse.
func (s *ShipmentStore) SaveShipment(resp domain.ShipmentResponse) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.shipments[resp.TrackingNumber] = resp
}

// GetShipment retrieves a ShipmentResponse by tracking number.
// Returns the response and true if found, or an empty struct and false otherwise.
func (s *ShipmentStore) GetShipment(trackingNumber string) (domain.ShipmentResponse, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.shipments[trackingNumber]
	return r, ok
}

// ListShipments returns shipments optionally filtered by provider.
// When provider is empty all shipments are returned, up to limit entries.
// A limit of 0 means "no limit".
func (s *ShipmentStore) ListShipments(provider string, limit int) []domain.ShipmentResponse {
	s.mu.RLock()
	defer s.mu.RUnlock()

	results := make([]domain.ShipmentResponse, 0, len(s.shipments))
	for _, sh := range s.shipments {
		if provider != "" && !strings.EqualFold(string(sh.Provider), provider) {
			continue
		}
		results = append(results, sh)
		if limit > 0 && len(results) >= limit {
			break
		}
	}
	return results
}

// SaveTracking persists a TrackingResponse.
func (s *ShipmentStore) SaveTracking(resp domain.TrackingResponse) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tracking[resp.TrackingNumber] = resp
}

// GetTracking retrieves a TrackingResponse by tracking number.
func (s *ShipmentStore) GetTracking(trackingNumber string) (domain.TrackingResponse, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.tracking[trackingNumber]
	return r, ok
}

// Stats returns a simple diagnostic map for the health endpoint.
func (s *ShipmentStore) Stats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return map[string]interface{}{
		"shipments": fmt.Sprintf("%d", len(s.shipments)),
		"tracking":  fmt.Sprintf("%d", len(s.tracking)),
	}
}
