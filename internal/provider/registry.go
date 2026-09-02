package provider

import (
	"fmt"
	"sync"
)

type Registry struct {
	mu        sync.RWMutex
	providers map[string]Provider
}

func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]Provider),
	}
}

func (r *Registry) Register(modelName string, p Provider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[modelName] = p
}

func (r *Registry) Get(modelName string) (Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.providers[modelName]
	if !ok {
		return nil, fmt.Errorf("No registered model: %s", modelName)
	}
	return p, nil
}
func (r *Registry) GetAll(names []string) ([]Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	providers := make([]Provider, 0, len(names))
	for _, name := range names {
		p, ok := r.providers[name]
		if !ok {
			return nil, fmt.Errorf("provider %q not registered", name)
		}
		providers = append(providers, p)
	}
	return providers, nil
}
