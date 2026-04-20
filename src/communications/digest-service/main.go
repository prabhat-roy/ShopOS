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

	"github.com/segmentio/kafka-go"
	"github.com/shopos/digest-service/internal/config"
	"github.com/shopos/digest-service/internal/handler"
	"github.com/shopos/digest-service/internal/scheduler"
	"github.com/shopos/digest-service/internal/store"
)

func main() {
	cfg := config.Load()

	// Connect to PostgreSQL.
	pgStore, err := store.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("store.New: %v", err)
	}
	defer pgStore.Close()
	pingCtx, pingCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer pingCancel()
	if err := pgStore.Ping(pingCtx); err != nil {
		log.Printf("warning: cannot reach Postgres: %v (continuing anyway)", err)
	} else {
		log.Println("connected to PostgreSQL")
	}

	// Build Kafka writer (producer only).
	kafkaWriter := &kafka.Writer{
		Addr:                   kafka.TCP(cfg.KafkaBrokers...),
		Topic:                  cfg.KafkaTopic,
		Balancer:               &kafka.LeastBytes{},
		RequiredAcks:           kafka.RequireOne,
		Async:                  false,
		AllowAutoTopicCreation: true,
	}
	defer func() {
		if err := kafkaWriter.Close(); err != nil {
			log.Printf("kafka writer close: %v", err)
		}
	}()
	log.Printf("Kafka writer configured (brokers=%v, topic=%s)", cfg.KafkaBrokers, cfg.KafkaTopic)

	// Wire HTTP handler.
	mux := http.NewServeMux()
	h := handler.New(pgStore)
	h.RegisterRoutes(mux)

	httpAddr := fmt.Sprintf(":%s", cfg.HTTPPort)
	httpServer := &http.Server{
		Addr:         httpAddr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start HTTP server.
	go func() {
		log.Printf("HTTP server listening on %s", httpAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Start background digest scheduler.
	sched := scheduler.New(pgStore, kafkaWriter, cfg.KafkaTopic, cfg.DigestCheckInterval)
	schedCtx, schedCancel := context.WithCancel(context.Background())
	defer schedCancel()
	go sched.Start(schedCtx)

	// Graceful shutdown.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down digest-service...")

	schedCancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server forced shutdown: %v", err)
	}
	log.Println("digest-service stopped")
}
