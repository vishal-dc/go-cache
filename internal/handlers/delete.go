package handlers

import (
	"net/http"

	"github.com/vishaldc/go-cache/internal/cache"
	"github.com/vishaldc/go-cache/internal/log"
	"github.com/vishaldc/go-cache/internal/registry"
	"go.uber.org/zap"
)

func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "invalid request method", http.StatusMethodNotAllowed)
		return
	}

	key := r.URL.Query().Get("key")
	if key == "" {
		log.Logger.Warn("missing key in request")
		http.Error(w, "missing key in request", http.StatusBadRequest)
		return
	}
	cache.Delete(key)
	w.WriteHeader(http.StatusNoContent)

	reg := registry.GetRegistry()
	go func() {
		err := reg.DeleteFromPool(key)
		if err != nil {
			log.Logger.Error("failed to write to pool", zap.Error(err))
		}
	}()

	log.Logger.Debug("delete request completed", zap.String("key", key))
}

func SyncDeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "invalid request method", http.StatusMethodNotAllowed)
		return
	}

	key := r.URL.Query().Get("key")
	if key == "" {
		log.Logger.Warn("missing key in request")
		http.Error(w, "missing key in request", http.StatusBadRequest)
		return
	}
	cache.Delete(key)
	w.WriteHeader(http.StatusNoContent)

	log.Logger.Debug("delete request completed", zap.String("key", key))
}
