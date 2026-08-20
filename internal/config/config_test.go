package config

import (
	"fmt"
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	dir, _ := os.Getwd()
	fmt.Println(dir)
	cfg, err := Load("C:\\Users\\ashwi\\IdeaProjects\\llm-gateway\\config.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Models) != 2 {
		t.Errorf("expected 2 models, got %d", len(cfg.Models))
	}
}
