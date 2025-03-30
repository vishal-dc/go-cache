package registry

import (
	"database/sql"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite"
)

// MockRoundTripper is a custom implementation of http.RoundTripper
type MockRoundTripper struct {
	roundTripFunc func(req *http.Request) *http.Response
}

func (m *MockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.roundTripFunc(req), nil
}

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to create in-memory database: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE go_cache.workers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			worker TEXT NOT NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		);
	`)
	if err != nil {
		t.Fatalf("failed to create workers table: %v", err)
	}

	return db
}

func TestGetSelfWorker(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	reg := &defaultRegistry{
		db: db,
		self: &Worker{
			ID:        1,
			Hostname:  "localhost:8080",
			CreatedAt: time.Now(),
			Updated:   time.Now(),
		},
	}

	selfWorker := reg.GetSelfWorker()
	assert.NotNil(t, selfWorker)
	assert.Equal(t, "localhost:8080", selfWorker.Hostname)
}

func TestWriteToPool(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	reg := &defaultRegistry{
		db:   db,
		pool: make(map[string]Worker),
		client: &http.Client{
			Timeout: 1 * time.Second,
		},
	}

	worker := Worker{
		ID:        1,
		Hostname:  "localhost:8081",
		CreatedAt: time.Now(),
		Updated:   time.Now(),
	}
	reg.pool["localhost:8081"] = worker

	err := reg.WriteToPool("testKey", map[string]any{"value": "testValue"})
	assert.NoError(t, err)
}

func TestWriteToPoolWithMockClient(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create a mock HTTP client
	mockClient := &http.Client{
		Timeout: 1 * time.Second,
		Transport: &MockRoundTripper{
			roundTripFunc: func(req *http.Request) *http.Response {
				// Assert the request details
				assert.Equal(t, "POST", req.Method)
				assert.Contains(t, req.URL.String(), "localhost:8081/cache?key=testKey")

				// Return a mocked response
				return &http.Response{
					StatusCode: http.StatusNoContent,
					Body:       nil,
					Header:     make(http.Header),
				}
			},
		},
	}

	// Create a registry with the mock client
	reg := &defaultRegistry{
		db:     db,
		pool:   make(map[string]Worker),
		client: mockClient,
	}

	worker := Worker{
		ID:        1,
		Hostname:  "localhost:8081",
		CreatedAt: time.Now(),
		Updated:   time.Now(),
	}
	reg.pool["localhost:8081"] = worker

	// Call the method to test
	err := reg.WriteToPool("testKey", map[string]any{"value": "testValue"})
	assert.NoError(t, err)
}

func TestDeleteFromPool(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	reg := &defaultRegistry{
		db:   db,
		pool: make(map[string]Worker),
		client: &http.Client{
			Timeout: 1 * time.Second,
		},
	}

	worker := Worker{
		ID:        1,
		Hostname:  "localhost:8081",
		CreatedAt: time.Now(),
		Updated:   time.Now(),
	}
	reg.pool["localhost:8081"] = worker

	err := reg.DeleteFromPool("testKey")
	assert.NoError(t, err)
}

func TestRefreshPool(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	_, err := db.Exec(`
		INSERT INTO go_cache.workers (worker, created_at, updated_at)
		VALUES ('localhost:8081', ?, ?);
	`, time.Now(), time.Now())
	assert.NoError(t, err)

	_, err = db.Exec(`
	INSERT INTO go_cache.workers (worker, created_at, updated_at)
	VALUES ('localhost:8080', ?, ?);
`, time.Now(), time.Now())
	assert.NoError(t, err)

	reg := &defaultRegistry{
		db:   db,
		pool: make(map[string]Worker),
		self: &Worker{
			Hostname: "localhost:8080",
		},
	}

	err = reg.RefreshPool()
	assert.NoError(t, err)
	assert.Len(t, reg.pool, 1)
	// Check if the pool contains the correct worker
	assert.Contains(t, reg.pool, "localhost:8081")
}

func TestCleanup(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	oldTime := time.Now().Add(-2 * STALE_WORKER_PERIOD)
	_, err := db.Exec(`
		INSERT INTO go_cache.workers (worker, created_at, updated_at)
		VALUES ('localhost:8081', ?, ?);
	`, oldTime, oldTime)
	assert.NoError(t, err)

	reg := &defaultRegistry{
		db: db,
	}

	reg.Cleanup()

	var count int
	err = db.QueryRow(`SELECT COUNT(*) FROM go_cache.workers`).Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}
