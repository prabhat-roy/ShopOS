package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/shopos/compare-service/internal/domain"
)

const keyPrefix = "compare:"

// Storer defines the persistence contract for compare list operations.
type Storer interface {
	GetList(ctx context.Context, customerID string) (*domain.CompareList, error)
	SaveList(ctx context.Context, customerID string, list *domain.CompareList, ttl time.Duration) error
	ClearList(ctx context.Context, customerID string) error
}

// RedisStore implements Storer using Redis.
type RedisStore struct {
	client *redis.Client
}

// New creates a new RedisStore with the given Redis client.
func New(client *redis.Client) *RedisStore {
	return &RedisStore{client: client}
}

// GetList retrieves the compare list for a customer from Redis.
// Returns an empty CompareList (not an error) if the key does not exist.
func (s *RedisStore) GetList(ctx context.Context, customerID string) (*domain.CompareList, error) {
	key := keyPrefix + customerID
	data, err := s.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			// Return an empty list for new customers.
			return &domain.CompareList{
				CustomerID: customerID,
				Items:      []domain.CompareItem{},
				UpdatedAt:  time.Now().UTC(),
			}, nil
		}
		return nil, fmt.Errorf("store.GetList redis GET: %w", err)
	}

	var list domain.CompareList
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, fmt.Errorf("store.GetList unmarshal: %w", err)
	}
	return &list, nil
}

// SaveList serialises the compare list to JSON and stores it in Redis with the given TTL.
func (s *RedisStore) SaveList(ctx context.Context, customerID string, list *domain.CompareList, ttl time.Duration) error {
	key := keyPrefix + customerID
	data, err := json.Marshal(list)
	if err != nil {
		return fmt.Errorf("store.SaveList marshal: %w", err)
	}
	if err := s.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("store.SaveList redis SET: %w", err)
	}
	return nil
}

// ClearList deletes the compare list key for a customer.
func (s *RedisStore) ClearList(ctx context.Context, customerID string) error {
	key := keyPrefix + customerID
	if err := s.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("store.ClearList redis DEL: %w", err)
	}
	return nil
}
