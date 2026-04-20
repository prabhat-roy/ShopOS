package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Port            string
	Env             string
	JWTSecret       string
	CORSOrigins     []string
	RateLimitRPS    float64
	RateLimitBurst  int
	WebBFFAddr      string
	MobileBFFAddr   string
	PartnerBFFAddr  string
	AdminPortalAddr string
	UpstreamTimeout time.Duration
}

func Load() *Config {
	return &Config{
		Port:            getEnv("PORT", "8080"),
		Env:             getEnv("ENV", "development"),
		JWTSecret:       getEnv("JWT_SECRET", "changeme-in-production"),
		CORSOrigins:     strings.Split(getEnv("CORS_ALLOWED_ORIGINS", "*"), ","),
		RateLimitRPS:    getFloat("RATE_LIMIT_RPS", 100),
		RateLimitBurst:  getInt("RATE_LIMIT_BURST", 200),
		WebBFFAddr:      getEnv("WEB_BFF_ADDR", "http://web-bff:8081"),
		MobileBFFAddr:   getEnv("MOBILE_BFF_ADDR", "http://mobile-bff:8082"),
		PartnerBFFAddr:  getEnv("PARTNER_BFF_ADDR", "http://partner-bff:8083"),
		AdminPortalAddr: getEnv("ADMIN_PORTAL_ADDR", "http://admin-portal:8085"),
		UpstreamTimeout: getDuration("UPSTREAM_TIMEOUT", 30*time.Second),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getFloat(key string, fallback float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return fallback
}

func getInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
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
