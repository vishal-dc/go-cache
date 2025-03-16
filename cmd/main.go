package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/vishaldc/go-cache/internal/handlers"
	"github.com/vishaldc/go-cache/internal/log"
	"go.uber.org/zap"
)

func main() {
	http.HandleFunc("GET /cache", handlers.GetHandler)
	http.HandleFunc("POST /cache", handlers.PostHandler)
	http.HandleFunc("DELETE /cache", handlers.DeleteHandler)
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		log.Logger.Fatal("SERVER_PORT environment variable is missing")
	}
	log.Logger.Info("starting server on :", zap.String("port", port))
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil); err != nil {
		log.Logger.Fatal("could not start server:", zap.String("error", err.Error()))
	}
}
