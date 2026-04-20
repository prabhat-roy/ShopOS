package config

import (
	"os"
	"strings"
	"time"
)

type Config struct {
	GRPCPort    string
	HTTPPort    string
	Env         string
	EtcdAddrs   []string
	EtcdTimeout time.Duration
	Prefix      string
}

func Load() *Config {
	return &Config{
		GRPCPort:    getEnv("GRPC_PORT", "50051"),
		HTTPPort:    getEnv("HTTP_PORT", "8090"),
		Env:         getEnv("ENV", "development"),
		EtcdAddrs:   strings.Split(getEnv("ETCD_ADDRS", "localhost:2379"), ","),
		EtcdTimeout: getDuration("ETCD_TIMEOUT", 5*time.Second),
		Prefix:      getEnv("CONFIG_PREFIX", "/shopos/config"),
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
