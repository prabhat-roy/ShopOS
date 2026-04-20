package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"

	"github.com/shopos/dead-letter-service/internal/config"
	"github.com/shopos/dead-letter-service/internal/handler"
	"github.com/shopos/dead-letter-service/internal/kafka"
	"github.com/shopos/dead-letter-service/internal/service"
	"github.com/shopos/dead-letter-service/internal/store"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	cfg := config.Load()

	// ── Database ──────────────────────────────────────────────────────────────
	db, err := openDB(cfg)
	if err != nil {
		logger.Error("connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	logger.Info("database connected")

	// ── Wire layers ───────────────────────────────────────────────────────────
	dlqStore := store.New(db)
	dlqService := service.New(dlqStore)
	dlqHandler := handler.New(dlqService, logger)

	mux := http.NewServeMux()
	dlqHandler.RegisterRoutes(mux)

	// ── HTTP server ───────────────────────────────────────────────────────────
	srv := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// ── Graceful shutdown ─────────────────────────────────────────────────────
	// signal.NotifyContext cancels ctx on SIGINT / SIGTERM.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// ── Kafka consumer ────────────────────────────────────────────────────────
	var consumer *kafka.Consumer
	if len(cfg.KafkaBrokers) > 0 && len(cfg.KafkaDLQTopics) > 0 {
		consumer = kafka.NewConsumer(cfg.KafkaBrokers, cfg.KafkaGroupID, cfg.KafkaDLQTopics, dlqService, logger)
		go func() {
			logger.Info("kafka consumer starting", "topics", cfg.KafkaDLQTopics)
			consumer.Start(ctx)
			logger.Info("kafka consumer stopped")
		}()
	} else {
		logger.Warn("kafka not configured — consumer disabled")
	}

	// Start HTTP server in a goroutine so we can listen for shutdown signals.
	go func() {
		logger.Info("HTTP server starting", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server error", "error", err)
			stop() // trigger shutdown
		}
	}()

	// Block until a shutdown signal is received.
	<-ctx.Done()
	logger.Info("shutdown signal received")

	// Give in-flight requests up to 10 seconds to finish.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("HTTP server shutdown", "error", err)
	}

	if consumer != nil {
		consumer.Close()
	}

	logger.Info("dead-letter-service stopped")
}

// openDB opens and configures the Postgres connection pool.
func openDB(cfg *config.Config) (*sql.DB, error) {
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is not set")
	}
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %w", err)
	}
	db.SetMaxOpenConns(cfg.DBMaxConns)
	db.SetMaxIdleConns(cfg.DBMaxConns / 2)
	db.SetConnMaxLifetime(5 * time.Minute)
	return db, nil
}
