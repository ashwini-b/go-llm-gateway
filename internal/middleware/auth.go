package middleware

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const apiKeyContextKey contextKey = "apiKey"

func APIKeyAuth(validKeys map[string]bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				http.Error(w, "missing Authorization header", http.StatusUnauthorized)
				return
			}
			key, ok := strings.CutPrefix(header, "Bearer ")
			if !ok {
				http.Error(w, "Authorization header must use Bearer scheme", http.StatusUnauthorized)
				return
			}
			if !validKeys[key] {
				http.Error(w, "invalid API key", http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), apiKeyContextKey, key)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
