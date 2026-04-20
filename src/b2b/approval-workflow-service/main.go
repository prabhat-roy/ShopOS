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

	"github.com/shopos/approval-workflow-service/internal/config"
	"github.com/shopos/approval-workflow-service/internal/handler"
	"github.com/shopos/approval-workflow-service/internal/service"
	"github.com/shopos/approval-workflow-service/internal/store"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "error", err)
		os.Exit(1)
	}

	st, err := store.New(cfg.DSN)
	if err != nil {
		slog.Error("connect to postgres", "error", err)
		os.Exit(1)
	}
	defer st.Close()
	slog.Info("connected to postgres")

	svc := service.New(st)
	h := handler.New(svc)

	httpServer := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      h,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// gRPC listener (reserved for future proto implementation)
	grpcLn, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		slog.Error("bind gRPC port", "addr", cfg.GRPCAddr, "error", err)
		os.Exit(1)
	}
	defer grpcLn.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		slog.Info("gRPC listener ready", "addr", cfg.GRPCAddr)
		<-ctx.Done()
	}()

	go func() {
		slog.Info("HTTP server starting", "addr", cfg.HTTPAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server error", "error", err)
			stop()
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("HTTP graceful shutdown", "error", err)
	}
	fmt.Println("approval-workflow-service stopped")
}
