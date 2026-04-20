package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPPort    string
	GRPCPort    string
	DatabaseURL string
	DBMaxConns  int
	DBTimeout   time.Duration
	TickInterval time.Duration
}

func Load() *Config {
	maxConns, _ := strconv.Atoi(env("DB_MAX_CONNS", "10"))
	dbTimeout, _ := time.ParseDuration(env("DB_TIMEOUT", "5s"))
	tick, _ := time.ParseDuration(env("TICK_INTERVAL", "30s"))

	return &Config{
		HTTPPort:     env("HTTP_PORT", "8095"),
		GRPCPort:     env("GRPC_PORT", "50056"),
		DatabaseURL:  env("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/scheduler?sslmode=disable"),
		DBMaxConns:   maxConns,
		DBTimeout:    dbTimeout,
		TickInterval: tick,
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
