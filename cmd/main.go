package main

import (
	"llm-gateway/internal/cache"
	"llm-gateway/internal/config"
	"llm-gateway/internal/logger"
	"llm-gateway/internal/middleware"
	"llm-gateway/internal/provider"
	"log"
	"net/http"
	"os"
	"time"

	"llm-gateway/internal/handler"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/time/rate"
)

/*
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello from the LLM gateway!")
	})

log.Println("Listening on port:8080")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
*/
func main() {

	/*registry := provider.NewRegistry()
	ollama := provider.NewOllamaProvider("http://localhost:11434")
	registry.Register("llama3", ollama)
	*/

	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	reg, modelProviders, err := config.BuildRegistry(cfg)
	if err != nil {
		log.Fatalf("failed to build provider registry: %v", err)
	}

	retryCfg := provider.RetryConfig{
		MaxAttemptsPerProvider: 3,
		PerAttemptTimeout:      5 * time.Second,
	}
	validKeys := make(map[string]bool)
	for _, k := range cfg.Auth.APIKeys {
		validKeys[k] = true
	}

	rateLimiter := middleware.NewRateLimiter(rate.Limit(cfg.RateLimit.RPS), cfg.RateLimit.Burst) // new
	logger := logger.New(logger.ParseLevel(cfg.Log.Level))
	respCache := cache.NewCache(time.Duration(cfg.Cache.TTLSeconds) * time.Second)

	r := chi.NewRouter()
	r.Use(middleware.RequestLogger(logger))
	r.Use(middleware.HTTPMetrics)

	r.Get("/healthz", handler.Healthz) // no auth
	r.Handle("/metrics", promhttp.Handler())

	r.Group(func(r chi.Router) {
		r.Use(middleware.APIKeyAuth(validKeys))
		r.Use(rateLimiter.Middleware)
		r.Post("/v1/chat/completions", handler.ChatCompletions(reg, respCache, modelProviders, retryCfg, cfg.Cache.Enabled, logger))
	})

	/*r.Get("/healthz", handler.Healthz)
	r.Post("/v1/chat/completions", handler.ChatCompletions(reg))*/

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second, // generous — CPU inference is slow
	}

	logger.Info("listening on :8080")
	if err := srv.ListenAndServe(); err != nil {
		logger.Error("server failed: %v", err)
		os.Exit(1)
	}
}
