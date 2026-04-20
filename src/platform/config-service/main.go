package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/shopos/config-service/internal/config"
	grpcserver "github.com/shopos/config-service/internal/grpc"
	"github.com/shopos/config-service/internal/service"
	"github.com/shopos/config-service/internal/store"
	"go.uber.org/zap"
)

func main() {
	_ = godotenv.Load()

	cfg := config.Load()

	var logger *zap.Logger
	if cfg.Env == "production" {
		logger, _ = zap.NewProduction()
	} else {
		logger, _ = zap.NewDevelopment()
	}
	defer logger.Sync()

	st, err := store.New(cfg.EtcdAddrs, cfg.Prefix, cfg.EtcdTimeout)
	if err != nil {
		log.Fatalf("etcd init: %v", err)
	}
	defer st.Close()

	svc := service.New(st, logger)
	srv := grpcserver.NewServer(svc, logger)

	mux := http.NewServeMux()
	srv.RegisterHTTP(mux)

	httpSrv := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("config-service HTTP listening", zap.String("port", cfg.HTTPPort))
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("shutting down config-service...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpSrv.Shutdown(ctx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}
	logger.Info("config-service stopped")
}
