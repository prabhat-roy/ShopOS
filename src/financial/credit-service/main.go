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

	"github.com/shopos/credit-service/internal/config"
	"github.com/shopos/credit-service/internal/handler"
	"github.com/shopos/credit-service/internal/service"
	"github.com/shopos/credit-service/internal/store"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("credit-service: loading config: %v", err)
	}

	st, err := store.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("credit-service: connecting to database: %v", err)
	}
	defer st.Close()

	svc := service.New(st)
	h := handler.New(svc)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	httpAddr := fmt.Sprintf(":%s", cfg.HTTPPort)
	server := &http.Server{
		Addr:         httpAddr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown on SIGINT / SIGTERM.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("credit-service: HTTP server listening on %s", httpAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("credit-service: HTTP server error: %v", err)
		}
	}()

	<-quit
	log.Println("credit-service: shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("credit-service: forced shutdown: %v", err)
	}
	log.Println("credit-service: stopped")
}
