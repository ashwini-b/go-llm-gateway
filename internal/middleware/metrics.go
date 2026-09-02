// internal/middleware/metrics.go
package middleware

import (
	"net/http"
	"strconv"
	"time"

	chimw "github.com/go-chi/chi/v5/middleware" // aliased — your own package is also named "middleware"

	"llm-gateway/internal/metrics"
)

// HTTPMetrics records request count and duration for every request that
// passes through it, labeled by method, path, and status code.
func HTTPMetrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		ww := chimw.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)

		status := strconv.Itoa(ww.Status())
		metrics.HTTPRequestsTotal.WithLabelValues(r.Method, r.URL.Path, status).Inc()
		metrics.HTTPRequestDuration.WithLabelValues(r.URL.Path).Observe(time.Since(start).Seconds())
	})
}
