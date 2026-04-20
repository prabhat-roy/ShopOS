package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/shopos/rate-limiter-service/internal/domain"
	"github.com/shopos/rate-limiter-service/internal/store"
	"go.uber.org/zap"
)

// PolicyStorer is satisfied by store.PolicyStore.
type PolicyStorer interface {
	Save(ctx context.Context, p *domain.Policy) error
	Get(ctx context.Context, id string) (*domain.Policy, error)
	List(ctx context.Context) ([]*domain.Policy, error)
	Delete(ctx context.Context, id string) error
}

// CounterStorer is satisfied by store.CounterStore.
type CounterStorer interface {
	SlidingWindowAllow(ctx context.Context, key string, limit int, window time.Duration, cost int) (bool, int64, int64, error)
	FixedWindowAllow(ctx context.Context, key string, limit int, window time.Duration, cost int) (bool, int64, int64, error)
}

// LimiterService handles all rate-limiting business logic.
type LimiterService struct {
	policies PolicyStorer
	counters CounterStorer
	tb       *store.InMemoryTokenBucket
	// policyCache is a simple in-process cache: policyKey -> *Policy
	mu          sync.RWMutex
	policyCache map[string]*domain.Policy
	log         *zap.Logger
}

func New(ps PolicyStorer, cs CounterStorer, tb *store.InMemoryTokenBucket, log *zap.Logger) *LimiterService {
	return &LimiterService{
		policies:    ps,
		counters:    cs,
		tb:          tb,
		policyCache: make(map[string]*domain.Policy),
		log:         log,
	}
}

// CreatePolicy stores a new rate-limit policy and caches it.
func (s *LimiterService) CreatePolicy(ctx context.Context, req *domain.CreatePolicyRequest) (*domain.Policy, error) {
	if req.Key == "" || req.Name == "" {
		return nil, domain.ErrInvalidInput
	}
	if req.Limit <= 0 || req.WindowSecs <= 0 {
		return nil, domain.ErrInvalidInput
	}
	if req.Algorithm == "" {
		req.Algorithm = domain.AlgoSlidingWindow
	}

	p := &domain.Policy{
		ID:         uuid.New().String(),
		Name:       req.Name,
		Key:        req.Key,
		Algorithm:  req.Algorithm,
		Limit:      req.Limit,
		WindowSecs: req.WindowSecs,
		BurstSize:  req.BurstSize,
		Enabled:    req.Enabled,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}
	if err := s.policies.Save(ctx, p); err != nil {
		return nil, err
	}
	s.cacheSet(p.Key, p)
	return p, nil
}

// GetPolicy retrieves a policy by ID.
func (s *LimiterService) GetPolicy(ctx context.Context, id string) (*domain.Policy, error) {
	return s.policies.Get(ctx, id)
}

// ListPolicies returns all stored policies.
func (s *LimiterService) ListPolicies(ctx context.Context) ([]*domain.Policy, error) {
	return s.policies.List(ctx)
}

// UpdatePolicy applies partial updates to an existing policy.
func (s *LimiterService) UpdatePolicy(ctx context.Context, id string, req *domain.UpdatePolicyRequest) (*domain.Policy, error) {
	p, err := s.policies.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.Enabled != nil {
		p.Enabled = *req.Enabled
	}
	if req.Limit != nil {
		if *req.Limit <= 0 {
			return nil, domain.ErrInvalidInput
		}
		p.Limit = *req.Limit
	}
	if req.WindowSecs != nil {
		if *req.WindowSecs <= 0 {
			return nil, domain.ErrInvalidInput
		}
		p.WindowSecs = *req.WindowSecs
	}
	if req.BurstSize != nil {
		p.BurstSize = *req.BurstSize
	}
	p.UpdatedAt = time.Now().UTC()

	if err := s.policies.Save(ctx, p); err != nil {
		return nil, err
	}
	s.cacheSet(p.Key, p)
	return p, nil
}

// DeletePolicy removes a policy by ID.
func (s *LimiterService) DeletePolicy(ctx context.Context, id string) error {
	p, err := s.policies.Get(ctx, id)
	if err != nil {
		return err
	}
	s.cacheDel(p.Key)
	return s.policies.Delete(ctx, id)
}

// Check evaluates whether the request is within the rate limit for the matching policy.
func (s *LimiterService) Check(ctx context.Context, req domain.CheckRequest) (*domain.CheckResponse, error) {
	if req.PolicyKey == "" || req.Subject == "" {
		return nil, domain.ErrInvalidInput
	}
	cost := req.Cost
	if cost <= 0 {
		cost = 1
	}

	policy, err := s.resolvePolicy(ctx, req.PolicyKey)
	if err != nil {
		return nil, err
	}
	if !policy.Enabled {
		// Disabled policies always allow
		return &domain.CheckResponse{Allowed: true, Remaining: int64(policy.Limit)}, nil
	}

	window := time.Duration(policy.WindowSecs) * time.Second
	redisKey := fmt.Sprintf("rl:counter:%s:%s", policy.Key, req.Subject)

	switch policy.Algorithm {
	case domain.AlgoTokenBucket:
		tbKey := fmt.Sprintf("%s:%s", policy.Key, req.Subject)
		allowed, remaining := s.tb.Allow(tbKey, policy.Limit, policy.BurstSize, cost)
		return &domain.CheckResponse{
			Allowed:   allowed,
			Remaining: remaining,
			RetryAfter: func() int64 {
				if !allowed {
					return 1
				}
				return 0
			}(),
		}, nil

	case domain.AlgoFixedWindow:
		allowed, remaining, resetAfter, err := s.counters.FixedWindowAllow(ctx, redisKey, policy.Limit, window, cost)
		if err != nil {
			return nil, err
		}
		return &domain.CheckResponse{
			Allowed:    allowed,
			Remaining:  remaining,
			ResetAfter: resetAfter,
			RetryAfter: func() int64 {
				if !allowed {
					return resetAfter
				}
				return 0
			}(),
		}, nil

	default: // sliding_window
		allowed, remaining, resetAfter, err := s.counters.SlidingWindowAllow(ctx, redisKey, policy.Limit, window, cost)
		if err != nil {
			return nil, err
		}
		return &domain.CheckResponse{
			Allowed:    allowed,
			Remaining:  remaining,
			ResetAfter: resetAfter,
			RetryAfter: func() int64 {
				if !allowed {
					return resetAfter
				}
				return 0
			}(),
		}, nil
	}
}

func (s *LimiterService) resolvePolicy(ctx context.Context, key string) (*domain.Policy, error) {
	// Check in-process cache first
	if p := s.cacheGet(key); p != nil {
		return p, nil
	}
	// Fall back to listing all and finding by key
	policies, err := s.policies.List(ctx)
	if err != nil {
		return nil, err
	}
	for _, p := range policies {
		s.cacheSet(p.Key, p)
		if p.Key == key {
			return p, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (s *LimiterService) cacheGet(key string) *domain.Policy {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.policyCache[key]
}

func (s *LimiterService) cacheSet(key string, p *domain.Policy) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.policyCache[key] = p
}

func (s *LimiterService) cacheDel(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.policyCache, key)
}
