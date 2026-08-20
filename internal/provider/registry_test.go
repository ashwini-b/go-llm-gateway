package provider

import (
	"context"
	"llm-gateway/internal/model"
	"testing"
)

func TestRegistry_GetAndComplete(t *testing.T) {
	registry := NewRegistry()
	ollama := NewOllamaProvider("http://localhost:11434")
	registry.Register("llama3", ollama)

	p, err := registry.Get("llama3")
	if err != nil {
		t.Fatalf("provider lookup failed: %v", err)
	}

	resp, err := p.Complete(context.Background(), model.ChatRequest{
		Model: "llama3",
		Messages: []model.Message{
			{Role: "user", Content: "Hello"},
		},
	})
	if err != nil {
		t.Fatalf("Complete failed: %v", err)
	}

	if len(resp.Choices) == 0 {
		t.Fatal("no choices returned")
	}
	if resp.Choices[0].Message.Content == "" {
		t.Error("expected non-empty content")
	}
	t.Logf("Model: %s", resp.Model)
	t.Logf("Reply: %s", resp.Choices[0].Message.Content)
	t.Logf("Finish reason: %s", resp.Choices[0].FinishReason)
}

func TestRegistry_GetUnknownModel(t *testing.T) {
	registry := NewRegistry()

	_, err := registry.Get("nonexistent-model")
	if err == nil {
		t.Fatal("expected an error for unregistered model, got nil")
	}
	t.Logf("Got expected error: %v", err)
}
