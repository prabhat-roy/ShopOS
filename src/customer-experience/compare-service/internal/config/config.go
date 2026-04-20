package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all runtime configuration for the compare-service.
type Config struct {
	RedisURL   string
	HTTPPort   string
	GRPCPort   string
	CompareTTL time.Duration
	MaxItems   int
}

// Load reads configuration from environment variables and returns a populated Config.
// Mandatory variable: REDIS_URL.
func Load() (*Config, error) {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		return nil, fmt.Errorf("REDIS_URL environment variable is required")
	}

	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8403"
	}

	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50123"
	}

	compareTTL := 24 * time.Hour
	if v := os.Getenv("COMPARE_TTL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			compareTTL = d
		}
	}

	maxItems := 4
	if v := os.Getenv("MAX_ITEMS"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 {
			return nil, fmt.Errorf("invalid MAX_ITEMS %q: must be a positive integer", v)
		}
		maxItems = n
	}

	return &Config{
		RedisURL:   redisURL,
		HTTPPort:   httpPort,
		GRPCPort:   grpcPort,
		CompareTTL: compareTTL,
		MaxItems:   maxItems,
	}, nil
}
