package config

import (
	"os"
)

// Config holds all runtime configuration for the service.
type Config struct {
	HTTPPort      string
	GRPCPort      string
	DatabaseURL   string
	DefaultLocale string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	return &Config{
		HTTPPort:      getEnv("HTTP_PORT", "8606"),
		GRPCPort:      getEnv("GRPC_PORT", "50146"),
		DatabaseURL:   getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/i18ndb?sslmode=disable"),
		DefaultLocale: getEnv("DEFAULT_LOCALE", "en"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
