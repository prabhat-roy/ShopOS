package config

import (
	"os"
	"time"
)

type Config struct {
	Port            string
	Env             string
	GRPCTimeout     time.Duration
	AuthServiceAddr string
	UserServiceAddr string
	CatalogAddr     string
	InventoryAddr   string
	CartServiceAddr string
	OrderServiceAddr string
	PricingAddr     string
	SearchAddr      string
}

func Load() *Config {
	return &Config{
		Port:             getEnv("PORT", "8081"),
		Env:              getEnv("ENV", "development"),
		GRPCTimeout:      getDuration("GRPC_TIMEOUT", 10*time.Second),
		AuthServiceAddr:  getEnv("AUTH_SERVICE_ADDR", "auth-service:50060"),
		UserServiceAddr:  getEnv("USER_SERVICE_ADDR", "user-service:50061"),
		CatalogAddr:      getEnv("CATALOG_SERVICE_ADDR", "product-catalog-service:50070"),
		InventoryAddr:    getEnv("INVENTORY_SERVICE_ADDR", "inventory-service:50074"),
		CartServiceAddr:  getEnv("CART_SERVICE_ADDR", "cart-service:50080"),
		OrderServiceAddr: getEnv("ORDER_SERVICE_ADDR", "order-service:50082"),
		PricingAddr:      getEnv("PRICING_SERVICE_ADDR", "pricing-service:50073"),
		SearchAddr:       getEnv("SEARCH_SERVICE_ADDR", "search-service:50078"),
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
