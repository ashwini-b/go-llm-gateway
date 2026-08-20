package main

import (
	"llm-gateway/internal/config"
	"log"
	"net/http"
	"time"

	"llm-gateway/internal/handler"

	"github.com/go-chi/chi/v5"
)

/*
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello from the LLM gateway!")
	})

log.Println("Listening on port:8080")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
*/
func main() {

	/*registry := provider.NewRegistry()
	ollama := provider.NewOllamaProvider("http://localhost:11434")
	registry.Register("llama3", ollama)
	*/

	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	reg, err := config.BuildRegistry(cfg)
	if err != nil {
		log.Fatalf("failed to build provider registry: %v", err)
	}

	r := chi.NewRouter()
	r.Get("/healthz", handler.Healthz)
	r.Post("/v1/chat/completions", handler.ChatCompletions(reg))

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second, // generous — CPU inference is slow
	}

	log.Println("listening on :8080")
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
