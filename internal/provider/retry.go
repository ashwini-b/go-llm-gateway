package provider

import (
	"context"
	"errors"
	"fmt"
	"llm-gateway/internal/model"
	"math/rand"
	"net"
	"time"
)

type RetryConfig struct {
	MaxAttemptsPerProvider int
	PerAttemptTimeout      time.Duration
}

func isRetryable(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}
	return false
}

func backoffDuration(attempt int) time.Duration {
	base := 100 * time.Millisecond
	d := base * time.Duration(1<<attempt)
	maxDelay := 2 * time.Second
	if d > maxDelay {
		d = maxDelay
	}
	jitter := time.Duration(rand.Int63n(int64(d) / 2))
	return d + jitter
}
func CompleteWithFailover(
	ctx context.Context,
	providers []Provider,
	req model.ChatRequest,
	cfg RetryConfig,
) (model.ChatResponse, error) {
	var lastErr error

	for _, p := range providers {
		for attempt := 0; attempt < cfg.MaxAttemptsPerProvider; attempt++ {
			attemptCtx, cancel := context.WithTimeout(ctx, cfg.PerAttemptTimeout)
			resp, err := p.Complete(attemptCtx, req)
			cancel()

			if err == nil {
				return resp, nil
			}
			lastErr = err

			if !isRetryable(err) {
				break
			}
			if attempt < cfg.MaxAttemptsPerProvider-1 {
				select {
				case <-time.After(backoffDuration(attempt)):
				case <-ctx.Done():
					return model.ChatResponse{}, ctx.Err()
				}
			}
		}
	}

	return model.ChatResponse{}, fmt.Errorf("all providers exhausted: %w", lastErr)
}
