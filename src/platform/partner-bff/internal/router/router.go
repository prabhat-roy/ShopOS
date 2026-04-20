package router

import (
	"net/http"

	"github.com/shopos/partner-bff/internal/config"
	"github.com/shopos/partner-bff/internal/handler"
	"github.com/shopos/partner-bff/internal/middleware"
	"github.com/shopos/partner-bff/internal/service"
	"go.uber.org/zap"
)

func New(cfg *config.Config) http.Handler {
	log := buildLogger(cfg.Env)

	catalog   := handler.NewCatalogHandler(service.NewGRPCCatalogService(cfg.CatalogAddr))
	inventory := handler.NewInventoryHandler(service.NewGRPCInventoryService(cfg.InventoryAddr))
	order     := handler.NewOrderHandler(service.NewGRPCOrderService(cfg.OrderServiceAddr))
	webhook   := handler.NewWebhookHandler(service.NewGRPCWebhookService(cfg.WebhookAddr))
	b2b       := handler.NewB2BHandler(service.NewGRPCB2BService(cfg.OrgServiceAddr))

	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", handler.Health)
	mux.HandleFunc("GET /metrics", metricsHandler)

	// Catalog
	mux.HandleFunc("GET /catalog/products",      catalog.ListProducts)
	mux.HandleFunc("GET /catalog/products/{id}", catalog.GetProduct)
	mux.HandleFunc("GET /catalog/categories",    catalog.ListCategories)

	// Inventory
	mux.HandleFunc("GET /inventory/products/{id}", inventory.GetStock)
	mux.HandleFunc("POST /inventory/bulk",         inventory.GetBulkStock)

	// Orders
	mux.HandleFunc("GET /orders",      order.ListOrders)
	mux.HandleFunc("GET /orders/{id}", order.GetOrder)
	mux.HandleFunc("POST /orders",     order.PlaceOrder)

	// Webhooks
	mux.HandleFunc("GET /webhooks",        webhook.ListWebhooks)
	mux.HandleFunc("POST /webhooks",       webhook.CreateWebhook)
	mux.HandleFunc("DELETE /webhooks/{id}", webhook.DeleteWebhook)

	// B2B
	mux.HandleFunc("GET /b2b/organization", b2b.GetOrganization)
	mux.HandleFunc("GET /b2b/contracts",    b2b.ListContracts)
	mux.HandleFunc("GET /b2b/quotes",       b2b.ListQuotes)
	mux.HandleFunc("POST /b2b/quotes",      b2b.CreateQuote)

	return chain(
		mux,
		middleware.Recovery(log),
		middleware.Logger(log),
		middleware.APIKey(cfg.ValidAPIKeys, log),
	)
}

func chain(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

func metricsHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	w.Write([]byte("# Prometheus metrics — Phase 4\n"))
}

func buildLogger(env string) *zap.Logger {
	if env == "production" {
		l, _ := zap.NewProduction()
		return l
	}
	l, _ := zap.NewDevelopment()
	return l
}
