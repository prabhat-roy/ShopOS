package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all runtime configuration for the service.
type Config struct {
	HTTPPort           string
	GRPCPort           string
	Env                string
	RedisAddr          string
	RedisPassword      string
	RedisDB            int
	WorkerConcurrency  int
	WorkerQueues       []string
	CallbackTimeout    time.Duration
}

// Load reads configuration from environment variables, applying defaults where
// values are absent. An error is returned only when a present value cannot be
// parsed into the required type.
func Load() (*Config, error) {
	redisDB, err := parseInt(getEnv("REDIS_DB", "1"))
	if err != nil {
		return nil, fmt.Errorf("REDIS_DB: %w", err)
	}

	concurrency, err := parseInt(getEnv("WORKER_CONCURRENCY", "5"))
	if err != nil {
		return nil, fmt.Errorf("WORKER_CONCURRENCY: %w", err)
	}

	callbackTimeout, err := time.ParseDuration(getEnv("CALLBACK_TIMEOUT", "30s"))
	if err != nil {
		return nil, fmt.Errorf("CALLBACK_TIMEOUT: %w", err)
	}

	rawQueues := getEnv("WORKER_QUEUES", "default")
	queues := splitAndTrim(rawQueues, ",")
	if len(queues) == 0 {
		queues = []string{"default"}
	}

	return &Config{
		HTTPPort:          getEnv("HTTP_PORT", "8096"),
		GRPCPort:          getEnv("GRPC_PORT", "50057"),
		Env:               getEnv("ENV", "development"),
		RedisAddr:         getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:     getEnv("REDIS_PASSWORD", ""),
		RedisDB:           redisDB,
		WorkerConcurrency: concurrency,
		WorkerQueues:      queues,
		CallbackTimeout:   callbackTimeout,
	}, nil
}

// getEnv returns the value of the named environment variable, or fallback when
// the variable is unset or empty.
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func parseInt(s string) (int, error) {
	return strconv.Atoi(strings.TrimSpace(s))
}

func splitAndTrim(s, sep string) []string {
	parts := strings.Split(s, sep)
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
