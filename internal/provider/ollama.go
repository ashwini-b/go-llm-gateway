package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"llm-gateway/internal/model"
	"net/http"
	"time"
)

type ollamaOptions struct {
	Temperature *float64 `json:"temperature,omitempty"`
}

type ollamaChatRequest struct {
	Model    string          `json:"model"`
	Messages []ollamaMessage `json:"messages"`
	Stream   bool            `json:"stream"`
	Options  *ollamaOptions  `json:"options,omitempty"`
}

type ollamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ollamaChatResponse struct {
	Message ollamaMessage `json:"message"`
	Done    bool          `json:"done"`
}
type OllamaProvider struct {
	baseURL    string
	httpClient *http.Client
}

func NewOllamaProvider(baseURL string) *OllamaProvider {
	return &OllamaProvider{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (o *OllamaProvider) Complete(ctx context.Context, req model.ChatRequest) (model.ChatResponse, error) {
	// translate model.ChatRequest -> ollamaChatRequest, call Ollama, translate back

	ollamaMsg := make([]ollamaMessage, len(req.Messages))
	var opts *ollamaOptions
	if req.Temperature != nil {
		opts = &ollamaOptions{Temperature: req.Temperature}
	}

	for i, m := range req.Messages {
		ollamaMsg[i] = ollamaMessage{Role: m.Role,
			Content: m.Content}
	}
	ollamaReq := ollamaChatRequest{Model: req.Model,
		Messages: ollamaMsg,
		Stream:   false,
		Options:  opts,
	}

	payload, err := json.Marshal(ollamaReq)
	if err != nil {
		return model.ChatResponse{}, fmt.Errorf("marshal ollama request: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(
		ctx, http.MethodPost, o.baseURL+"/api/chat", bytes.NewReader(payload),
	)
	if err != nil {
		return model.ChatResponse{}, fmt.Errorf("build ollama request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := o.httpClient.Do(httpReq)
	if err != nil {
		return model.ChatResponse{}, fmt.Errorf("call ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return model.ChatResponse{}, fmt.Errorf("ollama returned status %d", resp.StatusCode)
	}
	var ollamaResp ollamaChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return model.ChatResponse{}, fmt.Errorf("decode ollama response: %w", err)
	}

	finishReason := "stop"
	if !ollamaResp.Done {
		finishReason = "length"
	}
	return model.ChatResponse{
		Object: "chat.completion",
		Model:  req.Model,
		Choices: []model.Choice{
			{
				Index: 0,
				Message: model.Message{
					Role:    ollamaResp.Message.Role,
					Content: ollamaResp.Message.Content,
				},
				FinishReason: finishReason,
			},
		},
	}, nil
}
