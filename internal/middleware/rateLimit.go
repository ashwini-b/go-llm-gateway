package middleware

import (
	"net/http"
	"sync"
)
import "golang.org/x/time/rate"

type RateLimiter struct {
	mu       sync.RWMutex
	limiters map[string]*rate.Limiter
	r        rate.Limit
	b        int
}

func NewRateLimiter(r rate.Limit, b int) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		r:        r,
		b:        b,
	}
}

func (rl *RateLimiter) getLimiter(key string) *rate.Limiter {
	rl.mu.RLock()
	limiter, ok := rl.limiters[key]
	rl.mu.RUnlock()
	if ok {
		return limiter
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()
	// Recheck: another goroutine may have created it between the RUnlock above and this Lock.
	if limiter, ok = rl.limiters[key]; ok {
		return limiter
	}
	limiter = rate.NewLimiter(rl.r, rl.b)
	rl.limiters[key] = limiter
	return limiter
}
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		apiKey, ok := req.Context().Value(apiKeyContextKey).(string)
		if !ok {
			// Should be unreachable if middleware ordering is correct —
			// APIKeyAuth always sets this before RateLimiter runs.
			http.Error(w, "internal error: missing api key", http.StatusInternalServerError)
			return
		}

		if !rl.getLimiter(apiKey).Allow() {
			w.Header().Set("Retry-After", "1")
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, req)
	})
}
