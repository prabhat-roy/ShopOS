package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/shopos/recently-viewed-service/internal/domain"
)

const keyPrefix = "recently_viewed:"

// Storer defines the persistence contract for recently-viewed operations.
type Storer interface {
	RecordView(ctx context.Context, customerID string, item domain.ViewedItem, maxItems int, ttl time.Duration) error
	GetRecent(ctx context.Context, customerID string, limit int) ([]domain.ViewedItem, error)
	ClearHistory(ctx context.Context, customerID string) error
	GetCount(ctx context.Context, customerID string) (int, error)
}

// RedisStore implements Storer using Redis sorted sets.
// Key pattern: "recently_viewed:{customerId}"
// Score = Unix timestamp (higher = more recent); members = JSON-encoded ViewedItem.
type RedisStore struct {
	client *redis.Client
}

// New creates a new RedisStore.
func New(client *redis.Client) *RedisStore {
	return &RedisStore{client: client}
}

// RecordView adds a viewed item to the customer's sorted set, then trims to maxItems
// and refreshes the TTL.
func (s *RedisStore) RecordView(ctx context.Context, customerID string, item domain.ViewedItem, maxItems int, ttl time.Duration) error {
	key := keyPrefix + customerID

	data, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("store.RecordView marshal: %w", err)
	}

	score := float64(item.ViewedAt.Unix())

	// Use a pipeline to ZADD + ZREMRANGEBYRANK + EXPIRE atomically.
	pipe := s.client.Pipeline()
	pipe.ZAdd(ctx, key, redis.Z{Score: score, Member: string(data)})
	// Keep only the top maxItems (highest scores = most recent).
	// After adding, index 0 is oldest; trim so only last maxItems remain.
	pipe.ZRemRangeByRank(ctx, key, 0, int64(-maxItems-1))
	pipe.Expire(ctx, key, ttl)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("store.RecordView pipeline: %w", err)
	}
	return nil
}

// GetRecent returns the most recently viewed items (highest score first), up to limit.
func (s *RedisStore) GetRecent(ctx context.Context, customerID string, limit int) ([]domain.ViewedItem, error) {
	key := keyPrefix + customerID

	// ZREVRANGE returns members from highest to lowest score.
	results, err := s.client.ZRevRange(ctx, key, 0, int64(limit-1)).Result()
	if err != nil {
		return nil, fmt.Errorf("store.GetRecent ZRevRange: %w", err)
	}

	items := make([]domain.ViewedItem, 0, len(results))
	for _, raw := range results {
		var item domain.ViewedItem
		if err := json.Unmarshal([]byte(raw), &item); err != nil {
			// Skip corrupt entries rather than failing the whole request.
			continue
		}
		items = append(items, item)
	}
	return items, nil
}

// ClearHistory removes the customer's recently-viewed sorted set.
func (s *RedisStore) ClearHistory(ctx context.Context, customerID string) error {
	key := keyPrefix + customerID
	if err := s.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("store.ClearHistory DEL: %w", err)
	}
	return nil
}

// GetCount returns the number of items in the customer's recently-viewed list.
func (s *RedisStore) GetCount(ctx context.Context, customerID string) (int, error) {
	key := keyPrefix + customerID
	count, err := s.client.ZCard(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("store.GetCount ZCard: %w", err)
	}
	return int(count), nil
}
