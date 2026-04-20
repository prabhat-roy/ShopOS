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

	kafkago "github.com/segmentio/kafka-go"
	"github.com/shopos/marketplace-connector-service/internal/adapter"
	"github.com/shopos/marketplace-connector-service/internal/config"
	"github.com/shopos/marketplace-connector-service/internal/handler"
	"github.com/shopos/marketplace-connector-service/internal/service"
	"github.com/shopos/marketplace-connector-service/internal/store"
)

func main() {
	cfg := config.Load()

	log.Printf("marketplace-connector-service starting on port %s", cfg.HTTPPort)
	log.Printf("kafka brokers: %v, topic: %s", cfg.KafkaBrokers, cfg.KafkaTopic)

	// Kafka writer — sync with error logging.
	writer := &kafkago.Writer{
		Addr:         kafkago.TCP(cfg.KafkaBrokers...),
		Topic:        cfg.KafkaTopic,
		Balancer:     &kafkago.LeastBytes{},
		RequiredAcks: kafkago.RequireOne,
		Async:        false,
		ErrorLogger: kafkago.LoggerFunc(func(msg string, args ...interface{}) {
			log.Printf("[kafka-error] "+msg, args...)
		}),
	}

	st := store.New()
	ad := adapter.New()
	svc := service.New(st, ad, writer, cfg.KafkaTopic)
	h := handler.New(svc, ad)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.HTTPPort),
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	log.Printf("marketplace-connector-service listening on :%s", cfg.HTTPPort)
	<-quit
	log.Println("shutting down gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("server shutdown error: %v", err)
	}

	if err := writer.Close(); err != nil {
		log.Printf("kafka writer close error: %v", err)
	}

	log.Println("marketplace-connector-service stopped")
}