package main

import (
	"log"
	"net/http"

	"github.com/vishaldc/go-cache/internal/handlers"
)

func main() {
	http.HandleFunc("GET /cache", handlers.GetHandler)
	http.HandleFunc("POST /cache", handlers.PostHandler)
	http.HandleFunc("DELETE /cache", handlers.DeleteHandler)

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Could not start server: %s\n", err.Error())
	}
}
