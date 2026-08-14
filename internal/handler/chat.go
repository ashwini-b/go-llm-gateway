package handler

import (
	"encoding/json"
	"llm-gateway/internal/model"
	"net/http"
)

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
