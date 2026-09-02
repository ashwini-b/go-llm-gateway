package cache

import (
	"llm-gateway/internal/model"
	"sync"
	"time"
)

type entry struct {
	response  model.ChatResponse
	expiresAt time.Time
}
type Cache struct {
	mu      sync.RWMutex
	entries map[string]entry
	ttl     time.Duration
}

func NewCache(ttl time.Duration) *Cache {
	c := &Cache{
		entries: make(map[string]entry),
		ttl:     ttl,
	}
	go c.startSweeper(30 * time.Second)
	return c
}

func (c *Cache) Get(key string) (model.ChatResponse, bool) {
	c.mu.RLock()
	e, ok := c.entries[key]
	c.mu.RUnlock()

	if !ok || time.Now().After(e.expiresAt) {
		return model.ChatResponse{}, false
	}
	return e.response, true
}

func (c *Cache) Set(key string, resp model.ChatResponse) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[key] = entry{
		response:  resp,
		expiresAt: time.Now().Add(c.ttl),
	}
}
func (c *Cache) startSweeper(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		c.mu.Lock()
		for k, e := range c.entries {
			if now.After(e.expiresAt) {
				delete(c.entries, k)
			}
		}
		c.mu.Unlock()
	}
}
