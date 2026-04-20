package config

import (
	"os"
	"strconv"
	"strings"
)

// Config holds all runtime configuration loaded from environment variables.
type Config struct {
	HTTPPort    string
	DatabaseURL string
	DBMaxConns  int
	// KafkaBrokers is a slice of broker addresses parsed from KAFKA_BROKERS.
	KafkaBrokers []string
	KafkaGroupID string
	// KafkaDLQTopics is the list of DLQ topics to consume, parsed from KAFKA_DLQ_TOPICS.
	KafkaDLQTopics []string
}

// Load reads configuration from environment variables, applying defaults where applicable.
func Load() *Config {
	cfg := &Config{
		HTTPPort:    getEnv("HTTP_PORT", "8092"),
		DatabaseURL: getEnv("DATABASE_URL", ""),
		DBMaxConns:  getEnvInt("DB_MAX_CONNS", 10),
		KafkaGroupID: getEnv("KAFKA_GROUP_ID", "dead-letter-service"),
	}

	brokers := getEnv("KAFKA_BROKERS", "")
	if brokers != "" {
		cfg.KafkaBrokers = splitTrim(brokers)
	}

	topics := getEnv("KAFKA_DLQ_TOPICS", "")
	if topics != "" {
		cfg.KafkaDLQTopics = splitTrim(topics)
	}

	return cfg
}

func getEnv(key, defaultVal string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		n, err := strconv.Atoi(v)
		if err == nil {
			return n
		}
	}
	return defaultVal
}

func splitTrim(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
