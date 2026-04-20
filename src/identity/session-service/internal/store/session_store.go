package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/shopos/session-service/internal/domain"
)

const (
	sessionKeyPrefix   = "session:"
	userSessionsPrefix = "user_sessions:"
)

// Cacher is the minimal Redis interface required by SessionStore.
// It is defined here so tests can inject a mock without a real Redis instance.
type Cacher interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd
	SAdd(ctx context.Context, key string, members ...interface{}) *redis.IntCmd
	SRem(ctx context.Context, key string, members ...interface{}) *redis.IntCmd
	SMembers(ctx context.Context, key string) *redis.StringSliceCmd
}

// SessionStore is a Redis-backed session store.
type SessionStore struct {
	client Cacher
}

// New creates a SessionStore wrapping the provided Cacher.
func New(client Cacher) *SessionStore {
	return &SessionStore{client: client}
}

// NewRedisClient is a convenience constructor that builds a go-redis client from
// the provided address / password / DB index.
func NewRedisClient(addr, password string, db int) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
}

// sessionKey returns the Redis key for a session document.
func sessionKey(id string) string {
	return sessionKeyPrefix + id
}

// userSessionsKey returns the Redis key for the set of session IDs belonging to a user.
func userSessionsKey(userID string) string {
	return userSessionsPrefix + userID
}

// Create persists a new session in Redis. It stores the JSON-encoded session
// at key session:{id} with the given TTL, and adds the session ID to the
// user's session set.
func (s *SessionStore) Create(ctx context.Context, session *domain.Session, ttl time.Duration) error {
	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("store.Create marshal: %w", err)
	}

	if err := s.client.Set(ctx, sessionKey(session.ID), data, ttl).Err(); err != nil {
		return fmt.Errorf("store.Create SET: %w", err)
	}

	if err := s.client.SAdd(ctx, userSessionsKey(session.UserID), session.ID).Err(); err != nil {
		// Best-effort: session was stored; index update failed.
		return fmt.Errorf("store.Create SADD: %w", err)
	}

	return nil
}

// Get retrieves a session by its ID. Returns domain.ErrNotFound if the key
// does not exist in Redis.
func (s *SessionStore) Get(ctx context.Context, id string) (*domain.Session, error) {
	data, err := s.client.Get(ctx, sessionKey(id)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("store.Get GET: %w", err)
	}

	var session domain.Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("store.Get unmarshal: %w", err)
	}

	return &session, nil
}

// Touch refreshes the session's LastActiveAt timestamp and resets the Redis TTL.
func (s *SessionStore) Touch(ctx context.Context, id string, ttl time.Duration) error {
	session, err := s.Get(ctx, id)
	if err != nil {
		return err
	}

	session.LastActiveAt = time.Now().UTC()

	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("store.Touch marshal: %w", err)
	}

	if err := s.client.Set(ctx, sessionKey(id), data, ttl).Err(); err != nil {
		return fmt.Errorf("store.Touch SET: %w", err)
	}

	return nil
}

// Delete removes a session document and removes the session ID from the user's
// session set.
func (s *SessionStore) Delete(ctx context.Context, id string) error {
	// Retrieve first to obtain userID for index cleanup.
	session, err := s.Get(ctx, id)
	if err != nil {
		return err
	}

	if err := s.client.Del(ctx, sessionKey(id)).Err(); err != nil {
		return fmt.Errorf("store.Delete DEL: %w", err)
	}

	if err := s.client.SRem(ctx, userSessionsKey(session.UserID), id).Err(); err != nil {
		return fmt.Errorf("store.Delete SREM: %w", err)
	}

	return nil
}

// ListByUser returns all sessions belonging to a user. Sessions whose keys no
// longer exist in Redis (already expired) are silently skipped and their IDs
// are pruned from the set.
func (s *SessionStore) ListByUser(ctx context.Context, userID string) ([]*domain.Session, error) {
	ids, err := s.client.SMembers(ctx, userSessionsKey(userID)).Result()
	if err != nil {
		return nil, fmt.Errorf("store.ListByUser SMEMBERS: %w", err)
	}

	sessions := make([]*domain.Session, 0, len(ids))
	for _, id := range ids {
		sess, err := s.Get(ctx, id)
		if err != nil {
			if err == domain.ErrNotFound {
				// Session expired; clean up stale set member.
				_ = s.client.SRem(ctx, userSessionsKey(userID), id)
				continue
			}
			return nil, err
		}
		sessions = append(sessions, sess)
	}

	return sessions, nil
}

// DeleteAllByUser removes every active session for the specified user.
func (s *SessionStore) DeleteAllByUser(ctx context.Context, userID string) error {
	ids, err := s.client.SMembers(ctx, userSessionsKey(userID)).Result()
	if err != nil {
		return fmt.Errorf("store.DeleteAllByUser SMEMBERS: %w", err)
	}

	for _, id := range ids {
		keys := []string{sessionKey(id)}
		if err := s.client.Del(ctx, keys...).Err(); err != nil {
			return fmt.Errorf("store.DeleteAllByUser DEL %s: %w", id, err)
		}
	}

	// Remove the user sessions index key itself.
	if err := s.client.Del(ctx, userSessionsKey(userID)).Err(); err != nil {
		return fmt.Errorf("store.DeleteAllByUser DEL user set: %w", err)
	}

	return nil
}
