package service

import (
	"context"
	"errors"

	pb "github.com/shopos/config-service/internal/proto"
	"github.com/shopos/config-service/internal/store"
	"go.uber.org/zap"
)

var ErrNotFound = errors.New("config key not found")

// ConfigService implements the business logic for config operations.
type ConfigService struct {
	store *store.Store
	log   *zap.Logger
}

func New(st *store.Store, log *zap.Logger) *ConfigService {
	return &ConfigService{store: st, log: log}
}

func (s *ConfigService) Get(ctx context.Context, key string) (*pb.ConfigEntry, error) {
	entry, found, err := s.store.Get(ctx, key)
	if err != nil {
		s.log.Error("store.Get failed", zap.String("key", key), zap.Error(err))
		return nil, err
	}
	if !found {
		return nil, ErrNotFound
	}
	return entry, nil
}

func (s *ConfigService) Set(ctx context.Context, key, value string) error {
	if key == "" {
		return errors.New("key must not be empty")
	}
	if err := s.store.Set(ctx, key, value); err != nil {
		s.log.Error("store.Set failed", zap.String("key", key), zap.Error(err))
		return err
	}
	s.log.Info("config key set", zap.String("key", key))
	return nil
}

func (s *ConfigService) Delete(ctx context.Context, key string) error {
	if err := s.store.Delete(ctx, key); err != nil {
		s.log.Error("store.Delete failed", zap.String("key", key), zap.Error(err))
		return err
	}
	s.log.Info("config key deleted", zap.String("key", key))
	return nil
}

func (s *ConfigService) List(ctx context.Context, prefix string) ([]*pb.ConfigEntry, error) {
	entries, err := s.store.List(ctx, prefix)
	if err != nil {
		s.log.Error("store.List failed", zap.String("prefix", prefix), zap.Error(err))
		return nil, err
	}
	return entries, nil
}

func (s *ConfigService) Watch(ctx context.Context, prefix string) (<-chan *pb.WatchEvent, error) {
	return s.store.WatchCh(ctx, prefix)
}
