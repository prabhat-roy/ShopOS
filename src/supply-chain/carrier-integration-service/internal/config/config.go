package config

import (
	"os"
)

// Config holds all configuration values for the carrier-integration-service.
type Config struct {
	HTTPPort    string
	GRPCPort    string
	LogLevel    string
	Environment string
}

// Load reads configuration from environment variables, falling back to defaults.
func Load() *Config {
	return &Config{
		HTTPPort:    getEnv("HTTP_PORT", "8206"),
		GRPCPort:    getEnv("GRPC_PORT", "50106"),
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
func (c *Config) HTTPAddr() string {
	return ":" + c.HTTPPort
}

// GRPCAddr returns the full address string for the gRPC server.
func (c *Config) GRPCAddr() string {
	return ":" + c.GRPCPort
}

// IsDevelopment returns true when running in development mode.
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}
