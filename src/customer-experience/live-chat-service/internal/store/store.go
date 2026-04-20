package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/shopos/live-chat-service/internal/domain"
)

const (
	sessionKeyFmt      = "chat:session:%s"
	messagesKeyFmt     = "chat:messages:%s"
	activeSessionsKey  = "chat:active_sessions"
	waitingSessionsKey = "chat:waiting_sessions"
)

// RedisStore provides persistence for chat sessions and messages backed by Redis.
type RedisStore struct {
	client     *redis.Client
	sessionTTL time.Duration
	maxMessages int
}

// New creates a new RedisStore connected to the given Redis URL.
func New(redisURL string, sessionTTL time.Duration, maxMessages int) (*RedisStore, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("invalid redis url: %w", err)
	}
	client := redis.NewClient(opts)
	return &RedisStore{
		client:      client,
		sessionTTL:  sessionTTL,
		maxMessages: maxMessages,
	}, nil
}

// CreateSession persists a new chat session and adds it to the appropriate index.
func (s *RedisStore) CreateSession(ctx context.Context, session domain.ChatSession) error {
	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("marshal session: %w", err)
	}
	key := fmt.Sprintf(sessionKeyFmt, session.ID)
	pipe := s.client.TxPipeline()
	pipe.Set(ctx, key, data, s.sessionTTL)
	score := float64(session.CreatedAt.UnixNano())
	if session.Status == domain.StatusWaiting {
		pipe.ZAdd(ctx, waitingSessionsKey, redis.Z{Score: score, Member: session.ID})
	}
	_, err = pipe.Exec(ctx)
	return err
}

// GetSession retrieves a session by its ID.
func (s *RedisStore) GetSession(ctx context.Context, id string) (domain.ChatSession, error) {
	key := fmt.Sprintf(sessionKeyFmt, id)
	data, err := s.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return domain.ChatSession{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.ChatSession{}, fmt.Errorf("redis get: %w", err)
	}
	var session domain.ChatSession
	if err := json.Unmarshal(data, &session); err != nil {
		return domain.ChatSession{}, fmt.Errorf("unmarshal session: %w", err)
	}
	return session, nil
}

// UpdateSession overwrites an existing session and updates index membership.
func (s *RedisStore) UpdateSession(ctx context.Context, session domain.ChatSession) error {
	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("marshal session: %w", err)
	}
	key := fmt.Sprintf(sessionKeyFmt, session.ID)
	pipe := s.client.TxPipeline()
	pipe.Set(ctx, key, data, s.sessionTTL)
	// Remove from waiting set whenever session status changes
	pipe.ZRem(ctx, waitingSessionsKey, session.ID)
	if session.Status == domain.StatusWaiting {
		score := float64(session.CreatedAt.UnixNano())
		pipe.ZAdd(ctx, waitingSessionsKey, redis.Z{Score: score, Member: session.ID})
	}
	_, err = pipe.Exec(ctx)
	return err
}

// SaveMessage appends a message to the session's message list, capping at maxMessages.
func (s *RedisStore) SaveMessage(ctx context.Context, msg domain.ChatMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}
	key := fmt.Sprintf(messagesKeyFmt, msg.SessionID)
	pipe := s.client.TxPipeline()
	pipe.RPush(ctx, key, data)
	pipe.LTrim(ctx, key, int64(-s.maxMessages), -1)
	pipe.Expire(ctx, key, s.sessionTTL)
	_, err = pipe.Exec(ctx)
	return err
}

// GetMessages retrieves the latest `limit` messages for a session.
func (s *RedisStore) GetMessages(ctx context.Context, sessionID string, limit int) ([]domain.ChatMessage, error) {
	key := fmt.Sprintf(messagesKeyFmt, sessionID)
	raw, err := s.client.LRange(ctx, key, int64(-limit), -1).Result()
	if err != nil {
		return nil, fmt.Errorf("redis lrange: %w", err)
	}
	messages := make([]domain.ChatMessage, 0, len(raw))
	for _, r := range raw {
		var msg domain.ChatMessage
		if err := json.Unmarshal([]byte(r), &msg); err != nil {
			continue
		}
		messages = append(messages, msg)
	}
	return messages, nil
}

// ListWaitingSessions returns all sessions currently in "waiting" status.
func (s *RedisStore) ListWaitingSessions(ctx context.Context) ([]domain.ChatSession, error) {
	ids, err := s.client.ZRangeByScore(ctx, waitingSessionsKey, &redis.ZRangeBy{
		Min: "-inf",
		Max: "+inf",
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("redis zrangebyscore: %w", err)
	}
	sessions := make([]domain.ChatSession, 0, len(ids))
	for _, id := range ids {
		sess, err := s.GetSession(ctx, id)
		if err != nil {
			// Session expired or removed — clean up the index entry
			s.client.ZRem(ctx, waitingSessionsKey, id) //nolint:errcheck
			continue
		}
		sessions = append(sessions, sess)
	}
	return sessions, nil
}
