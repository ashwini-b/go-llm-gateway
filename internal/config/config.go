package config

import (
	"fmt"
	"llm-gateway/internal/provider"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server    ServerConfig  `yaml:"server"`
	Models    []ModelConfig `yaml:"models"`
	Auth      AuthConfig    `yaml:"auth"`
	Log       LogConfig     `yaml:"log"`
	RateLimit RateLimiting  `yaml:"rate_limit"`
	Cache     CacheConfig   `yaml:"cache"`
}
type CacheConfig struct {
	Enabled    bool `yaml:"enabled"`
	TTLSeconds int  `yaml:"ttl_seconds"`
}
type RateLimiting struct {
	Burst int     `yaml:"burst""`
	RPS   float64 `yaml:"rps""`
}
type LogConfig struct {
	Level string `yaml:"level"`
}
type AuthConfig struct {
	APIKeys []string `yaml:"api_keys"`
}

type ServerConfig struct {
	Port int `yaml:"port"`
}

type ModelConfig struct {
	Name      string           `yaml:"name"`
	Providers []ProviderConfig `yaml:"providers"`
}

type ProviderConfig struct {
	Provider string `yaml:"provider"` // provider *type*, e.g. "ollama"
	BaseURL  string `yaml:"base_url"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error in reading config file: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}
	return &cfg, nil
}

func BuildRegistry(cfg *Config) (*provider.Registry, map[string][]string, error) {
	reg := provider.NewRegistry()
	modelProviders := make(map[string][]string)
	for _, m := range cfg.Models {
		var keys []string
		for i, pc := range m.Providers {
			var p provider.Provider
			switch pc.Provider {
			case "ollama":
				p = provider.NewOllamaProvider(pc.BaseURL)
			default:
				return nil, nil, fmt.Errorf("unknown provider type %q for model %q", pc.Provider, m.Name)
			}
			instanceKey := fmt.Sprintf("%s-%s-%d", m.Name, pc.Provider, i)
			reg.Register(instanceKey, p)
			keys = append(keys, instanceKey)
		}
		modelProviders[m.Name] = keys
	}
	return reg, modelProviders, nil
}
