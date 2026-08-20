package main

import (
	"context"
	"fmt"
	"llm-gateway/internal/model"
	"llm-gateway/internal/provider"
	"log"
)

// Use only for manual test...........................
///////////////////////////////

func main() {
	registry := provider.NewRegistry()
	ollama := provider.NewOllamaProvider("http://localhost:11434")
	registry.Register("llama3", ollama)

	p, err := registry.Get("llama3")
	if err != nil {
		log.Fatalf("provider lookup failed: %v", err)
	}
	resp, err := p.Complete(context.Background(), model.ChatRequest{
		Model: "llama3",
		Messages: []model.Message{
			{
				Role:    "user",
				Content: "Hello",
			},
		},
	})
	if err != nil {
		log.Fatalf("Complete failed: %v", err)
	}
	if len(resp.Choices) == 0 {
		log.Fatal("no choices returned")
	}
	fmt.Println("Model:", resp.Model)
	fmt.Println("Reply:", resp.Choices[0].Message.Content)
	fmt.Println("Finish reason:", resp.Choices[0].FinishReason)
}
