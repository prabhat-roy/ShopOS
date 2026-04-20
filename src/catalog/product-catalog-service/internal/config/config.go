package config

import (
	"fmt"
	"os"
)

// Config holds all runtime configuration for the service.
type Config struct {
	MongoURI string
	DBName   string
	HTTPPort string
	GRPCPort string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() (*Config, error) {
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		return nil, fmt.Errorf("MONGODB_URI environment variable is required")
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "catalog"
	}

	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8110"
	}

	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50070"
	}

	return &Config{
		MongoURI: mongoURI,
		DBName:   dbName,
		HTTPPort: httpPort,
		GRPCPort: grpcPort,
	}, nil
}
