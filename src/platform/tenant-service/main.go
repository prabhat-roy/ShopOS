package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/shopos/tenant-service/internal/config"
	"github.com/shopos/tenant-service/internal/handler"
	"github.com/shopos/tenant-service/internal/service"
	"github.com/shopos/tenant-service/internal/store"
)

func main() {
	cfg := config.Load()

	// — database —
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("opening database: %v", err)
	}
	db.SetMaxOpenConns(cfg.DBMaxConns)
	db.SetConnMaxLifetime(cfg.DBTimeout)

	_, cancelPing := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelPing()
	// — wire dependencies —
	tenantStore := store.New(db)
	tenantService := service.New(tenantStore)
	tenantHandler := handler.New(tenantService)

	// — HTTP server —
	addr := fmt.Sprintf(":%s", cfg.HTTPPort)
	srv := &http.Server{
		Addr:         addr,
		Handler:      tenantHandler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown: listen for SIGINT / SIGTERM.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("tenant-service listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Block until a signal is received.
	<-ctx.Done()
	log.Println("shutting down gracefully…")

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelShutdown()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown error: %v", err)
	}

	if err := db.Close(); err != nil {
		log.Printf("closing database: %v", err)
	}

	log.Println("tenant-service stopped")
}
