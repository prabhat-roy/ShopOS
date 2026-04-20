package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/shopos/customs-duties-service/internal/calculator"
	"github.com/shopos/customs-duties-service/internal/config"
	"github.com/shopos/customs-duties-service/internal/handler"
)

func main() {
	cfg := config.Load()

	log.Printf("customs-duties-service starting (env=%s, http=%s, grpc=%s)",
		cfg.Environment, cfg.HTTPPort, cfg.GRPCPort)

	calc := calculator.New()
	h := handler.New(calc)

	srv := &http.Server{
		Addr:         cfg.HTTPAddr(),
		Handler:      h,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("HTTP server listening on %s", cfg.HTTPAddr())
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutdown signal received — draining connections…")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("graceful shutdown failed: %v", err)
	}
	log.Println("customs-duties-service stopped")
}
