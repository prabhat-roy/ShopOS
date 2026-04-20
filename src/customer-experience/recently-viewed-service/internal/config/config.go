package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all runtime configuration for the recently-viewed-service.
type Config struct {
	RedisURL string
	HTTPPort string
	GRPCPort string
	MaxItems int
	ItemTTL  time.Duration
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
		httpPort = "8404"
	}

	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50124"
	}

	maxItems := 50
	if v := os.Getenv("MAX_ITEMS"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 {
			return nil, fmt.Errorf("invalid MAX_ITEMS %q: must be a positive integer", v)
		}
		maxItems = n
	}

	itemTTL := 7 * 24 * time.Hour // 7 days
	if v := os.Getenv("ITEM_TTL"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return nil, fmt.Errorf("invalid ITEM_TTL %q: %w", v, err)
		}
		itemTTL = d
	}

	return &Config{
		RedisURL: redisURL,
		HTTPPort: httpPort,
		GRPCPort: grpcPort,
		MaxItems: maxItems,
		ItemTTL:  itemTTL,
	}, nil
}
