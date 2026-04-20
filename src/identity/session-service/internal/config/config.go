package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all runtime configuration sourced from environment variables.
type Config struct {
	HTTPPort       string
	GRPCPort       string
	RedisAddr      string
	RedisPassword  string
	RedisDB        int
	SessionTTL     time.Duration
}

// Load reads configuration from environment variables, applying sensible defaults.
func Load() *Config {
	return &Config{
		HTTPPort:      getEnv("HTTP_PORT", "8100"),
		GRPCPort:      getEnv("GRPC_PORT", "50062"),
		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvInt("REDIS_DB", 0),
		SessionTTL:    getEnvDuration("SESSION_TTL", 24*time.Hour),
	}
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return defaultVal
}

func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return defaultVal
}
