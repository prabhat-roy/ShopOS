package config

import (
	"os"
	"strconv"
)

// Config holds all runtime configuration loaded from environment variables.
type Config struct {
	HTTPPort      string
	GRPCPort      string
	DatabaseURL   string
	DBMaxConns    int
	EventStoreURL string
}

// Load reads configuration from environment variables, applying defaults where
// values are absent.
func Load() *Config {
	return &Config{
		HTTPPort:      getEnv("HTTP_PORT", "8094"),
		GRPCPort:      getEnv("GRPC_PORT", "50059"),
		DatabaseURL:   getEnv("DATABASE_URL", ""),
		DBMaxConns:    getEnvInt("DB_MAX_CONNS", 10),
		EventStoreURL: getEnv("EVENT_STORE_URL", "http://event-store-service:8095"),
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
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return defaultVal
}
