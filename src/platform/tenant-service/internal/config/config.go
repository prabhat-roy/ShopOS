package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all runtime configuration loaded from environment variables.
type Config struct {
	HTTPPort    string
	DatabaseURL string
	DBMaxConns  int
	DBTimeout   time.Duration
}

// Load reads configuration from environment variables and returns a Config.
// Defaults: HTTP_PORT=8087, DB_MAX_CONNS=10, DB_TIMEOUT=5s.
func Load() *Config {
	port := getEnv("HTTP_PORT", "8087")
	dbURL := getEnv("DATABASE_URL", "")
	maxConns := getEnvInt("DB_MAX_CONNS", 10)
	dbTimeout := getEnvDuration("DB_TIMEOUT", 5*time.Second)

	return &Config{
		HTTPPort:    port,
		DatabaseURL: dbURL,
		DBMaxConns:  maxConns,
		DBTimeout:   dbTimeout,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		n, err := strconv.Atoi(v)
		if err == nil {
			return n
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		d, err := time.ParseDuration(v)
		if err == nil {
			return d
		}
	}
	return fallback
}
