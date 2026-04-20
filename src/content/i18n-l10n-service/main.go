package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/shopos/i18n-l10n-service/internal/config"
	"github.com/shopos/i18n-l10n-service/internal/handler"
	"github.com/shopos/i18n-l10n-service/internal/service"
	"github.com/shopos/i18n-l10n-service/internal/store"
)

func main() {
	cfg := config.Load()

	log.Printf("[main] Starting i18n-l10n-service HTTP=%s gRPC=%s", cfg.HTTPPort, cfg.GRPCPort)

	// Connect to database
	st, err := store.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("[main] Failed to connect to database: %v", err)
	}
	defer st.Close()
	log.Printf("[main] Connected to PostgreSQL")

	// Create service layer
	svc := service.New(st, cfg.DefaultLocale)

	// Create HTTP handler
	h := handler.New(svc)

	srv := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      h,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start HTTP server in background
	go func() {
		log.Printf("[main] HTTP server listening on :%s", cfg.HTTPPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[main] HTTP server error: %v", err)
		}
	}()

	// Graceful shutdown on SIGINT/SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("[main] Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("[main] Server forced shutdown: %v", err)
	}
	log.Println("[main] Server exited gracefully")
}
