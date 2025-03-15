package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetAndGet(t *testing.T) {
	key := "testKey"
	value := map[string]any{"field1": "value1", "field2": 2}

	// Test Set
	err := Set(key, value)
	assert.Nil(t, err, "Expected no error on Set")

	// Test Get
	retrievedValue, err := Get(key)
	assert.Nil(t, err, "Expected no error on Get")
	assert.Equal(t, value, retrievedValue, "Expected retrieved value to match set value")
}

func TestGetNonExistentKey(t *testing.T) {
	key := "nonExistentKey"

	// Test Get
	_, err := Get(key)
	assert.Equal(t, ErrorKeyNotFound, err, "Expected error for non-existent key")
}

func TestDelete(t *testing.T) {
	key := "testKeyToDelete"
	value := value{"field1": "value1", "field2": 2}

	// Test Set
	err := Set(key, value)
	assert.Nil(t, err, "Expected no error on Set")

	// Test Delete
	Delete(key)

	// Test Get after Delete
	_, err = Get(key)
	assert.Equal(t, ErrorKeyNotFound, err, "Expected error for deleted key")
}
