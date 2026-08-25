package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/time/rate"
)

func TestRateLimiter_PerKeyIsolation(t *testing.T) {
	rl := NewRateLimiter(rate.Limit(2), 5) // 2 req/sec, burst 5 — same as your config example

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := rl.Middleware(next)

	doRequest := func(key string) int {
		req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
		ctx := context.WithValue(req.Context(), apiKeyContextKey, key)
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		return rec.Code
	}

	t.Run("burst of 5 succeeds, 6th is rejected", func(t *testing.T) {
		for i := 0; i < 5; i++ {
			if code := doRequest("key-A"); code != http.StatusOK {
				t.Fatalf("request %d: got %d, want 200", i, code)
			}
		}
		if code := doRequest("key-A"); code != http.StatusTooManyRequests {
			t.Fatalf("6th request: got %d, want 429", code)
		}
	})

	t.Run("a different key is unaffected by key-A's exhausted bucket", func(t *testing.T) {
		if code := doRequest("key-B"); code != http.StatusOK {
			t.Fatalf("key-B first request: got %d, want 200", code)
		}
	})
}
