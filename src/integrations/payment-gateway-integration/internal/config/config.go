package config

import "os"

// Config holds all runtime configuration for the payment-gateway-integration service.
type Config struct {
	HTTPPort string
	GRPCPort string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	return &Config{
		HTTPPort: getEnv("HTTP_PORT", "8904"),
		GRPCPort: getEnv("GRPC_PORT", "50173"),
	}
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
