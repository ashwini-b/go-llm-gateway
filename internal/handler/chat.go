package handler

import (
	"encoding/json"
	"fmt"
	"llm-gateway/internal/cache"
	"llm-gateway/internal/metrics"
	"llm-gateway/internal/model"
	"llm-gateway/internal/provider"
	"log/slog"
	"net/http"
	"time"
)

/*
	func ChatCompletions(w http.ResponseWriter, r *http.Request) {
		var req model.ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid Request Body", http.StatusBadRequest)
			return
		}

		resp := model.ChatResponse{
			"chatcmpl-stub",
			"chat.completion",
			req.Model,
			[]model.Choice{
				{
					0,
					model.Message{
						Role:    "assistant",
						Content: "This is hard coded stub response"},
					"stop",
				}},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
*/

func ChatCompletions(registry *provider.Registry, respCache *cache.Cache, modelProviders map[string][]string,
	retryCfg provider.RetryConfig, cacheEnabled bool, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req model.ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Error("decode failed", "error", err) //
			http.Error(w, "Invalid Request Body", http.StatusBadRequest)
			return
		}
		if req.Model == "" {
			http.Error(w, "field 'model' is required", http.StatusBadRequest)
			return
		}
		cacheable := cacheEnabled && cache.IsCacheable(req)
		var key string
		if cacheable {
			key = cache.Key(req)
			if resp, ok := respCache.Get(key); ok {
				metrics.CacheHitsTotal.Inc()
				metrics.ChatRequestsTotal.WithLabelValues(req.Model, "200").Inc()
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
				return
			}
			metrics.CacheMissesTotal.Inc()
		}
		if len(req.Messages) == 0 {
			http.Error(w, "field 'messages' must not be empty", http.StatusBadRequest)
			return
		}
		/*p, err := registry.Get(req.Model)
		if err != nil {
			metrics.ChatRequestsTotal.WithLabelValues(req.Model, "404").Inc()
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}*/

		keys, ok := modelProviders[req.Model]
		if !ok {
			metrics.ChatRequestsTotal.WithLabelValues(req.Model, "404").Inc()
			http.Error(w, fmt.Sprintf("model %q not configured", req.Model), http.StatusNotFound)
			return
		}
		providers, err := registry.GetAll(keys)
		if err != nil {
			metrics.ChatRequestsTotal.WithLabelValues(req.Model, "500").Inc()
			logger.Error("registry inconsistent for model", "model", req.Model, "error", err)
			http.Error(w, "internal error resolving providers", http.StatusInternalServerError)
			return
		}

		start := time.Now()
		resp, err := provider.CompleteWithFailover(r.Context(), providers, req, retryCfg)
		metrics.ChatRequestDuration.WithLabelValues(req.Model).Observe(time.Since(start).Seconds())
		if err != nil {
			metrics.ChatRequestsTotal.WithLabelValues(req.Model, "502").Inc()
			logger.Error("all providers failed for model", "model", req.Model, "error", err)
			http.Error(w, "upstream provider error", http.StatusBadGateway)
			return
		}

		/*start := time.Now()
		resp, err := p.Complete(r.Context(), req)
		metrics.ChatRequestDuration.WithLabelValues(req.Model).Observe(time.Since(start).Seconds())
		if err != nil {
			metrics.ChatRequestsTotal.WithLabelValues(req.Model, "502").Inc()
			logger.Info("provider error for model %q: %v", req.Model, err)
			http.Error(w, "upstream provider error", http.StatusBadGateway)
			return
		}*/
		if cacheable {
			respCache.Set(key, resp)
		}

		metrics.ChatRequestsTotal.WithLabelValues(req.Model, "200").Inc()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}
