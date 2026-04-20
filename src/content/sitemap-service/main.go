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

	"github.com/shopos/sitemap-service/internal/config"
	"github.com/shopos/sitemap-service/internal/generator"
	"github.com/shopos/sitemap-service/internal/handler"
)

func main() {
	cfg := config.Load()

	log.Printf("sitemap-service starting (HTTP :%s, gRPC :%s)", cfg.HTTPPort, cfg.GRPCPort)
	log.Printf("base URL: %s, max URLs per sitemap: %d", cfg.BaseURL, cfg.MaxURLsPerSitemap)

	gen := generator.New(cfg.MaxURLsPerSitemap)
	h := handler.New(gen, cfg.BaseURL)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.HTTPPort),
		Handler:      h,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Printf("HTTP server listening on :%s", cfg.HTTPPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down gracefully…")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("HTTP server shutdown: %v", err)
	}
	log.Println("sitemap-service stopped")
}
