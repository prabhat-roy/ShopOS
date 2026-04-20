package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"github.com/shopos/rate-limiter-service/internal/config"
	"github.com/shopos/rate-limiter-service/internal/handler"
	"github.com/shopos/rate-limiter-service/internal/service"
	"github.com/shopos/rate-limiter-service/internal/store"
	"go.uber.org/zap"
)

func main() {
	_ = godotenv.Load()

	cfg := config.Load()

	var logger *zap.Logger
	if cfg.Env == "production" {
		logger, _ = zap.NewProduction()
	} else {
		logger, _ = zap.NewDevelopment()
	}
	defer logger.Sync()

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPass,
		DB:       cfg.RedisDB,
	})
	defer rdb.Close()

	tb := store.NewInMemoryTokenBucket()

	// Periodic cleanup of stale token-bucket entries
	go func() {
		ticker := time.NewTicker(time.Duration(cfg.CleanupSecs) * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			tb.Cleanup(time.Duration(cfg.CleanupSecs) * time.Second)
		}
	}()

	ps := store.NewPolicyStore(rdb, cfg.KeyTTL)
	cs := store.NewCounterStore(rdb)
	svc := service.New(ps, cs, tb, logger)
	h := handler.New(svc)

	mux := http.NewServeMux()
	h.Register(mux)

	srv := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("rate-limiter-service listening", zap.String("port", cfg.HTTPPort))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("shutting down rate-limiter-service...")

	shutCtx, shutCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutCancel()
	if err := srv.Shutdown(shutCtx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}
	logger.Info("rate-limiter-service stopped")
}
