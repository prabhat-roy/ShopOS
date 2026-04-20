package config

import (
	"os"
	"strings"
)

// Config holds all runtime configuration for the marketplace-connector-service.
type Config struct {
	HTTPPort      string
	KafkaBrokers  []string
	KafkaTopic    string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	brokerStr := getEnv("KAFKA_BROKERS", "localhost:9092")
	brokers := strings.Split(brokerStr, ",")
	for i, b := range brokers {
		brokers[i] = strings.TrimSpace(b)
	}

	return &Config{
		HTTPPort:     getEnv("HTTP_PORT", "8901"),
		KafkaBrokers: brokers,
		KafkaTopic:   getEnv("KAFKA_TOPIC", "marketplace.sync.completed"),
	}
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
