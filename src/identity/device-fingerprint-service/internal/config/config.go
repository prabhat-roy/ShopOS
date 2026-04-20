package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all runtime configuration for the service.
type Config struct {
	HTTPPort    string
	GRPCPort    string
	RedisAddr   string
	RedisPass   string
	RedisDB     int
	FPTTLDays   int
}

// Load reads configuration from environment variables and applies sensible
// defaults for local development.
func Load() (*Config, error) {
	redisDB, err := strconv.Atoi(getEnv("REDIS_DB", "2"))
	if err != nil {
		return nil, fmt.Errorf("invalid REDIS_DB value: %w", err)
	}

	fpTTL, err := strconv.Atoi(getEnv("FP_TTL_DAYS", "90"))
	if err != nil {
		return nil, fmt.Errorf("invalid FP_TTL_DAYS value: %w", err)
	}

	return &Config{
		HTTPPort:  getEnv("HTTP_PORT", "8105"),
		GRPCPort:  getEnv("GRPC_PORT", "50067"),
		RedisAddr: getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPass: getEnv("REDIS_PASSWORD", ""),
		RedisDB:   redisDB,
		FPTTLDays: fpTTL,
	}, nil
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
