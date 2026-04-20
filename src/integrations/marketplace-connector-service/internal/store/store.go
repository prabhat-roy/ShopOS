package store

import (
	"fmt"
	"sync"

	"github.com/shopos/marketplace-connector-service/internal/domain"
)

// SyncStore is a thread-safe in-memory store for SyncRecords.
type SyncStore struct {
	mu      sync.RWMutex
	records map[string]*domain.SyncRecord
}

// New returns an initialised SyncStore.
func New() *SyncStore {
	return &SyncStore{
		records: make(map[string]*domain.SyncRecord),
	}
}

// SaveSync persists or replaces a SyncRecord.
func (s *SyncStore) SaveSync(rec *domain.SyncRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.records[rec.ID] = rec
}

// GetSync retrieves a SyncRecord by ID.
func (s *SyncStore) GetSync(id string) (*domain.SyncRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.records[id]
	if !ok {
		return nil, fmt.Errorf("sync record %q not found", id)
	}
	return r, nil
}

// ListSyncs returns up to limit records, optionally filtered by marketplace.
// Passing an empty marketplace string returns records for all marketplaces.
func (s *SyncStore) ListSyncs(marketplace domain.Marketplace, limit int) []*domain.SyncRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]*domain.SyncRecord, 0, limit)
	for _, r := range s.records {
		if marketplace != "" && r.Marketplace != marketplace {
			continue
		}
		out = append(out, r)
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out
}

// GetStats returns a count of completed syncs grouped by marketplace.
func (s *SyncStore) GetStats() map[domain.Marketplace]int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := map[domain.Marketplace]int{
		domain.MarketplaceAmazon:  0,
		domain.MarketplaceEbay:    0,
		domain.MarketplaceEtsy:    0,
		domain.MarketplaceWalmart: 0,
	}
	for _, r := range s.records {
		if r.Status == domain.SyncStatusCompleted {
			stats[r.Marketplace]++
		}
	}
	return stats
}
