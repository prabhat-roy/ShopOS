package config

import (
	"fmt"
	"os"
)

// Config holds all runtime configuration for the warehouse-service.
type Config struct {
	HTTPPort    string
	GRPCPort    string
	DatabaseURL string
}

// Load reads configuration from environment variables, applying defaults where
// appropriate and returning an error when required values are missing.
func Load() (*Config, error) {
	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8202"
	}

	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50102"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is required")
	}

	return &Config{
		HTTPPort:    httpPort,
		GRPCPort:    grpcPort,
		DatabaseURL: databaseURL,
	}, nil
}
