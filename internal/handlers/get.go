package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/vishaldc/go-cache/internal/cache"
	"github.com/vishaldc/go-cache/internal/log"
	"go.uber.org/zap"
)

func GetHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		log.Logger.Warn("invalid request method", zap.String("method", r.Method))
		http.Error(w, "invalid request method", http.StatusMethodNotAllowed)
		return
	}

	key := r.URL.Query().Get("key")
	if key == "" {
		log.Logger.Warn("missing key in request")
		http.Error(w, "missing key in request", http.StatusBadRequest)
		return
	}
	defer log.Logger.Info("get request completed", zap.String("key", key))

	value, err := cache.Get(key)
	if err == cache.ErrorKeyNotFound {
		log.Logger.Warn("key not found in cache", zap.String("key", key))
		http.Error(w, "key not found in cache", http.StatusNotFound)
		return
	}

	if err != nil {
		log.Logger.Error("failed to get cache", zap.Error(err))
		http.Error(w, "failed to get cache", http.StatusInternalServerError)
		return
	}

	responseBody, err := json.Marshal(value)
	if err != nil {
		log.Logger.Error("failed to marshall response", zap.Error(err))
		http.Error(w, "failed to marshall response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseBody)

}
