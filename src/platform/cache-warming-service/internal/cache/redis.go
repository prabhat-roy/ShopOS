package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cacher is the Redis operations subset used by the warmer.
type Cacher interface {
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Incr(ctx context.Context, key string) (int64, error)
	Expire(ctx context.Context, key string, ttl time.Duration) error
}

type redisCache struct {
	rdb *redis.Client
}

func New(addr, password string, db int) Cacher {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	return &redisCache{rdb: rdb}
}

func (c *redisCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	return c.rdb.Set(ctx, key, value, ttl).Err()
}

func (c *redisCache) Incr(ctx context.Context, key string) (int64, error) {
	return c.rdb.Incr(ctx, key).Result()
}

func (c *redisCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return c.rdb.Expire(ctx, key, ttl).Err()
}
