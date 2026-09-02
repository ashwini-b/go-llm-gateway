package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"llm-gateway/internal/model"
)

// Key returns a deterministic identifier for req. Two requests with
// identical model, messages, and temperature always produce the same key.
func Key(req model.ChatRequest) string {
	// encoding/json marshals struct fields in fixed declaration order,
	// unlike a map — this is what makes the hash deterministic.
	b, err := json.Marshal(req)
	if err != nil {
		// ChatRequest is a plain data struct with no cyclic references
		// or unsupported field types, so Marshal cannot fail here in
		// practice — but treat it as a cache-miss key rather than panic.
		return ""
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// isCacheable reports whether req is safe to serve from cache. Only
// requests that explicitly ask for deterministic output (temperature 0)
// are cached — anything else (including an unset temperature, whose
// actual default is unknown here) is served fresh every time.
func IsCacheable(req model.ChatRequest) bool {
	return req.Temperature != nil && *req.Temperature == 0
}
