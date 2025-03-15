package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vishaldc/go-cache/internal/cache"
)

func TestPostHandler(t *testing.T) {
	// Create a test value
	value := map[string]any{"field1": "value1", "field2": 2}
	requestBody, err := json.Marshal(value)
	assert.NoError(t, err)

	// Create a request to pass to our handler
	req, err := http.NewRequest("POST", "/post?key=testKey", bytes.NewBuffer(requestBody))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(PostHandler)

	// Call the handler
	handler.ServeHTTP(rr, req)

	// Check the status code
	assert.Equal(t, http.StatusNoContent, rr.Code)

	// Check if the item was set in the cache
	cachedValue, err := cache.Get("testKey")
	assert.NoError(t, err)
	assert.Equal(t, value, cachedValue)
}

func TestPostHandlerInvalidMethod(t *testing.T) {
	// Create a request to pass to our handler
	req, err := http.NewRequest("GET", "/post?key=testKey", nil)
	assert.NoError(t, err)

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(PostHandler)

	// Call the handler
	handler.ServeHTTP(rr, req)

	// Check the status code
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)

	// Check the response body
	assert.Equal(t, "invalid request method\n", rr.Body.String())
}

func TestPostHandlerMissingKey(t *testing.T) {
	// Create a request to pass to our handler
	req, err := http.NewRequest("POST", "/post", nil)
	assert.NoError(t, err)

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(PostHandler)

	// Call the handler
	handler.ServeHTTP(rr, req)

	// Check the status code
	assert.Equal(t, http.StatusBadRequest, rr.Code)

	// Check the response body
	assert.Equal(t, "missing key in request\n", rr.Body.String())
}

func TestPostHandlerInvalidRequestBody(t *testing.T) {
	// Create a request to pass to our handler
	req, err := http.NewRequest("POST", "/post?key=testKey", bytes.NewBuffer([]byte("invalid body")))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(PostHandler)

	// Call the handler
	handler.ServeHTTP(rr, req)

	// Check the status code
	assert.Equal(t, http.StatusBadRequest, rr.Code)

	// Check the response body
	assert.Equal(t, "invalid request body\n", rr.Body.String())
}
