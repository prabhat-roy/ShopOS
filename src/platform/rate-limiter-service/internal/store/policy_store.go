package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/shopos/rate-limiter-service/internal/domain"
)

const policyPrefix = "rl:policy:"

// PolicyStore persists rate-limit policies in Redis as JSON values.
type PolicyStore struct {
	rdb *redis.Client
	ttl time.Duration
}

func NewPolicyStore(rdb *redis.Client, ttl time.Duration) *PolicyStore {
	return &PolicyStore{rdb: rdb, ttl: ttl}
}

func (s *PolicyStore) key(id string) string { return policyPrefix + id }

func (s *PolicyStore) Save(ctx context.Context, p *domain.Policy) error {
	b, err := json.Marshal(p)
	if err != nil {
		return err
	}
	return s.rdb.Set(ctx, s.key(p.ID), b, 0).Err() // policies don't expire
}

func (s *PolicyStore) Get(ctx context.Context, id string) (*domain.Policy, error) {
	b, err := s.rdb.Get(ctx, s.key(id)).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	var p domain.Policy
	return &p, json.Unmarshal(b, &p)
}

func (s *PolicyStore) List(ctx context.Context) ([]*domain.Policy, error) {
	keys, err := s.rdb.Keys(ctx, policyPrefix+"*").Result()
	if err != nil {
		return nil, err
	}
	if len(keys) == 0 {
		return []*domain.Policy{}, nil
	}
	vals, err := s.rdb.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}
	out := make([]*domain.Policy, 0, len(vals))
	for _, v := range vals {
		if v == nil {
			continue
		}
		var p domain.Policy
		if err := json.Unmarshal([]byte(v.(string)), &p); err == nil {
			out = append(out, &p)
		}
	}
	return out, nil
}

func (s *PolicyStore) Delete(ctx context.Context, id string) error {
	n, err := s.rdb.Del(ctx, s.key(id)).Result()
	if err != nil {
		return err
	}
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// CounterStore executes sliding-window and fixed-window counters in Redis.
type CounterStore struct {
	rdb *redis.Client
}

func NewCounterStore(rdb *redis.Client) *CounterStore {
	return &CounterStore{rdb: rdb}
}

// SlidingWindowAllow implements a Redis sorted-set sliding window.
// Returns (allowed, remaining, resetAfter, err).
func (s *CounterStore) SlidingWindowAllow(ctx context.Context, key string, limit int, window time.Duration, cost int) (bool, int64, int64, error) {
	now := time.Now()
	windowStart := now.Add(-window).UnixNano()
	expireAt := now.Add(window)

	pipe := s.rdb.TxPipeline()
	pipe.ZRemRangeByScore(ctx, key, "-inf", fmt.Sprintf("%d", windowStart))
	pipe.ZCard(ctx, key)
	results, err := pipe.Exec(ctx)
	if err != nil {
		return false, 0, 0, err
	}

	count := results[1].(*redis.IntCmd).Val()
	remaining := int64(limit) - count - int64(cost)
	resetAfter := int64(window.Seconds())

	if remaining < 0 {
		return false, 0, resetAfter, nil
	}

	// Add this request
	member := fmt.Sprintf("%d-%s", now.UnixNano(), key)
	s.rdb.ZAdd(ctx, key, redis.Z{Score: float64(now.UnixNano()), Member: member})
	s.rdb.ExpireAt(ctx, key, expireAt)

	return true, remaining, resetAfter, nil
}

// FixedWindowAllow implements a Redis INCR fixed window counter.
func (s *CounterStore) FixedWindowAllow(ctx context.Context, key string, limit int, window time.Duration, cost int) (bool, int64, int64, error) {
	pipe := s.rdb.TxPipeline()
	incrCmd := pipe.IncrBy(ctx, key, int64(cost))
	pipe.Expire(ctx, key, window)
	if _, err := pipe.Exec(ctx); err != nil {
		return false, 0, 0, err
	}

	count := incrCmd.Val()
	remaining := int64(limit) - count
	resetAfter := int64(window.Seconds())

	if count > int64(limit) {
		return false, 0, resetAfter, nil
	}
	return true, remaining, resetAfter, nil
}

// InMemoryTokenBucket implements a local token bucket (no Redis) for low-latency checks.
// In production this would be backed by Redis Lua script; this in-process version is
// used for unit tests and single-instance deployments.
type InMemoryTokenBucket struct {
	mu      sync.Mutex
	buckets map[string]*bucket
}

type bucket struct {
	tokens     float64
	capacity   float64
	refillRate float64 // tokens per second
	lastRefill time.Time
}

func NewInMemoryTokenBucket() *InMemoryTokenBucket {
	return &InMemoryTokenBucket{buckets: make(map[string]*bucket)}
}

func (tb *InMemoryTokenBucket) Allow(key string, limit int, burstSize int, cost int) (bool, int64) {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	capacity := float64(burstSize)
	if capacity <= 0 {
		capacity = float64(limit)
	}
	refill := float64(limit) // tokens per second = limit (1-second window)

	b, ok := tb.buckets[key]
	if !ok {
		b = &bucket{tokens: capacity, capacity: capacity, refillRate: refill, lastRefill: time.Now()}
		tb.buckets[key] = b
	}

	// Refill
	now := time.Now()
	elapsed := now.Sub(b.lastRefill).Seconds()
	b.tokens = min64(b.capacity, b.tokens+elapsed*b.refillRate)
	b.lastRefill = now

	if b.tokens < float64(cost) {
		return false, 0
	}
	b.tokens -= float64(cost)
	return true, int64(b.tokens)
}

func (tb *InMemoryTokenBucket) Cleanup(olderThan time.Duration) {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	cutoff := time.Now().Add(-olderThan)
	for k, b := range tb.buckets {
		if b.lastRefill.Before(cutoff) {
			delete(tb.buckets, k)
		}
	}
}

func min64(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
