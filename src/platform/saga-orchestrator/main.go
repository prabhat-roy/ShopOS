package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/shopos/saga-orchestrator/internal/config"
	"github.com/shopos/saga-orchestrator/internal/handler"
	kafkaclient "github.com/shopos/saga-orchestrator/internal/kafka"
	"github.com/shopos/saga-orchestrator/internal/service"
	"github.com/shopos/saga-orchestrator/internal/store"
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

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db open: %v", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	st := store.New(db)
	producer := kafkaclient.NewProducer(cfg.KafkaBrokers, logger)
	defer producer.Close()

	topics := service.Topics{
		ReserveInventory: cfg.TopicReserveInventory,
		ProcessPayment:   cfg.TopicProcessPayment,
		CreateShipment:   cfg.TopicCreateShipment,
		OrderCancelled:   cfg.TopicOrderCancelled,
	}

	orch := service.New(st, producer, topics, logger)

	// Kafka consumer wires incoming result events to orchestrator handlers
	consumerTopics := []string{
		cfg.TopicOrderPlaced,
		cfg.TopicPaymentResult,
		cfg.TopicInventoryResult,
		cfg.TopicShipmentResult,
	}
	consumer := kafkaclient.NewConsumer(cfg.KafkaBrokers, cfg.KafkaGroupID, consumerTopics, logger)

	consumer.Register(cfg.TopicOrderPlaced, func(ctx context.Context, _, key string, payload []byte) error {
		var msg struct {
			OrderID string            `json:"order_id"`
			Payload map[string]string `json:"payload"`
		}
		kafkaclient.ParseJSON(payload, &msg)
		_, err := orch.Start(ctx, service.StartRequest(msg.OrderID, msg.Payload))
		return err
	})

	consumer.Register(cfg.TopicInventoryResult, func(ctx context.Context, _, _ string, payload []byte) error {
		var msg struct {
			SagaID  string `json:"saga_id"`
			Success bool   `json:"success"`
			Error   string `json:"error"`
		}
		kafkaclient.ParseJSON(payload, &msg)
		return orch.OnInventoryResult(ctx, msg.SagaID, msg.Success, msg.Error)
	})

	consumer.Register(cfg.TopicPaymentResult, func(ctx context.Context, _, _ string, payload []byte) error {
		var msg struct {
			SagaID  string `json:"saga_id"`
			Success bool   `json:"success"`
			Error   string `json:"error"`
		}
		kafkaclient.ParseJSON(payload, &msg)
		return orch.OnPaymentResult(ctx, msg.SagaID, msg.Success, msg.Error)
	})

	consumer.Register(cfg.TopicShipmentResult, func(ctx context.Context, _, _ string, payload []byte) error {
		var msg struct {
			SagaID  string `json:"saga_id"`
			Success bool   `json:"success"`
			Error   string `json:"error"`
		}
		kafkaclient.ParseJSON(payload, &msg)
		return orch.OnShipmentResult(ctx, msg.SagaID, msg.Success, msg.Error)
	})

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	consumer.Start(ctx)

	h := handler.New(orch)
	mux := http.NewServeMux()
	h.Register(mux)

	srv := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("saga-orchestrator listening", zap.String("port", cfg.HTTPPort))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down saga-orchestrator...")

	shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutCtx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}
	consumer.Close()
	logger.Info("saga-orchestrator stopped")
}
