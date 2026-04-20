package config

import (
	"fmt"
	"os"
)

// Config holds all runtime configuration for the mfa-service.
type Config struct {
	HTTPPort    string
	GRPCPort    string
	DatabaseURL string
	ServiceName string
}

// Load reads configuration from environment variables and returns a populated
// Config. Missing required fields cause an error to be returned.
func Load() (*Config, error) {
	cfg := &Config{
		HTTPPort:    getEnv("HTTP_PORT", "8102"),
		GRPCPort:    getEnv("GRPC_PORT", "50064"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		ServiceName: getEnv("SERVICE_NAME", "mfa-service"),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is required")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
