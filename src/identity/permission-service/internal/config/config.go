package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all runtime configuration for the permission-service.
type Config struct {
	HTTPPort    string
	GRPCPort    string
	DatabaseURL string
	DBMaxConns  int
}

// Load reads configuration from environment variables, applying sensible defaults.
func Load() (*Config, error) {
	httpPort := getEnv("HTTP_PORT", "8101")
	grpcPort := getEnv("GRPC_PORT", "50063")

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is required")
	}

	dbMaxConns := 10
	if raw := os.Getenv("DB_MAX_CONNS"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil {
			return nil, fmt.Errorf("invalid DB_MAX_CONNS value %q: %w", raw, err)
		}
		dbMaxConns = v
	}

	return &Config{
		HTTPPort:    httpPort,
		GRPCPort:    grpcPort,
		DatabaseURL: databaseURL,
		DBMaxConns:  dbMaxConns,
	}, nil
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
