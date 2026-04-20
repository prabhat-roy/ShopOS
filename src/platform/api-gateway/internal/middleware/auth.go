package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const (
	UserIDKey   contextKey = "user_id"
	UserRoleKey contextKey = "user_role"
)

func Auth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isPublicPath(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				jsonError(w, "missing or invalid authorization header", http.StatusUnauthorized)
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			claims := jwt.MapClaims{}

			token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(jwtSecret), nil
			})

			if err != nil || !token.Valid {
				jsonError(w, "invalid or expired token", http.StatusUnauthorized)
				return
			}

			ctx := r.Context()
			if uid, ok := claims["sub"].(string); ok {
				ctx = context.WithValue(ctx, UserIDKey, uid)
				r.Header.Set("X-User-ID", uid)
			}
			if role, ok := claims["role"].(string); ok {
				ctx = context.WithValue(ctx, UserRoleKey, role)
				r.Header.Set("X-User-Role", role)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func isPublicPath(path string) bool {
	switch path {
	case "/healthz", "/metrics":
		return true
	}
	return strings.HasPrefix(path, "/auth/")
}
