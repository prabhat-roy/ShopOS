package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	HTTPPort string

	DatabaseURL string
	DBMaxConns  int
	DBTimeout   time.Duration

	RedisAddr string

	Brokers         []string
	GroupID         string
	WatchTopics     []string
	DeliveryTimeout time.Duration
}

func Load() *Config {
	maxConns, _ := strconv.Atoi(env("DB_MAX_CONNS", "10"))
	dbTimeout, _ := time.ParseDuration(env("DB_TIMEOUT", "5s"))
	deliveryTimeout, _ := time.ParseDuration(env("DELIVERY_TIMEOUT", "10s"))

	return &Config{
		HTTPPort:    env("HTTP_PORT", "8091"),
		DatabaseURL: env("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/webhooks?sslmode=disable"),
		DBMaxConns:  maxConns,
		DBTimeout:   dbTimeout,

		Brokers:         strings.Split(env("KAFKA_BROKERS", "localhost:9092"), ","),
		GroupID:         env("KAFKA_GROUP_ID", "webhook-service"),
		WatchTopics:     strings.Split(env("KAFKA_WATCH_TOPICS", "commerce.order.placed,commerce.order.cancelled,commerce.payment.processed"), ","),
		DeliveryTimeout: deliveryTimeout,
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
