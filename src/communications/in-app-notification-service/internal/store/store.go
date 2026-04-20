package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/shopos/in-app-notification-service/internal/domain"
)

const (
	sortedSetKey = "notif:%s"          // notif:{userId}
	hashKey      = "notif:data:%s"     // notif:data:{notifId}
)

// Storer defines the storage operations for notifications.
type Storer interface {
	SaveNotification(ctx context.Context, userID string, notif domain.Notification, ttl time.Duration, maxPerUser int) error
	GetNotifications(ctx context.Context, userID string, includeRead bool, limit, offset int) ([]domain.Notification, error)
	MarkRead(ctx context.Context, userID, notifID string) error
	MarkAllRead(ctx context.Context, userID string) error
	DeleteNotification(ctx context.Context, userID, notifID string) error
	GetUnreadCount(ctx context.Context, userID string) (int, error)
	ClearAll(ctx context.Context, userID string) error
}

// RedisStore implements Storer using Redis sorted sets and hashes.
type RedisStore struct {
	client *redis.Client
}

// NewRedisStore creates a new RedisStore from the given redis.Client.
func NewRedisStore(client *redis.Client) *RedisStore {
	return &RedisStore{client: client}
}

// userSetKey returns the sorted-set key for a user's notification index.
func userSetKey(userID string) string {
	return fmt.Sprintf(sortedSetKey, userID)
}

// notifHashKey returns the hash key for a notification's data.
func notifHashKey(notifID string) string {
	return fmt.Sprintf(hashKey, notifID)
}

// SaveNotification stores a notification, then trims the sorted set to maxPerUser.
func (s *RedisStore) SaveNotification(ctx context.Context, userID string, notif domain.Notification, ttl time.Duration, maxPerUser int) error {
	data, err := json.Marshal(notif)
	if err != nil {
		return fmt.Errorf("marshal notification: %w", err)
	}

	pipe := s.client.Pipeline()

	// Add notifID to user's sorted set with Unix timestamp as score.
	pipe.ZAdd(ctx, userSetKey(userID), redis.Z{
		Score:  float64(notif.CreatedAt.UnixNano()),
		Member: notif.ID,
	})

	// Store notification data in a hash.
	pipe.Set(ctx, notifHashKey(notif.ID), data, ttl)

	// Refresh sorted-set TTL so it expires along with the last notification.
	pipe.Expire(ctx, userSetKey(userID), ttl)

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("pipeline exec SaveNotification: %w", err)
	}

	// Trim to maxPerUser, keeping the most recent entries (highest scores).
	trimEnd := int64(-maxPerUser - 1)
	if err := s.client.ZRemRangeByRank(ctx, userSetKey(userID), 0, trimEnd).Err(); err != nil {
		return fmt.Errorf("trim notifications: %w", err)
	}
	return nil
}

// GetNotifications retrieves paginated notifications for a user.
// If includeRead is false, only unread notifications are returned.
func (s *RedisStore) GetNotifications(ctx context.Context, userID string, includeRead bool, limit, offset int) ([]domain.Notification, error) {
	// Fetch all IDs in descending score order (newest first).
	ids, err := s.client.ZRevRange(ctx, userSetKey(userID), 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("zrevrange: %w", err)
	}

	var results []domain.Notification
	for _, id := range ids {
		raw, err := s.client.Get(ctx, notifHashKey(id)).Bytes()
		if err != nil {
			// Notification data may have expired; skip.
			continue
		}
		var n domain.Notification
		if err := json.Unmarshal(raw, &n); err != nil {
			continue
		}
		if !includeRead && n.Read {
			continue
		}
		results = append(results, n)
	}

	// Apply offset and limit.
	if offset >= len(results) {
		return []domain.Notification{}, nil
	}
	results = results[offset:]
	if limit > 0 && limit < len(results) {
		results = results[:limit]
	}
	return results, nil
}

// MarkRead marks a single notification as read by updating its stored data.
func (s *RedisStore) MarkRead(ctx context.Context, userID, notifID string) error {
	raw, err := s.client.Get(ctx, notifHashKey(notifID)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("notification not found: %s", notifID)
		}
		return fmt.Errorf("get notification: %w", err)
	}
	var n domain.Notification
	if err := json.Unmarshal(raw, &n); err != nil {
		return fmt.Errorf("unmarshal notification: %w", err)
	}
	// Verify the notification belongs to the user.
	if n.UserID != userID {
		return fmt.Errorf("notification %s does not belong to user %s", notifID, userID)
	}
	n.Read = true
	data, err := json.Marshal(n)
	if err != nil {
		return fmt.Errorf("marshal notification: %w", err)
	}
	// Preserve existing TTL.
	ttl, err := s.client.TTL(ctx, notifHashKey(notifID)).Result()
	if err != nil || ttl <= 0 {
		ttl = 30 * 24 * time.Hour
	}
	return s.client.Set(ctx, notifHashKey(notifID), data, ttl).Err()
}

// MarkAllRead marks all notifications for a user as read.
func (s *RedisStore) MarkAllRead(ctx context.Context, userID string) error {
	ids, err := s.client.ZRevRange(ctx, userSetKey(userID), 0, -1).Result()
	if err != nil {
		return fmt.Errorf("zrevrange: %w", err)
	}
	for _, id := range ids {
		if err := s.MarkRead(ctx, userID, id); err != nil {
			// Best-effort; continue on individual errors.
			continue
		}
	}
	return nil
}

// DeleteNotification removes a notification from the user's set and deletes its data.
func (s *RedisStore) DeleteNotification(ctx context.Context, userID, notifID string) error {
	pipe := s.client.Pipeline()
	pipe.ZRem(ctx, userSetKey(userID), notifID)
	pipe.Del(ctx, notifHashKey(notifID))
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("pipeline exec DeleteNotification: %w", err)
	}
	return nil
}

// GetUnreadCount returns the count of unread notifications for a user.
func (s *RedisStore) GetUnreadCount(ctx context.Context, userID string) (int, error) {
	ids, err := s.client.ZRevRange(ctx, userSetKey(userID), 0, -1).Result()
	if err != nil {
		return 0, fmt.Errorf("zrevrange: %w", err)
	}
	count := 0
	for _, id := range ids {
		raw, err := s.client.Get(ctx, notifHashKey(id)).Bytes()
		if err != nil {
			continue
		}
		var n domain.Notification
		if err := json.Unmarshal(raw, &n); err != nil {
			continue
		}
		if !n.Read {
			count++
		}
	}
	return count, nil
}

// ClearAll removes all notifications for a user.
func (s *RedisStore) ClearAll(ctx context.Context, userID string) error {
	ids, err := s.client.ZRevRange(ctx, userSetKey(userID), 0, -1).Result()
	if err != nil {
		return fmt.Errorf("zrevrange: %w", err)
	}
	pipe := s.client.Pipeline()
	for _, id := range ids {
		pipe.Del(ctx, notifHashKey(id))
	}
	pipe.Del(ctx, userSetKey(userID))
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("pipeline exec ClearAll: %w", err)
	}
	return nil
}
