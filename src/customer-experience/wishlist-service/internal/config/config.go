package config

import (
	"fmt"
	"os"
)

// Config holds all runtime configuration for the wishlist-service.
type Config struct {
	HTTPPort    string
	GRPCPort    string
	DatabaseURL string
}

// Load reads configuration from environment variables and returns a populated Config.
// Mandatory variable: DATABASE_URL.
func Load() (*Config, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is required")
	}

	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8402"
	}

	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50122"
	}

	return &Config{
		HTTPPort:    httpPort,
		GRPCPort:    grpcPort,
		DatabaseURL: dbURL,
	}, nil
}
