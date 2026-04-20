package router

import (
	"net/http"

	"github.com/shopos/web-bff/internal/config"
	"github.com/shopos/web-bff/internal/handler"
	"github.com/shopos/web-bff/internal/middleware"
	"github.com/shopos/web-bff/internal/service"
	"go.uber.org/zap"
)

func New(cfg *config.Config) http.Handler {
	log := buildLogger(cfg.Env)

	auth    := handler.NewAuthHandler(service.NewGRPCAuthService(cfg.AuthServiceAddr))
	catalog := handler.NewCatalogHandler(service.NewGRPCCatalogService(cfg.CatalogAddr))
	cart    := handler.NewCartHandler(service.NewGRPCCartService(cfg.CartServiceAddr))
	order   := handler.NewOrderHandler(service.NewGRPCOrderService(cfg.OrderServiceAddr))
	user    := handler.NewUserHandler(service.NewGRPCUserService(cfg.UserServiceAddr))

	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", handler.Health)

	// Auth — public
	mux.HandleFunc("POST /login", auth.Login)
	mux.HandleFunc("POST /register", auth.Register)
	mux.HandleFunc("POST /refresh", auth.Refresh)

	// Catalog — public
	mux.HandleFunc("GET /products", catalog.ListProducts)
	mux.HandleFunc("GET /products/{id}", catalog.GetProduct)
	mux.HandleFunc("GET /categories", catalog.ListCategories)
	mux.HandleFunc("GET /search", catalog.Search)

	// Cart — authenticated (X-User-ID injected by api-gateway)
	mux.HandleFunc("GET /cart", cart.GetCart)
	mux.HandleFunc("POST /cart/items", cart.AddItem)
	mux.HandleFunc("PUT /cart/items/{itemId}", cart.UpdateItem)
	mux.HandleFunc("DELETE /cart/items/{itemId}", cart.RemoveItem)
	mux.HandleFunc("DELETE /cart", cart.ClearCart)

	// Orders — authenticated
	mux.HandleFunc("GET /orders", order.ListOrders)
	mux.HandleFunc("GET /orders/{id}", order.GetOrder)
	mux.HandleFunc("POST /orders", order.PlaceOrder)

	// User profile — authenticated
	mux.HandleFunc("GET /users/me", user.GetProfile)
	mux.HandleFunc("PUT /users/me", user.UpdateProfile)

	return chain(mux, middleware.Recovery(log), middleware.Logger(log))
}

func chain(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

func buildLogger(env string) *zap.Logger {
	if env == "production" {
		l, _ := zap.NewProduction()
		return l
	}
	l, _ := zap.NewDevelopment()
	return l
}
