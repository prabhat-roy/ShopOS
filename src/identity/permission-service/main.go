package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"

	"github.com/shopos/permission-service/internal/config"
	"github.com/shopos/permission-service/internal/handler"
	"github.com/shopos/permission-service/internal/service"
	"github.com/shopos/permission-service/internal/store"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// ---- database ------------------------------------------------------------
	pgStore, err := store.NewPostgresStore(cfg.DatabaseURL, cfg.DBMaxConns)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pgStore.Close()
	slog.Info("database connection established")

	// ---- service + handler ---------------------------------------------------
	svc := service.New(pgStore)
	h := handler.New(svc)

	// ---- HTTP server ---------------------------------------------------------
	httpAddr := fmt.Sprintf(":%s", cfg.HTTPPort)
	httpServer := &http.Server{
		Addr:         httpAddr,
		Handler:      h,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// ---- gRPC placeholder listener -------------------------------------------
	// Full gRPC implementation is added in a subsequent phase once proto files
	// are generated.  We bind the port now so Kubernetes readiness probes can
	// validate the declared port contract.
	grpcAddr := fmt.Sprintf(":%s", cfg.GRPCPort)
	grpcListener, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		slog.Error("failed to bind gRPC port", "addr", grpcAddr, "error", err)
		os.Exit(1)
	}
	defer grpcListener.Close()

	// ---- graceful shutdown ---------------------------------------------------
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("HTTP server starting", "addr", httpAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server error", "error", err)
			os.Exit(1)
		}
	}()

	slog.Info("gRPC port bound (placeholder)", "addr", grpcAddr)

	<-quit
	slog.Info("shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		slog.Error("HTTP server shutdown error", "error", err)
	}
	slog.Info("permission-service stopped")
}
