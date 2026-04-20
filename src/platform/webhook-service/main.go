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
	"github.com/shopos/webhook-service/internal/config"
	"github.com/shopos/webhook-service/internal/delivery"
	"github.com/shopos/webhook-service/internal/handler"
	"github.com/shopos/webhook-service/internal/kafka"
	"github.com/shopos/webhook-service/internal/service"
	"github.com/shopos/webhook-service/internal/store"
)

func main() {
	cfg := config.Load()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		slog.Error("db open failed", "err", err)
		os.Exit(1)
	}
	db.SetMaxOpenConns(cfg.DBMaxConns)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	st := store.New(db)
	dispatcher := delivery.NewDispatcher(cfg.DeliveryTimeout)
	svc := service.New(st, dispatcher)

	consumer := kafka.NewConsumer(cfg.Brokers, cfg.GroupID, cfg.WatchTopics, svc)
	go func() {
		if err := consumer.Run(ctx); err != nil {
			slog.Error("kafka consumer error", "err", err)
		}
	}()

	mux := http.NewServeMux()
	handler.New(svc).Register(mux)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.HTTPPort),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		srv.Shutdown(shutCtx)
	}()

	slog.Info("webhook-service listening", "port", cfg.HTTPPort)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("server error", "err", err)
		os.Exit(1)
	}
}