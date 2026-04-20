package config

import (
	"fmt"
	"os"
)

// Config holds all runtime configuration for the gdpr-service.
type Config struct {
	HTTPPort    string
	GRPCPort    string
	DatabaseURL string
}

// Load reads configuration from environment variables.
// All values fall back to safe defaults for local development.
func Load() (*Config, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is required")
	}

	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8103"
	}

	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50065"
	}

	return &Config{
		HTTPPort:    httpPort,
		GRPCPort:    grpcPort,
		DatabaseURL: dbURL,
	}, nil
}
