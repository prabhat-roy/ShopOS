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

	"github.com/shopos/live-chat-service/internal/config"
	"github.com/shopos/live-chat-service/internal/handler"
	"github.com/shopos/live-chat-service/internal/service"
	"github.com/shopos/live-chat-service/internal/store"
)

func main() {
	cfg := config.Load()

	redisStore, err := store.New(cfg.RedisURL, cfg.SessionTTL, cfg.MaxMessages)
	if err != nil {
		log.Fatalf("failed to connect to Redis: %v", err)
	}

	svc := service.New(redisStore, cfg.MaxMessages)
	h := handler.New(svc)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.HTTPPort),
		Handler:      h,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("live-chat-service listening on :%s", cfg.HTTPPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}
	log.Println("server exited")
}
