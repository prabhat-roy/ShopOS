package config

import (
	"fmt"
	"os"
)

// Config holds all runtime configuration sourced from environment variables.
type Config struct {
	HTTPPort    string
	GRPCPort    string
	DatabaseURL string
}

// Load reads configuration from environment variables, returning an error when
// a required variable is missing.
func Load() (*Config, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8104"
	}

	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50066"
	}

	return &Config{
		HTTPPort:    httpPort,
		GRPCPort:    grpcPort,
		DatabaseURL: dbURL,
	}, nil
}
