package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/enterprise/graphql-gateway/internal/client"
	"github.com/enterprise/graphql-gateway/internal/config"
	"github.com/enterprise/graphql-gateway/internal/handler"
	"github.com/enterprise/graphql-gateway/internal/resolver"
)

func main() {
	cfg := config.Load()

	httpClient := client.New(cfg.Timeout)
	res := resolver.New(cfg, httpClient)
	h := handler.New(res)

	mux := http.NewServeMux()
	h.Register(mux)

	srv := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine so we can listen for OS signals.
	go func() {
		log.Printf("graphql-gateway listening on :%s", cfg.HTTPPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Wait for interrupt or terminate signal.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down graphql-gateway...")

	// Graceful shutdown with a 10-second deadline.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}
	log.Println("graphql-gateway stopped")
}
