package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all runtime configuration for the graphql-gateway service.
type Config struct {
	HTTPPort   string
	CatalogURL string
	CartURL    string
	OrdersURL  string
	UserURL    string
	Timeout    time.Duration
}

// Load reads configuration from environment variables and returns a populated Config.
// Defaults are applied when a variable is not set.
func Load() *Config {
	return &Config{
		HTTPPort:   envOrDefault("HTTP_PORT", "8086"),
		CatalogURL: envOrDefault("CATALOG_URL", "http://product-catalog-service:50070"),
		CartURL:    envOrDefault("CART_URL", "http://cart-service:50080"),
		OrdersURL:  envOrDefault("ORDERS_URL", "http://order-service:50082"),
		UserURL:    envOrDefault("USER_URL", "http://user-service:50061"),
		Timeout:    parseDuration(envOrDefault("DOWNSTREAM_TIMEOUT_MS", "5000")),
	}
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// parseDuration converts a millisecond string (e.g. "5000") to a time.Duration.
// Falls back to 5 s on parse failure.
func parseDuration(ms string) time.Duration {
	n, err := strconv.ParseInt(ms, 10, 64)
	if err != nil || n <= 0 {
		return 5 * time.Second
	}
	return time.Duration(n) * time.Millisecond
}
