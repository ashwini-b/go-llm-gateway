package handler

import (
	"encoding/json"
	"llm-gateway/internal/model"
	"llm-gateway/internal/provider"
	"log"
	"net/http"
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
func ChatCompletions(registry *provider.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req model.ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid Request Body", http.StatusBadRequest)
			return
		}
		if req.Model == "" {
			http.Error(w, "field 'model' is required", http.StatusBadRequest)
			return
		}

		if len(req.Messages) == 0 {
			http.Error(w, "field 'messages' must not be empty", http.StatusBadRequest)
			return
		}
		p, err := registry.Get(req.Model)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		resp, err := p.Complete(r.Context(), req)
		if err != nil {
			log.Printf("provider error for model %q: %v", req.Model, err)
			http.Error(w, "upstream provider error", http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}
