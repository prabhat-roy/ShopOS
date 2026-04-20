package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the in-app-notification-service.
type Config struct {
	RedisURL    string
	HTTPPort    string
	GRPCPort    string
	NotifTTL    time.Duration
	MaxPerUser  int
}

// Load reads configuration from environment variables, falling back to defaults.
func Load() *Config {
	return &Config{
		RedisURL:   getEnv("REDIS_URL", "redis://localhost:6379"),
		HTTPPort:   getEnv("HTTP_PORT", "8505"),
		GRPCPort:   getEnv("GRPC_PORT", "50132"),
		NotifTTL:   getDurationEnv("NOTIF_TTL_DAYS", 30) * 24 * time.Hour,
		MaxPerUser: getIntEnv("MAX_PER_USER", 100),
	}
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func getIntEnv(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return defaultVal
}

func getDurationEnv(key string, defaultDays time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := strconv.ParseInt(v, 10, 64); err == nil {
			return time.Duration(d)
		}
	}
	return defaultDays
}
