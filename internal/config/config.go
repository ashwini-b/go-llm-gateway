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
	Name     string `yaml:"name"`
	Provider string `yaml:"provider"`
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

func BuildRegistry(cfg *Config) (*provider.Registry, error) {
	reg := provider.NewRegistry()
	for _, m := range cfg.Models {
		var p provider.Provider
		switch m.Provider {
		case "ollama":
			p = provider.NewOllamaProvider(m.BaseURL)
		default:
			return nil, fmt.Errorf("unknown provider type %q for model %q", m.Provider, m.Name)
		}
		reg.Register(m.Name, p)
	}
	return reg, nil
}
