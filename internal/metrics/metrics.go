package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTPRequestsTotal counts every request the gateway handles, at the
	// transport level — labeled by method, path, and the status code written.
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gateway_http_requests_total",
			Help: "Total HTTP requests, labeled by method, path, and status.",
		},
		[]string{"method", "path", "status"},
	)

	// HTTPRequestDuration tracks how long requests take, labeled only by
	// path — enough to see "is /v1/chat/completions slower than /healthz"
	// without multiplying series by status code too.
	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "gateway_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds, labeled by path.",
			Buckets: prometheus.DefBuckets, // 0.005s .. 10s, the client's standard spread
		},
		[]string{"path"},
	)

	// ChatRequestsTotal is the business-level counterpart to HTTPRequestsTotal —
	// labeled by model, which the generic HTTP middleware can't see.
	ChatRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gateway_chat_requests_total",
			Help: "Total chat completion requests, labeled by model and status.",
		},
		[]string{"model", "status"},
	)

	// ChatRequestDuration tracks provider call latency per model — this is
	// what tells you "is llama3 slower than mistral", not just "is the
	// gateway slow".
	ChatRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "gateway_chat_request_duration_seconds",
			Help:    "Chat completion request duration in seconds, labeled by model.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"model"},
	)

	// CacheHitsTotal and CacheMissesTotal are deliberately unlabeled —
	// no model, no key. See the cardinality discussion: these answer
	// "what's my overall cache hit rate", not "which model benefits most",
	// which is a reasonable thing to add later but not needed for v1.
	CacheHitsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "gateway_cache_hits_total",
			Help: "Total response cache hits.",
		},
	)
	CacheMissesTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "gateway_cache_misses_total",
			Help: "Total response cache misses.",
		},
	)

	// RateLimitRejectionsTotal is unlabeled for the same reason —
	// specifically NOT labeled by api_key, per the cardinality trap.
	RateLimitRejectionsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "gateway_rate_limit_rejections_total",
			Help: "Total requests rejected by the per-key rate limiter.",
		},
	)
)
