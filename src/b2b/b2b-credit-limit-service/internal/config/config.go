package config

import (
	"fmt"
	"os"
)

// Config holds all runtime configuration for the b2b-credit-limit-service.
type Config struct {
	HTTPAddr string
	GRPCAddr string
	DSN      string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() (*Config, error) {
	httpAddr := getEnv("HTTP_ADDR", "")
	if httpAddr == "" {
		httpAddr = ":" + getEnv("HTTP_PORT", "8804")
	}
	cfg := &Config{
		HTTPAddr: httpAddr,
		GRPCAddr: getEnv("GRPC_ADDR", ":50164"),
		DSN:      buildDSN(),
	}
	if cfg.DSN == "" {
		return nil, fmt.Errorf("database DSN is required: set DATABASE_URL or individual DB_* vars")
	}
	return cfg, nil
}

func buildDSN() string {
	if url := os.Getenv("DATABASE_URL"); url != "" {
		return url
	}
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := os.Getenv("DB_PASSWORD")
	dbname := getEnv("DB_NAME", "b2b_credit")
	sslmode := getEnv("DB_SSLMODE", "disable")

	if password == "" {
		return fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s",
			host, port, user, dbname, sslmode)
	}
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
