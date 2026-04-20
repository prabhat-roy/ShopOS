package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all runtime configuration for the live-chat-service.
type Config struct {
	RedisURL    string
	HTTPPort    string
	SessionTTL  time.Duration
	MaxMessages int
}

// Load reads configuration from environment variables, falling back to defaults.
func Load() *Config {
	return &Config{
		RedisURL:    getEnv("REDIS_URL", "redis://localhost:6379"),
		HTTPPort:    getEnv("HTTP_PORT", "8406"),
		SessionTTL:  getDuration("SESSION_TTL_SECONDS", 2*time.Hour),
		MaxMessages: getInt("MAX_MESSAGES", 100),
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

func getDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if secs, err := strconv.Atoi(v); err == nil {
			return time.Duration(secs) * time.Second
		}
	}
	return fallback
}
