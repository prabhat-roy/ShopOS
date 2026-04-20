package config

import (
	"os"
	"strconv"
)

// Config holds all runtime configuration for the age-verification-service.
type Config struct {
	HTTPPort     string
	GRPCPort     string
	DefaultMinAge int
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	return &Config{
		HTTPPort:      getEnv("HTTP_PORT", "8407"),
		GRPCPort:      getEnv("GRPC_PORT", "50128"),
		DefaultMinAge: getInt("DEFAULT_MIN_AGE", 18),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
