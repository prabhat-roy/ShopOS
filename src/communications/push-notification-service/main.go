// push-notification-service — entry point.
//
// Starts a Kafka consumer goroutine that processes push.send events and an
// HTTP server that exposes health, record lookup, list, and stats endpoints.
// Graceful shutdown is triggered by SIGINT or SIGTERM.
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

	"github.com/shopos/push-notification-service/internal/config"
	"github.com/shopos/push-notification-service/internal/consumer"
	"github.com/shopos/push-notification-service/internal/handler"
	"github.com/shopos/push-notification-service/internal/sender"
	"github.com/shopos/push-notification-service/internal/store"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("push-notification-service starting up")

	// --- Config ---
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// --- Core components ---
	pushStore := store.New(cfg.MaxStoreSize)
	pushSender := sender.New()

	// --- Kafka consumer ---
	kafkaConsumer := consumer.New(cfg, pushSender, pushStore)

	// Context that is cancelled on OS signal
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the consumer in a background goroutine
	go func() {
		kafkaConsumer.Run(ctx)
		log.Println("Kafka consumer goroutine exited")
	}()

	// --- HTTP server ---
	httpHandler := handler.New(pushStore, kafkaConsumer)
	srv := &http.Server{
		Addr:         cfg.ListenAddr(),
		Handler:      httpHandler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Serve in a background goroutine
	serverErr := make(chan error, 1)
	go func() {
		log.Printf("HTTP server listening on %s", cfg.ListenAddr())
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- fmt.Errorf("HTTP server error: %w", err)
		}
	}()

	// --- Graceful shutdown ---
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		log.Printf("Received signal %s — initiating graceful shutdown", sig)
	case err := <-serverErr:
		log.Printf("Server error: %v — initiating graceful shutdown", err)
	}

	// Cancel the consumer context so the Kafka loop exits
	cancel()

	// Give the HTTP server up to 15 s to drain in-flight requests
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server forced to shut down: %v", err)
	} else {
		log.Println("HTTP server shut down cleanly")
	}

	log.Println("push-notification-service stopped")
	os.Exit(0)
}
