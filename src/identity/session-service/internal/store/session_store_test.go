package store_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/shopos/session-service/internal/domain"
	"github.com/shopos/session-service/internal/store"
)

// --------------------------------------------------------------------------
// Minimal in-memory mock that satisfies store.Cacher
// --------------------------------------------------------------------------

type mockRedis struct {
	data    map[string][]byte
	ttls    map[string]time.Duration
	sets    map[string]map[string]struct{}
}

func newMockRedis() *mockRedis {
	return &mockRedis{
		data: make(map[string][]byte),
		ttls: make(map[string]time.Duration),
		sets: make(map[string]map[string]struct{}),
	}
}

func (m *mockRedis) Set(_ context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	var b []byte
	switch v := value.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		b, _ = json.Marshal(v)
	}
	m.data[key] = b
	m.ttls[key] = expiration
	cmd := redis.NewStatusCmd(context.Background())
	cmd.SetVal("OK")
	return cmd
}

func (m *mockRedis) Get(_ context.Context, key string) *redis.StringCmd {
	cmd := redis.NewStringCmd(context.Background())
	v, ok := m.data[key]
	if !ok {
		cmd.SetErr(redis.Nil)
		return cmd
	}
	cmd.SetVal(string(v))
	return cmd
}

func (m *mockRedis) Del(_ context.Context, keys ...string) *redis.IntCmd {
	cmd := redis.NewIntCmd(context.Background())
	var count int64
	for _, k := range keys {
		if _, ok := m.data[k]; ok {
			delete(m.data, k)
			delete(m.ttls, k)
			count++
		}
	}
	cmd.SetVal(count)
	return cmd
}

func (m *mockRedis) Expire(_ context.Context, key string, expiration time.Duration) *redis.BoolCmd {
	cmd := redis.NewBoolCmd(context.Background())
	if _, ok := m.data[key]; ok {
		m.ttls[key] = expiration
		cmd.SetVal(true)
		return cmd
	}
	cmd.SetVal(false)
	return cmd
}

func (m *mockRedis) SAdd(_ context.Context, key string, members ...interface{}) *redis.IntCmd {
	cmd := redis.NewIntCmd(context.Background())
	if m.sets[key] == nil {
		m.sets[key] = make(map[string]struct{})
	}
	for _, member := range members {
		m.sets[key][member.(string)] = struct{}{}
	}
	cmd.SetVal(int64(len(members)))
	return cmd
}

func (m *mockRedis) SRem(_ context.Context, key string, members ...interface{}) *redis.IntCmd {
	cmd := redis.NewIntCmd(context.Background())
	if m.sets[key] == nil {
		cmd.SetVal(0)
		return cmd
	}
	var count int64
	for _, member := range members {
		if _, ok := m.sets[key][member.(string)]; ok {
			delete(m.sets[key], member.(string))
			count++
		}
	}
	cmd.SetVal(count)
	return cmd
}

func (m *mockRedis) SMembers(_ context.Context, key string) *redis.StringSliceCmd {
	cmd := redis.NewStringSliceCmd(context.Background())
	s := m.sets[key]
	result := make([]string, 0, len(s))
	for k := range s {
		result = append(result, k)
	}
	cmd.SetVal(result)
	return cmd
}

// --------------------------------------------------------------------------
// Helpers
// --------------------------------------------------------------------------

func makeSession(id, userID string) *domain.Session {
	now := time.Now().UTC()
	return &domain.Session{
		ID:           id,
		UserID:       userID,
		DeviceInfo:   "Chrome/120",
		IPAddress:    "127.0.0.1",
		UserAgent:    "Mozilla/5.0",
		CreatedAt:    now,
		LastActiveAt: now,
		ExpiresAt:    now.Add(24 * time.Hour),
	}
}

// --------------------------------------------------------------------------
// Tests
// --------------------------------------------------------------------------

