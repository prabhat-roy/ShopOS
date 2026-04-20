package config

import "os"

// Config holds all configuration for the returns-logistics-service.
type Config struct {
	HTTPPort    string
	GRPCPort    string
	DatabaseURL string
	LogLevel    string
	Environment string
}

// Load reads configuration from environment variables with defaults.
func Load() *Config {
	return &Config{
		HTTPPort:    getEnv("HTTP_PORT", "8208"),
		GRPCPort:    getEnv("GRPC_PORT", "50108"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/returns_db?sslmode=disable"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		Environment: getEnv("ENVIRONMENT", "development"),
	}
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

// HTTPAddr returns the full address string for the HTTP server.
func (c *Config) HTTPAddr() string { return ":" + c.HTTPPort }

// GRPCAddr returns the full address string for the gRPC server.
func (c *Config) GRPCAddr() string { return ":" + c.GRPCPort }
