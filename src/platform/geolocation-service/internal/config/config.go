package config

import "os"

// Config holds all runtime configuration loaded from environment variables.
type Config struct {
	HTTPPort string
	GRPCPort string
}

// Load reads configuration from environment variables, applying sensible defaults.
func Load() *Config {
	return &Config{
		HTTPPort: getEnv("HTTP_PORT", "8093"),
		GRPCPort: getEnv("GRPC_PORT", "50058"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
