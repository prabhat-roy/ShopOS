package config

import (
	"os"
	"strconv"
)

// Config holds all runtime configuration for the logistics-provider-integration service.
type Config struct {
	HTTPPort int
	GRPCPort int
	LogLevel string
}

// Load reads configuration from environment variables, applying defaults where needed.
func Load() *Config {
	return &Config{
		HTTPPort: envInt("HTTP_PORT", 8905),
		GRPCPort: envInt("GRPC_PORT", 50174),
		LogLevel: envStr("LOG_LEVEL", "info"),
	}
}

func envInt(key string, defaultVal int) int {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return defaultVal
	}
	return n
}

func envStr(key, defaultVal string) string {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	return v
}
