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

	"github.com/shopos/crm-integration-service/internal/adapter"
	"github.com/shopos/crm-integration-service/internal/config"
	"github.com/shopos/crm-integration-service/internal/handler"
	"github.com/shopos/crm-integration-service/internal/service"
	"github.com/shopos/crm-integration-service/internal/store"
)

func main() {
	cfg := config.Load()

	log.Printf("crm-integration-service starting — HTTP :%s  gRPC :%s", cfg.HTTPPort, cfg.GRPCPort)

	st := store.New()
	ad := adapter.New()
	svc := service.New(st, ad)
	h := handler.New(svc)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.HTTPPort),
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	log.Printf("crm-integration-service listening on HTTP :%s", cfg.HTTPPort)
	<-quit
	log.Println("shutting down gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("server shutdown error: %v", err)
	}

	log.Println("crm-integration-service stopped")
}