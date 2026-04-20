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

	"github.com/shopos/budget-service/internal/config"
	"github.com/shopos/budget-service/internal/handler"
	"github.com/shopos/budget-service/internal/service"
	"github.com/shopos/budget-service/internal/store"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("budget-service: loading config: %v", err)
	}

	st, err := store.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("budget-service: connecting to database: %v", err)
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

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("budget-service: HTTP server listening on %s", httpAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("budget-service: HTTP server error: %v", err)
		}
	}()

	<-quit
	log.Println("budget-service: shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("budget-service: forced shutdown: %v", err)
	}
	log.Println("budget-service: stopped")
}
