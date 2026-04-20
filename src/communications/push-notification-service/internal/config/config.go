// Package config loads push-notification-service settings from environment variables.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds all runtime configuration for the service.
type Config struct {
	HTTPPort     string
	KafkaBrokers []string
	KafkaTopic   string
	KafkaGroupID string
	MaxStoreSize int
}

// Load reads configuration from environment variables and returns a Config.
// Sensible defaults are applied for every field that is not set.
func Load() (*Config, error) {
	port := getEnv("HTTP_PORT", "8504")
	brokersRaw := getEnv("KAFKA_BROKERS", "localhost:9092")
	topic := getEnv("KAFKA_TOPIC", "push.send")
	groupID := getEnv("KAFKA_GROUP_ID", "push-notification-service")
	maxStoreSizeStr := getEnv("MAX_STORE_SIZE", "10000")

	maxStoreSize, err := strconv.Atoi(maxStoreSizeStr)
	if err != nil {
		return nil, fmt.Errorf("invalid MAX_STORE_SIZE %q: %w", maxStoreSizeStr, err)
	}
	if maxStoreSize < 1 {
		return nil, fmt.Errorf("MAX_STORE_SIZE must be >= 1, got %d", maxStoreSize)
	}

	brokers := splitAndTrim(brokersRaw, ",")
	if len(brokers) == 0 {
		return nil, fmt.Errorf("KAFKA_BROKERS must not be empty")
	}

	return &Config{
		HTTPPort:     port,
		KafkaBrokers: brokers,
		KafkaTopic:   topic,
		KafkaGroupID: groupID,
		MaxStoreSize: maxStoreSize,
	}, nil
}

// ListenAddr returns the TCP listen address for the HTTP server.
func (c *Config) ListenAddr() string {
	return ":" + c.HTTPPort
}

// -------------------------------------------------------------------
// Helpers
// -------------------------------------------------------------------

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && strings.TrimSpace(v) != "" {
		return strings.TrimSpace(v)
	}
	return fallback
}

func splitAndTrim(s, sep string) []string {
	parts := strings.Split(s, sep)
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}
