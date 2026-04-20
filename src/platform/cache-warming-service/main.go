package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/shopos/cache-warming-service/internal/cache"
	"github.com/shopos/cache-warming-service/internal/config"
	"github.com/shopos/cache-warming-service/internal/kafka"
	"github.com/shopos/cache-warming-service/internal/warmer"
)

func main() {
	cfg := config.Load()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	rdb := cache.New(cfg.RedisAddr, cfg.RedisPassword, cfg.RDBDB)
	w := warmer.New(rdb, cfg.DefaultTTL)

	consumer := kafka.NewConsumer(cfg.Brokers, cfg.GroupID)
	consumer.Register(cfg.TopicProductViewed, w.HandleProductViewed)
	consumer.Register(cfg.TopicCartAbandoned, w.HandleCartAbandoned)
	consumer.Register(cfg.TopicOrderPlaced, w.HandleOrderPlaced)
	consumer.Register(cfg.TopicInventoryLow, w.HandleInventoryLow)
	consumer.Register(cfg.TopicSearchPerformed, w.HandleSearchPerformed)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	slog.Info("cache-warming-service starting", "brokers", cfg.Brokers)
loop:
	for {
		if err := consumer.Run(ctx); err != nil {
			if ctx.Err() != nil {
				break loop
			}
			slog.Warn("consumer exited, retrying in 5s", "err", err)
			select {
			case <-ctx.Done():
				break loop
			case <-time.After(5 * time.Second):
			}
		} else {
			break loop
		}
	}
	slog.Info("cache-warming-service stopped")
}
