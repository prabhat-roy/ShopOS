package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all runtime configuration for the media-asset-service.
type Config struct {
	HTTPPort      string
	GRPCPort      string
	MinioEndpoint string
	MinioAccess   string
	MinioSecret   string
	MinioBucket   string
	MinioUseSSL   bool
	PresignExpiry time.Duration
}

// Load reads configuration from environment variables and returns a Config.
// Sane defaults are applied where not set.
func Load() *Config {
	return &Config{
		HTTPPort:      getEnv("HTTP_PORT", "8600"),
		GRPCPort:      getEnv("GRPC_PORT", "50140"),
		MinioEndpoint: getEnv("MINIO_ENDPOINT", "localhost:9000"),
		MinioAccess:   getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		MinioSecret:   getEnv("MINIO_SECRET_KEY", "minioadmin"),
		MinioBucket:   getEnv("MINIO_BUCKET", "media-assets"),
		MinioUseSSL:   getBoolEnv("MINIO_USE_SSL", false),
		PresignExpiry: getDurationEnv("PRESIGN_EXPIRY", time.Hour),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getBoolEnv(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}
