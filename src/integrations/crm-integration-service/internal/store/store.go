package store

import (
	"fmt"
	"sync"

	"github.com/shopos/crm-integration-service/internal/domain"
)

// SyncStore is a thread-safe in-memory store for CRM sync state.
type SyncStore struct {
	mu       sync.RWMutex
	results  map[string]*domain.SyncResult
	contacts map[string]*domain.CrmContact // key: "<crmSystem>:<crmId>"
}

// New returns an initialised SyncStore.
func New() *SyncStore {
	return &SyncStore{
		results:  make(map[string]*domain.SyncResult),
		contacts: make(map[string]*domain.CrmContact),
	}
}

// SaveResult persists or replaces a SyncResult.
func (s *SyncStore) SaveResult(r *domain.SyncResult) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.results[r.SyncID] = r
}

// GetResult retrieves a SyncResult by ID.
func (s *SyncStore) GetResult(id string) (*domain.SyncResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.results[id]
	if !ok {
		return nil, fmt.Errorf("sync result %q not found", id)
	}
	return r, nil
}

// ListResults returns up to limit results, optionally filtered by CRM system.
func (s *SyncStore) ListResults(crmSystem domain.CrmSystem, limit int) []*domain.SyncResult {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]*domain.SyncResult, 0, limit)
	for _, r := range s.results {
		if crmSystem != "" && r.CrmSystem != crmSystem {
			continue
		}
		out = append(out, r)
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out
}

// contactKey builds the composite lookup key.
func contactKey(crmSystem domain.CrmSystem, crmID string) string {
	return fmt.Sprintf("%s:%s", crmSystem, crmID)
}

// SaveContact upserts a CrmContact.
func (s *SyncStore) SaveContact(c *domain.CrmContact) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.contacts[contactKey(c.CrmSystem, c.CrmID)] = c
}

// GetContactByCrmId retrieves a contact by CRM system and CRM-side ID.
func (s *SyncStore) GetContactByCrmId(crmSystem domain.CrmSystem, crmID string) (*domain.CrmContact, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.contacts[contactKey(crmSystem, crmID)]
	if !ok {
		return nil, fmt.Errorf("contact %q not found in %s", crmID, crmSystem)
	}
	return c, nil
}

// ListContacts returns all contacts for a given CRM system.
func (s *SyncStore) ListContacts(crmSystem domain.CrmSystem) []*domain.CrmContact {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := []*domain.CrmContact{}
	for _, c := range s.contacts {
		if crmSystem == "" || c.CrmSystem == crmSystem {
			out = append(out, c)
		}
	}
	return out
}
