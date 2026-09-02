package cache

import (
	"testing"
	"time"

	"llm-gateway/internal/model"
)

func float64Ptr(f float64) *float64 { return &f }

func sampleRequest(temp *float64) model.ChatRequest {
	return model.ChatRequest{
		Model: "llama3",
		Messages: []model.Message{
			{Role: "user", Content: "What is the capital of France?"},
		},
		Temperature: temp,
	}
}

func sampleResponse() model.ChatResponse {
	return model.ChatResponse{
		ID:     "chatcmpl-test",
		Object: "chat.completion",
		Model:  "llama3",
		Choices: []model.Choice{
			{
				Message:      model.Message{Role: "assistant", Content: "Paris."},
				FinishReason: "stop",
			},
		},
	}
}

func TestCache_SetThenGet_HitsWithinTTL(t *testing.T) {
	c := NewCache(1 * time.Minute)
	req := sampleRequest(float64Ptr(0))
	key := Key(req)
	resp := sampleResponse()

	c.Set(key, resp)

	got, ok := c.Get(key)
	if !ok {
		t.Fatal("expected a hit immediately after Set, got a miss")
	}
	if got.ID != resp.ID {
		t.Errorf("got response ID %q, want %q", got.ID, resp.ID)
	}
}

func TestCache_Get_MissesAfterTTLExpires(t *testing.T) {
	c := NewCache(10 * time.Millisecond)
	req := sampleRequest(float64Ptr(0))
	key := Key(req)

	c.Set(key, sampleResponse())
	time.Sleep(20 * time.Millisecond)

	if _, ok := c.Get(key); ok {
		t.Fatal("expected a miss after TTL elapsed, got a hit")
	}
}

func TestCache_Get_MissesForUnknownKey(t *testing.T) {
	c := NewCache(1 * time.Minute)

	if _, ok := c.Get("never-set"); ok {
		t.Fatal("expected a miss for a key that was never Set, got a hit")
	}
}

func TestKey_DifferentRequests_ProduceDifferentKeys(t *testing.T) {
	reqA := sampleRequest(float64Ptr(0))
	reqB := model.ChatRequest{
		Model: "llama3",
		Messages: []model.Message{
			{Role: "user", Content: "What is the capital of Germany?"}, // different content
		},
		Temperature: float64Ptr(0),
	}

	if Key(reqA) == Key(reqB) {
		t.Fatal("expected different requests to produce different cache keys, got the same key")
	}
}

func TestKey_IdenticalRequests_ProduceSameKey(t *testing.T) {
	reqA := sampleRequest(float64Ptr(0))
	reqB := sampleRequest(float64Ptr(0))

	if Key(reqA) != Key(reqB) {
		t.Fatal("expected identical requests to produce the same cache key, got different keys")
	}
}

func TestIsCacheable(t *testing.T) {
	tests := []struct {
		name string
		req  model.ChatRequest
		want bool
	}{
		{"temperature explicitly 0 is cacheable", sampleRequest(float64Ptr(0)), true},
		{"temperature above 0 is not cacheable", sampleRequest(float64Ptr(0.8)), false},
		{"unset temperature is not cacheable", sampleRequest(nil), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsCacheable(tt.req); got != tt.want {
				t.Errorf("IsCacheable() = %v, want %v", got, tt.want)
			}
		})
	}
}
