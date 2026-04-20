package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	RedisAddr     string
	RedisPassword string
	RDBDB         int
	DefaultTTL    time.Duration

	Brokers  []string
	GroupID  string

	TopicProductViewed   string
	TopicCartAbandoned   string
	TopicOrderPlaced     string
	TopicInventoryLow    string
	TopicSearchPerformed string
}

func Load() *Config {
	db, _ := strconv.Atoi(env("REDIS_DB", "0"))
	ttl, _ := time.ParseDuration(env("CACHE_DEFAULT_TTL", "10m"))

	return &Config{
		RedisAddr:     env("REDIS_ADDR", "localhost:6379"),
		RedisPassword: env("REDIS_PASSWORD", ""),
		RDBDB:         db,
		DefaultTTL:    ttl,

		Brokers: strings.Split(env("KAFKA_BROKERS", "localhost:9092"), ","),
		GroupID: env("KAFKA_GROUP_ID", "cache-warming-service"),

		TopicProductViewed:   env("TOPIC_PRODUCT_VIEWED", "analytics.product.clicked"),
		TopicCartAbandoned:   env("TOPIC_CART_ABANDONED", "commerce.cart.abandoned"),
		TopicOrderPlaced:     env("TOPIC_ORDER_PLACED", "commerce.order.placed"),
		TopicInventoryLow:    env("TOPIC_INVENTORY_LOW", "supplychain.inventory.low"),
		TopicSearchPerformed: env("TOPIC_SEARCH_PERFORMED", "analytics.search.performed"),
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
