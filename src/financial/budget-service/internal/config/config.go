package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all runtime configuration for the budget-service.
type Config struct {
	HTTPPort    string
	GRPCPort    string
	DatabaseURL string
	LogLevel    string
}

// Load reads configuration from environment variables and returns a Config.
func Load() (*Config, error) {
	cfg := &Config{
		HTTPPort: getEnv("HTTP_PORT", "8308"),
		GRPCPort: getEnv("GRPC_PORT", "50117"),
		LogLevel: getEnv("LOG_LEVEL", "info"),
	}

	dbURL, err := buildDatabaseURL()
	if err != nil {
		return nil, fmt.Errorf("config: building database URL: %w", err)
	}
	cfg.DatabaseURL = dbURL

	return cfg, nil
}

func buildDatabaseURL() (string, error) {
	if url := os.Getenv("DATABASE_URL"); url != "" {
		return url, nil
	}

	host := getEnv("DB_HOST", "localhost")
	portStr := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := os.Getenv("DB_PASSWORD")
	dbname := getEnv("DB_NAME", "budget_service")
	sslmode := getEnv("DB_SSLMODE", "disable")

	if _, err := strconv.Atoi(portStr); err != nil {
		return "", fmt.Errorf("config: DB_PORT %q is not a valid integer", portStr)
	}

	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, portStr, user, password, dbname, sslmode,
	), nil
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
