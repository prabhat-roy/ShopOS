package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/shopos/in-app-notification-service/internal/config"
	"github.com/shopos/in-app-notification-service/internal/handler"
	"github.com/shopos/in-app-notification-service/internal/service"
	"github.com/shopos/in-app-notification-service/internal/store"
)

func main() {
	cfg := config.Load()

	// Connect to Redis.
	opts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Fatalf("invalid REDIS_URL %q: %v", cfg.RedisURL, err)
	}
	redisClient := redis.NewClient(opts)

	// Wire dependencies.
	redisStore := store.NewRedisStore(redisClient)
	svc := service.New(redisStore, cfg)
	h := handler.New(svc)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	httpAddr := fmt.Sprintf(":%s", cfg.HTTPPort)
	httpServer := &http.Server{
		Addr:         httpAddr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start HTTP server in a goroutine.
	go func() {
		log.Printf("HTTP server listening on %s", httpAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// gRPC port is reserved for future implementation (health probe etc.).
	log.Printf("gRPC port reserved on :%s (not yet implemented)", cfg.GRPCPort)

	// Graceful shutdown on SIGINT / SIGTERM.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down in-app-notification-service...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server forced shutdown: %v", err)
	}
	if err := redisClient.Close(); err != nil {
		log.Printf("Redis close error: %v", err)
	}
	log.Println("in-app-notification-service stopped")
}
