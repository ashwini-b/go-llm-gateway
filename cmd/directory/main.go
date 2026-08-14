package main

import (
	"log"
	"net/http"

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
	r := chi.NewRouter()
	r.Get("/healthz", handler.Healthz)
	r.Post("/v1/chat/completions", handler.ChatCompletions)
	log.Println("listening on :8080")
	http.ListenAndServe(":8080", r)
}
