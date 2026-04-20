package router

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/shopos/api-gateway/internal/config"
	"github.com/shopos/api-gateway/internal/middleware"
	"github.com/shopos/api-gateway/internal/proxy"
	"go.uber.org/zap"
)

func New(cfg *config.Config) http.Handler {
	log := buildLogger(cfg.Env)

	webProxy     := proxy.New(cfg.WebBFFAddr, "/web", cfg.UpstreamTimeout, log)
	mobileProxy  := proxy.New(cfg.MobileBFFAddr, "/mobile", cfg.UpstreamTimeout, log)
	partnerProxy := proxy.New(cfg.PartnerBFFAddr, "/partner", cfg.UpstreamTimeout, log)
	adminProxy   := proxy.New(cfg.AdminPortalAddr, "/admin", cfg.UpstreamTimeout, log)
	authProxy    := proxy.New(cfg.WebBFFAddr, "/auth", cfg.UpstreamTimeout, log)

	rl := middleware.NewRateLimiter(cfg.RateLimitRPS, cfg.RateLimitBurst, log)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthHandler)
	mux.HandleFunc("GET /metrics", metricsHandler)
	mux.Handle("/auth/", authProxy)
	mux.Handle("/web/", webProxy)
	mux.Handle("/mobile/", mobileProxy)
	mux.Handle("/partner/", partnerProxy)
	mux.Handle("/admin/", adminProxy)

	return chain(
		mux,
		requestIDMw,
		middleware.Recovery(log),
		middleware.CORS(cfg.CORSOrigins),
		middleware.Logger(log),
		rl.Middleware,
		middleware.Auth(cfg.JWTSecret),
	)
}

func chain(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

func requestIDMw(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = uuid.NewString()
			r.Header.Set("X-Request-ID", id)
		}
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r)
	})
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func metricsHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	w.Write([]byte("# Prometheus metrics — Phase 4 (OpenTelemetry)\n"))
}

func buildLogger(env string) *zap.Logger {
	if env == "production" {
		l, _ := zap.NewProduction()
		return l
	}
	l, _ := zap.NewDevelopment()
	return l
}
