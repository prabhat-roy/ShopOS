package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/shopos/checkout-service/internal/client"
	"github.com/shopos/checkout-service/internal/config"
	"github.com/shopos/checkout-service/internal/handler"
	"github.com/shopos/checkout-service/internal/service"
)

func main() {
	logger := log.New(os.Stdout, "[checkout-service] ", log.LstdFlags|log.Lmsgprefix)

	cfg := config.Load()

	httpClient := client.New(cfg.HTTPTimeout)
	svc := service.New(httpClient, cfg.CartURL, cfg.TaxURL, cfg.OrderURL)
	h := handler.New(svc, logger)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.HTTPPort),
		Handler:      h,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start HTTP server in a goroutine so we can handle shutdown signals.
	go func() {
		logger.Printf("HTTP server listening on :%s", cfg.HTTPPort)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Block until SIGINT or SIGTERM.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Println("Shutting down gracefully …")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatalf("Server forced to shutdown: %v", err)
	}
	logger.Println("Server stopped.")
}
