package config

import (
	"os"
	"strings"
	"time"
)

type Config struct {
	Port             string
	Env              string
	GRPCTimeout      time.Duration
	ValidAPIKeys     map[string]string // apiKey -> partnerID
	CatalogAddr      string
	InventoryAddr    string
	OrderServiceAddr string
	WebhookAddr      string
	OrgServiceAddr   string
	ContractAddr     string
	QuoteAddr        string
}

func Load() *Config {
	return &Config{
		Port:             getEnv("PORT", "8083"),
		Env:              getEnv("ENV", "development"),
		GRPCTimeout:      getDuration("GRPC_TIMEOUT", 10*time.Second),
		ValidAPIKeys:     parseAPIKeys(getEnv("PARTNER_API_KEYS", "")),
		CatalogAddr:      getEnv("CATALOG_SERVICE_ADDR", "product-catalog-service:50070"),
		InventoryAddr:    getEnv("INVENTORY_SERVICE_ADDR", "inventory-service:50074"),
		OrderServiceAddr: getEnv("ORDER_SERVICE_ADDR", "order-service:50082"),
		WebhookAddr:      getEnv("WEBHOOK_SERVICE_ADDR", "webhook-service:8091"),
		OrgServiceAddr:   getEnv("ORG_SERVICE_ADDR", "organization-service:50160"),
		ContractAddr:     getEnv("CONTRACT_SERVICE_ADDR", "contract-service:50161"),
		QuoteAddr:        getEnv("QUOTE_SERVICE_ADDR", "quote-rfq-service:50162"),
	}
}

// parseAPIKeys parses "key1:partnerA,key2:partnerB" format
func parseAPIKeys(raw string) map[string]string {
	keys := make(map[string]string)
	if raw == "" {
		return keys
	}
	for _, pair := range strings.Split(raw, ",") {
		parts := strings.SplitN(strings.TrimSpace(pair), ":", 2)
		if len(parts) == 2 {
			keys[parts[0]] = parts[1]
		}
	}
	return keys
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
