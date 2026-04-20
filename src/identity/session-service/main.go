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

	"github.com/shopos/session-service/internal/config"
	"github.com/shopos/session-service/internal/handler"
	"github.com/shopos/session-service/internal/service"
	"github.com/shopos/session-service/internal/store"
)

func main() {
	cfg := config.Load()

	// -------------------------------------------------------------------------
	// Redis client
	// -------------------------------------------------------------------------
	redisClient := store.NewRedisClient(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)

	// -------------------------------------------------------------------------
	// Wiring
	// -------------------------------------------------------------------------
	sessionStore := store.New(redisClient)
	sessionService := service.New(sessionStore, cfg.SessionTTL)
	httpHandler := handler.New(sessionService)

	// -------------------------------------------------------------------------
	// HTTP server
	// -------------------------------------------------------------------------
	httpAddr := fmt.Sprintf(":%s", cfg.HTTPPort)
	httpServer := &http.Server{
		Addr:         httpAddr,
		Handler:      httpHandler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start listening in a goroutine so we can handle shutdown signals.
	go func() {
		log.Printf("HTTP server listening on %s", httpAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// -------------------------------------------------------------------------
	// Graceful shutdown
	// -------------------------------------------------------------------------
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	log.Printf("received signal %s — shutting down", sig)

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP shutdown error: %v", err)
	}

	if err := redisClient.Close(); err != nil {
		log.Printf("Redis close error: %v", err)
	}

	log.Println("session-service stopped")
}
