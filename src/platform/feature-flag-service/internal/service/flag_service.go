package service

import (
	"context"
	"hash/fnv"

	"github.com/shopos/feature-flag-service/internal/domain"
	"go.uber.org/zap"
)

// Storer is the persistence interface, satisfied by store.Store.
type Storer interface {
	GetByKey(ctx context.Context, key string) (*domain.Flag, error)
	GetByID(ctx context.Context, id string) (*domain.Flag, error)
	List(ctx context.Context) ([]*domain.Flag, error)
	Create(ctx context.Context, req *domain.CreateFlagRequest) (*domain.Flag, error)
	Update(ctx context.Context, id string, req *domain.UpdateFlagRequest) (*domain.Flag, error)
	Delete(ctx context.Context, id string) error
}

// FlagService contains all feature-flag business logic.
type FlagService struct {
	store Storer
	log   *zap.Logger
}

func New(st Storer, log *zap.Logger) *FlagService {
	return &FlagService{store: st, log: log}
}

func (s *FlagService) GetFlag(ctx context.Context, key string) (*domain.Flag, error) {
	return s.store.GetByKey(ctx, key)
}

func (s *FlagService) ListFlags(ctx context.Context) ([]*domain.Flag, error) {
	return s.store.List(ctx)
}

func (s *FlagService) CreateFlag(ctx context.Context, req *domain.CreateFlagRequest) (*domain.Flag, error) {
	if req.Key == "" || req.Name == "" {
		return nil, domain.ErrInvalidInput
	}
	if req.Strategy == "" {
		req.Strategy = domain.StrategyAll
	}
	if req.Strategy == domain.StrategyPercentage && (req.Percentage < 0 || req.Percentage > 100) {
		return nil, domain.ErrInvalidInput
	}
	return s.store.Create(ctx, req)
}

func (s *FlagService) UpdateFlag(ctx context.Context, id string, req *domain.UpdateFlagRequest) (*domain.Flag, error) {
	if req.Percentage != nil && (*req.Percentage < 0 || *req.Percentage > 100) {
		return nil, domain.ErrInvalidInput
	}
	return s.store.Update(ctx, id, req)
}

func (s *FlagService) DeleteFlag(ctx context.Context, id string) error {
	return s.store.Delete(ctx, id)
}

// Evaluate determines whether a flag is active for the given request context.
func (s *FlagService) Evaluate(ctx context.Context, req domain.EvalRequest) (bool, error) {
	flag, err := s.store.GetByKey(ctx, req.Key)
	if err != nil {
		return false, err
	}
	if !flag.Enabled {
		return false, nil
	}

	switch flag.Strategy {
	case domain.StrategyAll:
		return true, nil

	case domain.StrategyPercentage:
		// Deterministic: hash(userID + flagKey) mod 100 < percentage
		h := fnv.New32a()
		h.Write([]byte(req.UserID + ":" + req.Key))
		bucket := int(h.Sum32() % 100)
		return bucket < flag.Percentage, nil

	case domain.StrategyUserList:
		for _, uid := range flag.UserIDs {
			if uid == req.UserID {
				return true, nil
			}
		}
		return false, nil

	case domain.StrategyContext:
		v, ok := req.Context[flag.ContextKey]
		return ok && v == flag.ContextVal, nil

	default:
		return false, nil
	}
}
