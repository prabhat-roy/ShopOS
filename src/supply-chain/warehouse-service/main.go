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

	"github.com/shopos/warehouse-service/internal/config"
	"github.com/shopos/warehouse-service/internal/handler"
	"github.com/shopos/warehouse-service/internal/service"
	"github.com/shopos/warehouse-service/internal/store"
)

func main() {
	logger := log.New(os.Stdout, "[warehouse-service] ", log.LstdFlags|log.Lmsgprefix)

	cfg, err := config.Load()
	if err != nil {
		logger.Fatalf("config: %v", err)
	}

	st, err := store.New(cfg.DatabaseURL)
	if err != nil {
		logger.Fatalf("store: %v", err)
	}
	defer func() {
		if err := st.Close(); err != nil {
			logger.Printf("store close: %v", err)
		}
	}()

	svc := service.New(st)
	h := handler.New(svc)

	addr := fmt.Sprintf(":%s", cfg.HTTPPort)
	srv := &http.Server{
		Addr:         addr,
		Handler:      h,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown channel.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.Printf("HTTP server listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("listen: %v", err)
		}
	}()

	<-quit
	logger.Println("shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatalf("server shutdown: %v", err)
	}
	logger.Println("server stopped")
}
