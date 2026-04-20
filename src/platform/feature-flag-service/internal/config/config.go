package config

import (
	"os"
	"time"
)

type Config struct {
	HTTPPort    string
	GRPCPort    string
	Env         string
	DatabaseURL string
	DBTimeout   time.Duration
}

func Load() *Config {
	return &Config{
		HTTPPort:    getEnv("HTTP_PORT", "8092"),
		GRPCPort:    getEnv("GRPC_PORT", "50052"),
		Env:         getEnv("ENV", "development"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/feature_flags?sslmode=disable"),
		DBTimeout:   getDuration("DB_TIMEOUT", 5*time.Second),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
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
