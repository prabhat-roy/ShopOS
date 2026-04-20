package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all runtime configuration for the credit-service.
type Config struct {
	HTTPPort    string
	GRPCPort    string
	DatabaseURL string
	LogLevel    string
}

// Load reads configuration from environment variables and returns a Config.
// Sensible defaults are applied where appropriate.
func Load() (*Config, error) {
	cfg := &Config{
		HTTPPort: getEnv("HTTP_PORT", "8306"),
		GRPCPort: getEnv("GRPC_PORT", "50115"),
		LogLevel: getEnv("LOG_LEVEL", "info"),
	}

	dbURL, err := buildDatabaseURL()
	if err != nil {
		return nil, fmt.Errorf("config: building database URL: %w", err)
	}
	cfg.DatabaseURL = dbURL

	return cfg, nil
}

// buildDatabaseURL constructs a PostgreSQL connection string from individual
// environment variables or returns DATABASE_URL directly if set.
func buildDatabaseURL() (string, error) {
	if url := os.Getenv("DATABASE_URL"); url != "" {
		return url, nil
	}

	host := getEnv("DB_HOST", "localhost")
	portStr := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := os.Getenv("DB_PASSWORD")
	dbname := getEnv("DB_NAME", "credit_service")
	sslmode := getEnv("DB_SSLMODE", "disable")

	if _, err := strconv.Atoi(portStr); err != nil {
		return "", fmt.Errorf("config: DB_PORT %q is not a valid integer", portStr)
	}

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, portStr, user, password, dbname, sslmode,
	)
	return dsn, nil
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
