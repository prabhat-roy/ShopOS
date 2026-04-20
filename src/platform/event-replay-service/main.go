package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/shopos/event-replay-service/internal/config"
	"github.com/shopos/event-replay-service/internal/handler"
	"github.com/shopos/event-replay-service/internal/service"
	"github.com/shopos/event-replay-service/internal/store"
)

func main() {
	cfg := config.Load()

	// ---- database -----------------------------------------------------------
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	db.SetMaxOpenConns(cfg.DBMaxConns)
	db.SetMaxIdleConns(cfg.DBMaxConns / 2)
	db.SetConnMaxLifetime(5 * time.Minute)

	// ---- wire dependencies --------------------------------------------------
	replayStore := store.New(db)
	replaySvc := service.New(replayStore, cfg.EventStoreURL)
	h := handler.New(replaySvc)

	// ---- HTTP server --------------------------------------------------------
	srv := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      h,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine so we can listen for shutdown signals.
	go func() {
		log.Printf("event-replay-service listening on :%s (gRPC port reserved: %s)",
			cfg.HTTPPort, cfg.GRPCPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	// ---- graceful shutdown --------------------------------------------------
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutdown signal received")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server shutdown: %v", err)
	}

	if err := db.Close(); err != nil {
		log.Printf("db close: %v", err)
	}

	log.Println("event-replay-service stopped")
}
