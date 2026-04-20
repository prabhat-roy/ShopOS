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

	"github.com/shopos/tax-provider-integration/internal/adapter"
	"github.com/shopos/tax-provider-integration/internal/config"
	"github.com/shopos/tax-provider-integration/internal/handler"
	"github.com/shopos/tax-provider-integration/internal/service"
)

func main() {
	cfg := config.Load()

	// Wire dependencies.
	adp := adapter.New()
	svc := service.New(adp)
	h := handler.New(svc)

	// HTTP server.
	httpAddr := fmt.Sprintf(":%d", cfg.HTTPPort)
	srv := &http.Server{
		Addr:         httpAddr,
		Handler:      h,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start HTTP server in background.
	go func() {
		log.Printf("tax-provider-integration HTTP server listening on %s (default provider: %s)",
			httpAddr, cfg.DefaultProvider)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	log.Printf("gRPC port configured: %d (gRPC server not yet wired — Phase 2)", cfg.GRPCPort)

	// Block until SIGINT/SIGTERM.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down tax-provider-integration...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Forced shutdown: %v", err)
	}
	log.Println("Server exited cleanly")
}