func TestCreate_StoresSession(t *testing.T) {
	ctx := context.Background()
	mock := newMockRedis()
	s := store.New(mock)

	sess := makeSession("sess-1", "user-42")
	if err := s.Create(ctx, sess, 24*time.Hour); err != nil {
		t.Fatalf("Create returned unexpected error: %v", err)
	}

	// Verify session key exists in mock storage.
	got, err := s.Get(ctx, "sess-1")
	if err != nil {
		t.Fatalf("Get after Create returned unexpected error: %v", err)
	}
	if got.UserID != "user-42" {
		t.Errorf("expected UserID user-42, got %s", got.UserID)
	}

	// Verify user sessions set was populated.
	members := mock.sets["user_sessions:user-42"]
	if _, ok := members["sess-1"]; !ok {
		t.Error("expected sess-1 in user_sessions:user-42 set")
	}
}

func TestGet_ReturnsErrNotFound_ForMissingKey(t *testing.T) {
	ctx := context.Background()
	s := store.New(newMockRedis())

	_, err := s.Get(ctx, "nonexistent")
	if err != domain.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestDelete_RemovesSessionAndUserIndex(t *testing.T) {
	ctx := context.Background()
	mock := newMockRedis()
	s := store.New(mock)

	sess := makeSession("sess-del", "user-99")
	if err := s.Create(ctx, sess, time.Hour); err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := s.Delete(ctx, "sess-del"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	// Session key must be gone.
	if _, err := s.Get(ctx, "sess-del"); err != domain.ErrNotFound {
		t.Errorf("expected ErrNotFound after Delete, got %v", err)
	}

	// User sessions set must not contain the ID.
	if _, ok := mock.sets["user_sessions:user-99"]["sess-del"]; ok {
		t.Error("session ID still present in user set after Delete")
	}
}

func TestListByUser_ReturnsAllActiveSessions(t *testing.T) {
	ctx := context.Background()
	s := store.New(newMockRedis())

	for _, id := range []string{"s1", "s2", "s3"} {
		if err := s.Create(ctx, makeSession(id, "user-list"), time.Hour); err != nil {
			t.Fatalf("Create %s: %v", id, err)
		}
	}

	sessions, err := s.ListByUser(ctx, "user-list")
	if err != nil {
		t.Fatalf("ListByUser: %v", err)
	}
	if len(sessions) != 3 {
		t.Errorf("expected 3 sessions, got %d", len(sessions))
	}
}

func TestDeleteAllByUser_RemovesEverything(t *testing.T) {
	ctx := context.Background()
	mock := newMockRedis()
	s := store.New(mock)

	for _, id := range []string{"a1", "a2"} {
		if err := s.Create(ctx, makeSession(id, "user-all"), time.Hour); err != nil {
			t.Fatalf("Create %s: %v", id, err)
		}
	}

	if err := s.DeleteAllByUser(ctx, "user-all"); err != nil {
		t.Fatalf("DeleteAllByUser: %v", err)
	}

	sessions, err := s.ListByUser(ctx, "user-all")
	if err != nil {
		t.Fatalf("ListByUser after DeleteAll: %v", err)
	}
	if len(sessions) != 0 {
		t.Errorf("expected 0 sessions after DeleteAll, got %d", len(sessions))
	}
}

func TestTouch_UpdatesLastActiveAt(t *testing.T) {
	ctx := context.Background()
	s := store.New(newMockRedis())

	sess := makeSession("touch-sess", "user-touch")
	original := sess.LastActiveAt
	if err := s.Create(ctx, sess, time.Hour); err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Ensure at least 1 ns passes so timestamps differ.
	time.Sleep(time.Millisecond)

	if err := s.Touch(ctx, "touch-sess", time.Hour); err != nil {
		t.Fatalf("Touch: %v", err)
	}

	updated, err := s.Get(ctx, "touch-sess")
	if err != nil {
		t.Fatalf("Get after Touch: %v", err)
	}
	if !updated.LastActiveAt.After(original) {
		t.Errorf("expected LastActiveAt to advance; original=%v updated=%v", original, updated.LastActiveAt)
	}
}
