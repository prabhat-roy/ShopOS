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

	"github.com/shopos/expense-management-service/internal/config"
	"github.com/shopos/expense-management-service/internal/handler"
	"github.com/shopos/expense-management-service/internal/service"
	"github.com/shopos/expense-management-service/internal/store"
)

func main() {
	logger := log.New(os.Stdout, "[expense-management] ", log.LstdFlags|log.Lmsgprefix)

	cfg := config.Load()

	// Connect to PostgreSQL.
	pgStore, err := store.New(cfg.Database.DSN())
	if err != nil {
		logger.Fatalf("failed to connect to database: %v", err)
	}
	defer pgStore.Close()
	logger.Printf("connected to PostgreSQL at %s:%s", cfg.Database.Host, cfg.Database.Port)

	// Build service and handler layers.
	svc := service.New(pgStore)
	h := handler.New(svc)

	// HTTP server.
	httpSrv := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      h,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Printf("HTTP server listening on %s", cfg.HTTPAddr)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("HTTP server error: %v", err)
		}
	}()

	logger.Printf("gRPC address configured at %s (implementation ready for Phase 2)", cfg.GRPCAddr)

	// Graceful shutdown on SIGINT / SIGTERM.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	logger.Printf("received signal %v — shutting down", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpSrv.Shutdown(ctx); err != nil {
		logger.Printf("HTTP shutdown error: %v", err)
	}

	fmt.Println("expense-management-service stopped")
}
