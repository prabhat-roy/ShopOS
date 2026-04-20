package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all runtime configuration for the video-service.
type Config struct {
	HTTPPort      string
	GRPCPort      string
	MinioEndpoint string
	MinioAccess   string
	MinioSecret   string
	MinioBucket   string
	MinioUseSSL   bool
	StreamExpiry  time.Duration
}

// Load reads configuration from environment variables and applies defaults.
func Load() *Config {
	return &Config{
		HTTPPort:      getEnv("HTTP_PORT", "8604"),
		GRPCPort:      getEnv("GRPC_PORT", "50144"),
		MinioEndpoint: getEnv("MINIO_ENDPOINT", "localhost:9000"),
		MinioAccess:   getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		MinioSecret:   getEnv("MINIO_SECRET_KEY", "minioadmin"),
		MinioBucket:   getEnv("MINIO_BUCKET", "videos"),
		MinioUseSSL:   getBoolEnv("MINIO_USE_SSL", false),
		StreamExpiry:  getDurationEnv("STREAM_EXPIRY", 4*time.Hour),
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
