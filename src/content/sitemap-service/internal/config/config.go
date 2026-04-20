package config

import (
	"os"
	"strconv"
)

// Config holds all runtime configuration for the sitemap-service.
type Config struct {
	HTTPPort           string
	GRPCPort           string
	BaseURL            string
	MaxURLsPerSitemap  int
}

// Load reads configuration from environment variables and applies defaults.
func Load() *Config {
	return &Config{
		HTTPPort:          getEnv("HTTP_PORT", "8605"),
		GRPCPort:          getEnv("GRPC_PORT", "50145"),
		BaseURL:           getEnv("BASE_URL", "https://example.com"),
		MaxURLsPerSitemap: getIntEnv("MAX_URLS_PER_SITEMAP", 50000),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getIntEnv(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}
