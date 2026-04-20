package config

import (
	"os"
	"time"
)

// Config holds all runtime configuration for the checkout-service.
type Config struct {
	HTTPPort    string
	GRPCPort    string
	CartURL     string
	TaxURL      string
	OrderURL    string
	HTTPTimeout time.Duration
}

// Load reads configuration from environment variables and returns a Config.
// Sensible defaults are applied for every field.
func Load() *Config {
	return &Config{
		HTTPPort:    getEnv("HTTP_PORT", "8132"),
		GRPCPort:    getEnv("GRPC_PORT", "50081"),
		CartURL:     getEnv("CART_URL", "http://cart-service:50080"),
		TaxURL:      getEnv("TAX_URL", "http://tax-service:8133"),
		OrderURL:    getEnv("ORDER_URL", "http://order-service:50082"),
		HTTPTimeout: getDuration("HTTP_TIMEOUT", 10*time.Second),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getDuration(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}
