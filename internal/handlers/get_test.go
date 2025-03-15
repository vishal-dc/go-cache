package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vishaldc/go-cache/internal/cache"
)

func TestGetHandler(t *testing.T) {
	// Set up a test cache item
	key := "testKey"
	value := map[string]any{"field1": "value1", "field2": 2}
	cache.Set(key, value)

	// Create a request to pass to our handler
	req, err := http.NewRequest("GET", "/get?key="+key, nil)
	assert.NoError(t, err)

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(GetHandler)

	// Call the handler
	handler.ServeHTTP(rr, req)

	// Check the status code
	assert.Equal(t, http.StatusOK, rr.Code)

	// Check the response body
	expected := `{"field1":"value1","field2":2}`
	assert.JSONEq(t, expected, rr.Body.String())
}

func TestGetHandlerMissingKey(t *testing.T) {
	// Create a request to pass to our handler
	req, err := http.NewRequest("GET", "/get", nil)
	assert.NoError(t, err)

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(GetHandler)

	// Call the handler
	handler.ServeHTTP(rr, req)

	// Check the status code
	assert.Equal(t, http.StatusBadRequest, rr.Code)

	// Check the response body
	assert.Equal(t, "missing key in request\n", rr.Body.String())
}

func TestGetHandlerInvalidMethod(t *testing.T) {
	// Create a request to pass to our handler
	req, err := http.NewRequest("POST", "/get?key=testKey", nil)
	assert.NoError(t, err)

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(GetHandler)

	// Call the handler
	handler.ServeHTTP(rr, req)

	// Check the status code
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)

	// Check the response body
	assert.Equal(t, "invalid request method\n", rr.Body.String())
}

func TestGetHandlerKeyNotFound(t *testing.T) {
	// Create a request to pass to our handler
	req, err := http.NewRequest("GET", "/get?key=nonExistentKey", nil)
	assert.NoError(t, err)

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(GetHandler)

	// Call the handler
	handler.ServeHTTP(rr, req)

	// Check the status code
	assert.Equal(t, http.StatusNotFound, rr.Code)

	assert.Equal(t, "key not found in cache\n", rr.Body.String())
}
