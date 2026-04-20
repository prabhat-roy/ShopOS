package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPPort    string
	GRPCPort    string
	Env         string
	RedisAddr   string
	RedisPass   string
	RedisDB     int
	KeyTTL      time.Duration // how long a rate-limit window lives in Redis
	CleanupSecs int           // how often the in-process fallback map is GC'd
}

func Load() *Config {
	return &Config{
		HTTPPort:    getEnv("HTTP_PORT", "8093"),
		GRPCPort:    getEnv("GRPC_PORT", "50053"),
		Env:         getEnv("ENV", "development"),
		RedisAddr:   getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPass:   getEnv("REDIS_PASSWORD", ""),
		RedisDB:     getInt("REDIS_DB", 0),
		KeyTTL:      getDuration("RATE_LIMIT_TTL", time.Minute),
		CleanupSecs: getInt("CLEANUP_INTERVAL_SECS", 300),
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
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}
