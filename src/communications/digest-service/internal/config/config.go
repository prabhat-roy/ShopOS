package config

import (
	"os"
	"strings"
	"time"
)

// Config holds all runtime configuration for the digest-service.
type Config struct {
	HTTPPort            string
	DatabaseURL         string
	KafkaBrokers        []string
	KafkaTopic          string
	DigestCheckInterval time.Duration
}

// Load reads configuration from environment variables, applying defaults where absent.
func Load() *Config {
	return &Config{
		HTTPPort:            getEnv("HTTP_PORT", "8506"),
		DatabaseURL:         getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/digest?sslmode=disable"),
		KafkaBrokers:        splitCSV(getEnv("KAFKA_BROKERS", "localhost:9092")),
		KafkaTopic:          getEnv("KAFKA_TOPIC", "email.send"),
		DigestCheckInterval: getDurationEnv("DIGEST_CHECK_INTERVAL", time.Hour),
	}
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

func getDurationEnv(key string, defaultVal time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return defaultVal
}
