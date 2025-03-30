package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/vishaldc/go-cache/internal/cache"
	"github.com/vishaldc/go-cache/internal/log"
	"github.com/vishaldc/go-cache/internal/registry"
	"go.uber.org/zap"
)

func PostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "invalid request method", http.StatusMethodNotAllowed)
		return
	}

	key := r.URL.Query().Get("key")
	if key == "" {
		log.Logger.Warn("missing key in request")
		http.Error(w, "missing key in request", http.StatusBadRequest)
		return
	}

	jsonDecoder := json.NewDecoder(r.Body)
	var value map[string]any
	err := jsonDecoder.Decode(&value)
	if err != nil {
		log.Logger.Error("invalid request body", zap.Error(err))
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	err = cache.Set(key, value)
	if err != nil {
		log.Logger.Error("failed to set cache", zap.Error(err))
		http.Error(w, "failed to set cache", http.StatusInternalServerError)
		return
	}

	reg := registry.GetRegistry()

	go func() {
		err := reg.WriteToPool(key, value)
		if err != nil {
			log.Logger.Error("failed to write to pool", zap.Error(err))
		}
	}()

	log.Logger.Info("sync request completed", zap.String("key", key))
	w.WriteHeader(http.StatusNoContent)
}

// SyncPostHandler handles the sync request from the worker
func SyncPostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "invalid request method", http.StatusMethodNotAllowed)
		return
	}

	key := r.URL.Query().Get("key")
	if key == "" {
		log.Logger.Warn("missing key in request")
		http.Error(w, "missing key in request", http.StatusBadRequest)
		return
	}

	jsonDecoder := json.NewDecoder(r.Body)
	var value map[string]any
	err := jsonDecoder.Decode(&value)
	if err != nil {
		log.Logger.Error("invalid request body", zap.Error(err))
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	err = cache.Set(key, value)
	if err != nil {
		log.Logger.Error("failed to set cache", zap.Error(err))
		http.Error(w, "failed to set cache", http.StatusInternalServerError)
		return
	}

	log.Logger.Info("sync request completed", zap.String("key", key))
	w.WriteHeader(http.StatusNoContent)
}
