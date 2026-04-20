package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/shopos/gdpr-service/internal/config"
	"github.com/shopos/gdpr-service/internal/handler"
	"github.com/shopos/gdpr-service/internal/service"
	"github.com/shopos/gdpr-service/internal/store"
)

func main() {
	logger := log.New(os.Stdout, "[gdpr-service] ", log.LstdFlags|log.Lmsgprefix)

	cfg, err := config.Load()
	if err != nil {
		logger.Fatalf("config: %v", err)
	}

	// Persistence layer
	st, err := store.New(cfg.DatabaseURL)
	if err != nil {
		logger.Fatalf("store: %v", err)
	}

	// Business logic layer
	svc := service.New(st)

	// HTTP handler
	h := handler.New(svc)

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.HTTPPort),
		Handler:      h,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// gRPC placeholder listener — reserved for future gRPC implementation
	grpcAddr := fmt.Sprintf(":%s", cfg.GRPCPort)
	grpcListener, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		logger.Fatalf("grpc listener: %v", err)
	}
	defer grpcListener.Close()

	// Start HTTP server
	go func() {
		logger.Printf("HTTP server listening on %s", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("HTTP server error: %v", err)
		}
	}()

	logger.Printf("gRPC port reserved on %s", grpcAddr)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Println("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Fatalf("HTTP server shutdown: %v", err)
	}
	logger.Println("shutdown complete")
}
