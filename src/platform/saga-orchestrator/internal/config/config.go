package config

import (
	"os"
	"strings"
	"time"
)

type Config struct {
	HTTPPort    string
	GRPCPort    string
	Env         string
	DatabaseURL string
	KafkaBrokers []string
	KafkaGroupID string
	// Topics consumed by the orchestrator
	TopicOrderPlaced     string
	TopicPaymentResult   string
	TopicInventoryResult string
	TopicShipmentResult  string
	// Topics produced by the orchestrator
	TopicReserveInventory string
	TopicProcessPayment   string
	TopicCreateShipment   string
	TopicOrderCancelled   string
	StepTimeout           time.Duration
}

func Load() *Config {
	return &Config{
		HTTPPort:              getEnv("HTTP_PORT", "8094"),
		GRPCPort:              getEnv("GRPC_PORT", "50054"),
		Env:                   getEnv("ENV", "development"),
		DatabaseURL:           getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/saga?sslmode=disable"),
		KafkaBrokers:          strings.Split(getEnv("KAFKA_BROKERS", "localhost:9092"), ","),
		KafkaGroupID:          getEnv("KAFKA_GROUP_ID", "saga-orchestrator"),
		TopicOrderPlaced:      getEnv("TOPIC_ORDER_PLACED", "commerce.order.placed"),
		TopicPaymentResult:    getEnv("TOPIC_PAYMENT_RESULT", "commerce.payment.processed"),
		TopicInventoryResult:  getEnv("TOPIC_INVENTORY_RESULT", "supplychain.inventory.reserved"),
		TopicShipmentResult:   getEnv("TOPIC_SHIPMENT_RESULT", "supplychain.shipment.created"),
		TopicReserveInventory: getEnv("TOPIC_RESERVE_INVENTORY", "supplychain.inventory.reserve"),
		TopicProcessPayment:   getEnv("TOPIC_PROCESS_PAYMENT", "commerce.payment.process"),
		TopicCreateShipment:   getEnv("TOPIC_CREATE_SHIPMENT", "supplychain.shipment.create"),
		TopicOrderCancelled:   getEnv("TOPIC_ORDER_CANCELLED", "commerce.order.cancelled"),
		StepTimeout:           getDuration("STEP_TIMEOUT", 30*time.Second),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}
