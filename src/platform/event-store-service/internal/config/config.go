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
	DatabaseURL string
	DBMaxConns  int
	DBTimeout   time.Duration
	PageSizeMax int // hard cap on events returned per query
}

func Load() *Config {
	return &Config{
		HTTPPort:    getEnv("HTTP_PORT", "8095"),
		GRPCPort:    getEnv("GRPC_PORT", "50055"),
		Env:         getEnv("ENV", "development"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/event_store?sslmode=disable"),
		DBMaxConns:  getInt("DB_MAX_CONNS", 20),
		DBTimeout:   getDuration("DB_TIMEOUT", 5*time.Second),
		PageSizeMax: getInt("PAGE_SIZE_MAX", 1000),
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
