package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/shopos/device-fingerprint-service/internal/domain"
)

const (
	keyPrefixFP     = "fp:"
	keyPrefixFPID   = "fpid:"
	keyPrefixUserFP = "user_fps:"
)

// FingerprintStore defines the persistence contract used by the service layer.
type FingerprintStore interface {
	Save(ctx context.Context, fp *domain.Fingerprint) error
	Get(ctx context.Context, hash string) (*domain.Fingerprint, error)
	GetByID(ctx context.Context, id string) (*domain.Fingerprint, error)
	IncrSeen(ctx context.Context, hash string) error
	GetUserFingerprints(ctx context.Context, userID string) ([]*domain.Fingerprint, error)
	LinkToUser(ctx context.Context, userID, hash string) error
}

// redisFingerprintStore is the Redis-backed implementation of FingerprintStore.
type redisFingerprintStore struct {
	client *redis.Client
	ttl    time.Duration
}

// NewRedisFingerprintStore creates a new store backed by the provided Redis client.
// ttlDays is the number of days a fingerprint record is retained.
func NewRedisFingerprintStore(client *redis.Client, ttlDays int) FingerprintStore {
	return &redisFingerprintStore{
		client: client,
		ttl:    time.Duration(ttlDays) * 24 * time.Hour,
	}
}

// Save persists a Fingerprint record under fp:{hash} and creates a reverse index
// fpid:{id} → hash so the record is retrievable by UUID as well.
func (s *redisFingerprintStore) Save(ctx context.Context, fp *domain.Fingerprint) error {
	data, err := json.Marshal(fp)
	if err != nil {
		return fmt.Errorf("store.Save marshal: %w", err)
	}

	hashKey := keyPrefixFP + fp.Hash
	idKey := keyPrefixFPID + fp.ID

	pipe := s.client.Pipeline()
	pipe.Set(ctx, hashKey, data, s.ttl)
	pipe.Set(ctx, idKey, fp.Hash, s.ttl)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("store.Save pipeline exec: %w", err)
	}
	return nil
}

// Get retrieves a Fingerprint by its SHA-256 hash.
func (s *redisFingerprintStore) Get(ctx context.Context, hash string) (*domain.Fingerprint, error) {
	data, err := s.client.Get(ctx, keyPrefixFP+hash).Bytes()
	if err == redis.Nil {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("store.Get: %w", err)
	}

	var fp domain.Fingerprint
	if err = json.Unmarshal(data, &fp); err != nil {
		return nil, fmt.Errorf("store.Get unmarshal: %w", err)
	}
	return &fp, nil
}

// GetByID resolves a fingerprint UUID to its hash via the reverse index and then
// fetches the full record.
func (s *redisFingerprintStore) GetByID(ctx context.Context, id string) (*domain.Fingerprint, error) {
	hash, err := s.client.Get(ctx, keyPrefixFPID+id).Result()
	if err == redis.Nil {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("store.GetByID lookup: %w", err)
	}
	return s.Get(ctx, hash)
}

// IncrSeen atomically increments the seen_count on the stored JSON and resets
// the TTL.  It does a read-modify-write because Redis does not natively support
// a JSON field increment without the RedisJSON module.
func (s *redisFingerprintStore) IncrSeen(ctx context.Context, hash string) error {
	key := keyPrefixFP + hash

	// Optimistic loop: read → modify → write.
	for attempt := 0; attempt < 3; attempt++ {
		data, err := s.client.Get(ctx, key).Bytes()
		if err == redis.Nil {
			return domain.ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("store.IncrSeen get: %w", err)
		}

		var fp domain.Fingerprint
		if err = json.Unmarshal(data, &fp); err != nil {
			return fmt.Errorf("store.IncrSeen unmarshal: %w", err)
		}

		fp.SeenCount++
		fp.LastSeen = time.Now().UTC()

		updated, err := json.Marshal(fp)
		if err != nil {
			return fmt.Errorf("store.IncrSeen marshal: %w", err)
		}

		// Use GETSET to detect concurrent modification.
		err = s.client.Set(ctx, key, updated, s.ttl).Err()
		if err == nil {
			return nil
		}
	}
	return fmt.Errorf("store.IncrSeen: max retries exceeded")
}

// GetUserFingerprints returns all Fingerprint records associated with a user.
func (s *redisFingerprintStore) GetUserFingerprints(ctx context.Context, userID string) ([]*domain.Fingerprint, error) {
	members, err := s.client.SMembers(ctx, keyPrefixUserFP+userID).Result()
	if err != nil {
		return nil, fmt.Errorf("store.GetUserFingerprints smembers: %w", err)
	}

	fps := make([]*domain.Fingerprint, 0, len(members))
	for _, hash := range members {
		fp, err := s.Get(ctx, hash)
		if err == domain.ErrNotFound {
			// Record may have expired; skip silently.
			continue
		}
		if err != nil {
			return nil, err
		}
		fps = append(fps, fp)
	}
	return fps, nil
}

// LinkToUser adds the fingerprint hash to the user's set of known fingerprints.
func (s *redisFingerprintStore) LinkToUser(ctx context.Context, userID, hash string) error {
	err := s.client.SAdd(ctx, keyPrefixUserFP+userID, hash).Err()
	if err != nil {
		return fmt.Errorf("store.LinkToUser sadd: %w", err)
	}
	// Refresh the set TTL whenever a new fingerprint is linked.
	s.client.Expire(ctx, keyPrefixUserFP+userID, s.ttl)
	return nil
}
