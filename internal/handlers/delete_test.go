package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vishaldc/go-cache/internal/cache"
)

func TestDeleteHandler(t *testing.T) {
	// Set up a test cache item
	key := "testKey"
	value := map[string]any{"field1": "value1", "field2": 2}
	cache.Set(key, value)

	// Create a request to pass to our handler
	req, err := http.NewRequest("DELETE", "/delete?key="+key, nil)
	assert.NoError(t, err)

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(DeleteHandler)

	// Call the handler
	handler.ServeHTTP(rr, req)

	// Check the status code
	assert.Equal(t, http.StatusNoContent, rr.Code)

	// Check if the item was deleted from the cache
	_, err = cache.Get(key)
	assert.Equal(t, cache.ErrorKeyNotFound, err)
}

func TestDeleteHandlerMissingKey(t *testing.T) {
	// Create a request to pass to our handler
	req, err := http.NewRequest("DELETE", "/delete", nil)
	assert.NoError(t, err)

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(DeleteHandler)

	// Call the handler
	handler.ServeHTTP(rr, req)

	// Check the status code
	assert.Equal(t, http.StatusBadRequest, rr.Code)

	// Check the response body
	assert.Equal(t, "missing key in request\n", rr.Body.String())
}

func TestDeleteHandlerInvalidMethod(t *testing.T) {
	// Create a request to pass to our handler
	req, err := http.NewRequest("POST", "/delete?key=testKey", nil)
	assert.NoError(t, err)

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(DeleteHandler)

	// Call the handler
	handler.ServeHTTP(rr, req)

	// Check the status code
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)

	// Check the response body
	assert.Equal(t, "invalid request method\n", rr.Body.String())
}
