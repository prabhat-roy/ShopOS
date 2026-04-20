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
	CheckInterval   time.Duration // how often background checks run
	CheckTimeout    time.Duration // per-target HTTP timeout
	UnhealthyThresh int           // consecutive failures before marking unhealthy
	Targets         []Target
}

// Target is a service endpoint to probe.
type Target struct {
	Name    string
	URL     string
	Timeout time.Duration
}

func Load() *Config {
	targets := parseTargets(os.Getenv("HEALTH_TARGETS"))
	return &Config{
		Port:            getEnv("PORT", "8090"),
		Env:             getEnv("ENV", "development"),
		CheckInterval:   getDuration("CHECK_INTERVAL", 15*time.Second),
		CheckTimeout:    getDuration("CHECK_TIMEOUT", 3*time.Second),
		UnhealthyThresh: getInt("UNHEALTHY_THRESHOLD", 3),
		Targets:         targets,
	}
}

// parseTargets parses "name=url,name2=url2" format.
func parseTargets(raw string) []Target {
	if raw == "" {
		return nil
	}
	var out []Target
	for _, part := range strings.Split(raw, ",") {
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) == 2 {
			out = append(out, Target{Name: kv[0], URL: kv[1]})
		}
	}
	return out
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
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
