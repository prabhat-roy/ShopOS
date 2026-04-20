// Package store provides an in-memory store for PushRecord objects.
package store

import (
	"sync"

	"github.com/shopos/push-notification-service/internal/domain"
)

// Store is a bounded, thread-safe in-memory collection of PushRecords.
// When the store reaches maxSize it evicts the oldest entry before inserting a
// new one, preserving insertion order via a slice-based key list.
type Store struct {
	mu      sync.RWMutex
	records map[string]domain.PushRecord
	order   []string // insertion-order tracking for eviction
	maxSize int

	// aggregate counters
	sent      int
	delivered int
	failed    int
}

// New creates a Store with the given maximum capacity.
func New(maxSize int) *Store {
	return &Store{
		records: make(map[string]domain.PushRecord, maxSize),
		order:   make([]string, 0, maxSize),
		maxSize: maxSize,
	}
}

// Save persists a PushRecord, evicting the oldest entry when at capacity.
func (s *Store) Save(r domain.PushRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Evict oldest until we have room
	for len(s.records) >= s.maxSize {
		oldest := s.order[0]
		s.order = s.order[1:]
		delete(s.records, oldest)
	}

	s.records[r.MessageID] = r
	s.order = append(s.order, r.MessageID)

	s.sent++
	if r.Status == "delivered" {
		s.delivered++
	} else {
		s.failed++
	}
}

// Get returns the PushRecord for messageID and a boolean indicating whether it
// was found.
func (s *Store) Get(messageID string) (domain.PushRecord, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.records[messageID]
	return r, ok
}

// List returns up to limit records in reverse-insertion order (newest first).
func (s *Store) List(limit int) []domain.PushRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()

	n := len(s.order)
	if limit > n {
		limit = n
	}

	out := make([]domain.PushRecord, 0, limit)
	for i := n - 1; i >= 0 && len(out) < limit; i-- {
		if r, ok := s.records[s.order[i]]; ok {
			out = append(out, r)
		}
	}
	return out
}

// Stats returns a snapshot of aggregate delivery counters.
func (s *Store) Stats() domain.PushStats {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return domain.PushStats{
		Sent:      s.sent,
		Delivered: s.delivered,
		Failed:    s.failed,
	}
}
