package main

import (
	"llm-gateway/internal/provider"
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

	registry := provider.NewRegistry()
	ollama := provider.NewOllamaProvider("http://localhost:11434")
	registry.Register("llama3", ollama)

	r := chi.NewRouter()
	r.Get("/healthz", handler.Healthz)
	r.Post("/v1/chat/completions", handler.ChatCompletions(registry))

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
