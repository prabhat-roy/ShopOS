package middleware

import (
	"context"
	"net/http"

	"go.uber.org/zap"
)

type contextKey string

const PartnerIDKey contextKey = "partner_id"

func APIKey(validKeys map[string]string, log *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/healthz" || r.URL.Path == "/metrics" {
				next.ServeHTTP(w, r)
				return
			}

			key := r.Header.Get("X-API-Key")
			if key == "" {
				jsonError(w, "missing X-API-Key header", http.StatusUnauthorized)
				return
			}

			partnerID, ok := validKeys[key]
			if !ok {
				log.Warn("invalid API key", zap.String("path", r.URL.Path))
				jsonError(w, "invalid API key", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), PartnerIDKey, partnerID)
			r.Header.Set("X-Partner-ID", partnerID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
