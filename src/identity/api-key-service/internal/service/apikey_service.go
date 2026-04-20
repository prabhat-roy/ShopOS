package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopos/api-key-service/internal/domain"
)

// Storer defines the persistence interface consumed by Service.
// The store package satisfies this interface; tests can inject a mock.
type Storer interface {
	Create(ctx context.Context, key *domain.APIKey) (*domain.APIKey, error)
	GetByID(ctx context.Context, id string) (*domain.APIKey, error)
	GetByHash(ctx context.Context, hash string) (*domain.APIKey, error)
	List(ctx context.Context, ownerID string) ([]*domain.APIKey, error)
	Deactivate(ctx context.Context, id string) error
	TouchLastUsed(ctx context.Context, id string) error
	Delete(ctx context.Context, id string) error
}

// Service contains all business logic for API key management.
type Service struct {
	store Storer
}

// New wires a Service with the given Storer implementation.
func New(store Storer) *Service {
	return &Service{store: store}
}

// Create generates a new raw API key, stores only its hash and display prefix,
// and returns the stored APIKey together with the raw key (shown exactly once).
//
// The raw key format is:  sk_live_<uuid-no-dashes>
func (s *Service) Create(ctx context.Context, req *domain.CreateKeyRequest) (*domain.APIKey, string, error) {
	if req.OwnerID == "" {
		return nil, "", fmt.Errorf("owner_id is required")
	}
	if req.OwnerType == "" {
		req.OwnerType = "user"
	}
	if req.Name == "" {
		return nil, "", fmt.Errorf("name is required")
	}

	rawKey := "sk_live_" + uuid.New().String()
	// Strip dashes from the UUID portion for a cleaner key string.
	rawKey = "sk_live_" + stripDashes(uuid.New().String())

	prefix := rawKey[:8] // "sk_live_"
	hash := hashKey(rawKey)

	keyRecord := &domain.APIKey{
		ID:        uuid.New().String(),
		OwnerID:   req.OwnerID,
		OwnerType: req.OwnerType,
		Name:      req.Name,
		KeyPrefix: prefix,
		KeyHash:   hash,
		Scopes:    req.Scopes,
		Active:    true,
		ExpiresAt: req.ExpiresAt,
	}
	if keyRecord.Scopes == nil {
		keyRecord.Scopes = []string{}
	}

	stored, err := s.store.Create(ctx, keyRecord)
	if err != nil {
		return nil, "", fmt.Errorf("storing api key: %w", err)
	}
	return stored, rawKey, nil
}

// Validate hashes the supplied raw key, looks it up, verifies it is active and
// not expired, records the last-used timestamp, and returns a ValidateResponse.
func (s *Service) Validate(ctx context.Context, rawKey string) (*domain.ValidateResponse, error) {
	if rawKey == "" {
		return &domain.ValidateResponse{Valid: false, Reason: "key is empty"}, nil
	}

	hash := hashKey(rawKey)
	key, err := s.store.GetByHash(ctx, hash)
	if err == domain.ErrNotFound {
		return &domain.ValidateResponse{Valid: false, Reason: "key not found"}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("looking up key by hash: %w", err)
	}

	if !key.Active {
		return &domain.ValidateResponse{
			Valid:  false,
			KeyID:  key.ID,
			Reason: "key is inactive",
		}, nil
	}

	if key.ExpiresAt != nil && key.ExpiresAt.Before(time.Now()) {
		return &domain.ValidateResponse{
			Valid:  false,
			KeyID:  key.ID,
			Reason: "key has expired",
		}, nil
	}

	// Best-effort last-used update; do not fail validation on error.
	_ = s.store.TouchLastUsed(ctx, key.ID)

	return &domain.ValidateResponse{
		Valid:   true,
		KeyID:   key.ID,
		OwnerID: key.OwnerID,
		Scopes:  key.Scopes,
	}, nil
}

// List returns all API keys belonging to a given owner (raw key not included).
func (s *Service) List(ctx context.Context, ownerID string) ([]*domain.APIKey, error) {
	if ownerID == "" {
		return nil, fmt.Errorf("owner_id is required")
	}
	return s.store.List(ctx, ownerID)
}

// GetByID fetches a single key by ID (raw key not included).
func (s *Service) GetByID(ctx context.Context, id string) (*domain.APIKey, error) {
	return s.store.GetByID(ctx, id)
}

// Deactivate marks a key as inactive without deleting it.
func (s *Service) Deactivate(ctx context.Context, id string) error {
	return s.store.Deactivate(ctx, id)
}

// Delete permanently removes a key record.
func (s *Service) Delete(ctx context.Context, id string) error {
	return s.store.Delete(ctx, id)
}

// ---- helpers ----------------------------------------------------------------

func hashKey(rawKey string) string {
	sum := sha256.Sum256([]byte(rawKey))
	return hex.EncodeToString(sum[:])
}

func stripDashes(s string) string {
	out := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		if s[i] != '-' {
			out = append(out, s[i])
		}
	}
	return string(out)
}
