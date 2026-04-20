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

	"github.com/shopos/media-asset-service/internal/config"
	"github.com/shopos/media-asset-service/internal/handler"
	"github.com/shopos/media-asset-service/internal/service"
	"github.com/shopos/media-asset-service/internal/store"
)

func main() {
	cfg := config.Load()

	log.Printf("media-asset-service starting (HTTP :%s, gRPC :%s)", cfg.HTTPPort, cfg.GRPCPort)

	// Initialise MinIO store — this also ensures the bucket exists.
	assetStore, err := store.NewMinIOStore(
		cfg.MinioEndpoint,
		cfg.MinioAccess,
		cfg.MinioSecret,
		cfg.MinioBucket,
		cfg.MinioUseSSL,
	)
	if err != nil {
		log.Fatalf("init minio store: %v", err)
	}
	log.Printf("connected to MinIO at %s (bucket: %s)", cfg.MinioEndpoint, cfg.MinioBucket)

	svc := service.New(assetStore, cfg.MinioBucket, cfg.PresignExpiry)
	h := handler.New(svc)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.HTTPPort),
		Handler:      h,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start HTTP server in a goroutine.
	go func() {
		log.Printf("HTTP server listening on :%s", cfg.HTTPPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server: %v", err)
		}
	}()

	// Wait for termination signal.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down gracefully…")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("HTTP server shutdown: %v", err)
	}
	log.Println("media-asset-service stopped")
}
