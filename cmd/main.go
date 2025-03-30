package main

import (
	"context"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/vishaldc/go-cache/internal/handlers"
	"github.com/vishaldc/go-cache/internal/log"
	"github.com/vishaldc/go-cache/internal/registry"
	"go.uber.org/zap"
)

func main() {
	config := registry.LoadConfiguration()
	registry.Setup(config)
	// Create a context that listens for SIGTERM or SIGINT
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	go func() {
		// start a different server on a different port for the sync handlers
		// Sync handlers
		http.HandleFunc("POST /cache/sync", handlers.SyncPostHandler)
		http.HandleFunc("DELETE /cache/sync", handlers.SyncDeleteHandler)
		log.Logger.Info("starting sync server on:", zap.String("port", config.SyncPort))
		if err := http.ListenAndServe(fmt.Sprintf(":%s", config.SyncPort), nil); err != nil {
			log.Logger.Fatal("could not start sync server:", zap.String("error", err.Error()))
		}
	}()

	// Main server
	go func() {
		http.HandleFunc("GET /cache", handlers.GetHandler)
		http.HandleFunc("POST /cache", handlers.PostHandler)
		http.HandleFunc("DELETE /cache", handlers.DeleteHandler)

		log.Logger.Info("starting server on:", zap.String("port", config.ServerPort))
		if err := http.ListenAndServe(fmt.Sprintf(":%s", config.ServerPort), nil); err != nil {
			log.Logger.Fatal("could not start server:", zap.String("error", err.Error()))
		}
	}()

	// Wait for SIGTERM or SIGINT
	<-ctx.Done()
	log.Logger.Info("shutdown signal received")

}
