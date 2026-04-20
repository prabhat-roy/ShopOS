package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/shopos/mfa-service/internal/config"
	"github.com/shopos/mfa-service/internal/handler"
	"github.com/shopos/mfa-service/internal/service"
	"github.com/shopos/mfa-service/internal/store"
)

func main() {
	logger := log.New(os.Stdout, "[mfa-service] ", log.LstdFlags|log.Lmsgprefix)

	cfg, err := config.Load()
	if err != nil {
		logger.Fatalf("config: %v", err)
	}

	// ── Persistence ────────────────────────────────────────────────────────────
	mfaStore, err := store.New(cfg.DatabaseURL)
	if err != nil {
		logger.Fatalf("store: %v", err)
	}
	defer func() {
		if err := mfaStore.Close(); err != nil {
			logger.Printf("store close: %v", err)
		}
	}()

	// ── Service ────────────────────────────────────────────────────────────────
	mfaSvc := service.New(mfaStore, "ShopOS")

	// ── HTTP server ────────────────────────────────────────────────────────────
	h := handler.New(mfaSvc, logger)

	httpServer := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      h,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start HTTP server in background.
	go func() {
		logger.Printf("HTTP server listening on :%s", cfg.HTTPPort)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatalf("HTTP server error: %v", err)
		}
	}()

	// ── Graceful shutdown ──────────────────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Println("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Printf("HTTP shutdown error: %v", err)
	}

	logger.Println("stopped")
}
