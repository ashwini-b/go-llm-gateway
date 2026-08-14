package provider

import (
	"context"
	"llm-gateway/internal/model"
)

type Provider interface {
	Complete(ctx context.Context, request model.ChatRequest) (model.ChatResponse, error)
}
