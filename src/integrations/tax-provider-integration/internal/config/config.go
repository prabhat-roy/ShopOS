package config

import (
	"os"
	"strconv"

	"github.com/shopos/tax-provider-integration/internal/domain"
)

// Config holds all runtime configuration for the tax-provider-integration service.
type Config struct {
	HTTPPort        int
	GRPCPort        int
	DefaultProvider domain.TaxProvider
	LogLevel        string
}

// Load reads configuration from environment variables, applying defaults where needed.
func Load() *Config {
	defaultProvider := domain.TaxProvider(envStr("DEFAULT_PROVIDER", string(domain.ProviderInternal)))
	return &Config{
		HTTPPort:        envInt("HTTP_PORT", 8906),
		GRPCPort:        envInt("GRPC_PORT", 50175),
		DefaultProvider: defaultProvider,
		LogLevel:        envStr("LOG_LEVEL", "info"),
	}
}

func envInt(key string, defaultVal int) int {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return defaultVal
	}
	return n
}

func envStr(key, defaultVal string) string {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	return v
}
