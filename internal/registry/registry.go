package registry

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/lib/pq"
	"github.com/vishaldc/go-cache/internal/log"
	"go.uber.org/zap"
)

const (
	// HeartbeatInterval is the interval at which the worker sends a heartbeat
	HEARTBEAT_INTERVAL = 15 * time.Second

	// STALE_WORKER_PERIOD is the period after which a worker is considered stale
	STALE_WORKER_PERIOD = 1 * time.Minute

	// REFRESH_POOL_INTERVAL is the interval at which the pool is refreshed
	REFRESH_POOL_INTERVAL = 1 * time.Minute

	// CLEANUP_INTERVAL is the interval at which the cleanup function is run
	CLEANUP_INTERVAL = 30 * time.Minute
)

// create a registry to hold the db connection
type defaultRegistry struct {
	db     *sql.DB
	pool   map[string]Worker
	self   *Worker
	client *http.Client
	mu     sync.RWMutex
}

// Registry defines the methods for the Registry
type Registry interface {
	GetSelfWorker() *Worker
	WriteToPool(key string, value map[string]any) error
	DeleteFromPool(key string) error
	RefreshPool() error
	Cleanup()
}

type Worker struct {
	ID        int
	Hostname  string
	Port      int
	SyncPort  int
	CreatedAt time.Time
	Updated   time.Time
}

// String returns the string representation of the worker
func (w *Worker) String() string {
	return fmt.Sprintf("[id: %d, worker: %s, createdAt: %s, updatedAt: %s]", w.ID, w.Hostname, w.CreatedAt, w.Updated)
}

var conf = defaultRegistry{
	pool: make(map[string]Worker),
}

// GetRegistry returns the registry instance
func GetRegistry() Registry {
	return &conf
}

// Setup creates the cache table in the database
func Setup(config *Configuration) {
	db, err := connectToDB(config)
	if err != nil {
		log.Logger.Fatal("failed to connect to the database", zap.String("error", err.Error()))
	}

	conf.db = db

	self := &Worker{}
	self.Hostname = config.Hostname + ":" + config.SyncPort
	port, err := strconv.Atoi(config.ServerPort)
	if err != nil {
		log.Logger.Fatal("invalid server port", zap.String("error", err.Error()))
	}
	self.Port = port

	port, err = strconv.Atoi(config.SyncPort)
	if err != nil {
		log.Logger.Fatal("invalid server port", zap.String("error", err.Error()))
	}

	self.SyncPort = port
	self.CreatedAt = time.Now()
	self.Updated = time.Now()

	exists, err := checkWorker(self)
	if err != nil {
		log.Logger.Fatal("failed to check worker", zap.String("error", err.Error()))
	}

	if exists {
		if err := updateWorker(self); err != nil {
			log.Logger.Fatal("failed to update worker", zap.String("error", err.Error()))
		}
	} else {
		if err := insertWorker(self); err != nil {
			log.Logger.Fatal("failed to insert worker", zap.String("error", err.Error()))
		}
	}

	client := &http.Client{}
	client.Timeout = 1 * time.Second
	conf.client = client
	conf.self = self
	runHeartbeat(conf.self, conf.db)
	runRefreshPool()
	runDeleteStaleWorkers()
}

// RunDeleteStaleWorkers runs the delete stale workers function
func runDeleteStaleWorkers() {
	go func() {
		for {
			conf.Cleanup()
			time.Sleep(CLEANUP_INTERVAL)
		}
	}()
}

func connectToDB(config *Configuration) (*sql.DB, error) {
	// Create the connection string
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.DBHost, config.DBPort, config.DBUser, config.DBPassword, config.DBName)
	log.Logger.Info("connecting to the database", zap.String("connection_string", connStr))
	// Connect to the database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	// Verify the connection
	if err := db.Ping(); err != nil {
		return nil, err
	}
	log.Logger.Info("successfully connected to the database")
	return db, nil
}

func checkWorker(worker *Worker) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM go_cache.workers WHERE worker = $1)`
	var exists bool
	err := conf.db.QueryRow(query, worker.Hostname).Scan(&exists)
	if err != nil {
		log.Logger.Error("failed to check worker", zap.String("error", err.Error()))
		return false, err
	}
	return exists, nil
}

// insertWorker inserts a new worker into the workers table
func insertWorker(worker *Worker) error {
	query := `INSERT INTO go_cache.workers (worker, created_at, updated_at) VALUES ($1, $2, $3) RETURNING id`
	var id int
	err := conf.db.QueryRow(query, worker.Hostname, worker.CreatedAt, worker.Updated).Scan(&id)
	if err != nil {
		log.Logger.Error("failed to insert worker", zap.String("error", err.Error()))
		return err
	}
	worker.ID = int(id)
	log.Logger.Info("successfully inserted worker", zap.String("worker", worker.Hostname), zap.Time("created_at", worker.CreatedAt), zap.Time("updated_at", worker.Updated))
	return nil
}

// updateWorker updates the worker in the workers table
func updateWorker(worker *Worker) error {
	query := `UPDATE go_cache.workers SET updated_at = $1 WHERE id = $2`
	_, err := conf.db.Exec(query, worker.Updated, worker.ID)
	if err != nil {
		log.Logger.Error("failed to update worker", zap.String("error", err.Error()))
		return err
	}
	log.Logger.Info("successfully updated worker", zap.String("worker", worker.Hostname), zap.Time("updated_at", worker.Updated))
	return nil
}

// getOtherWorkers returns the workers from the workers table except the current worker
func getOtherWorkers(worker *Worker) ([]Worker, error) {
	query := `SELECT id, worker, created_at, updated_at FROM go_cache.workers where worker <> $1`
	rows, err := conf.db.Query(query, worker.Hostname)
	if err != nil {
		log.Logger.Error("failed to query workers", zap.String("error", err.Error()))
		return nil, err
	}
	defer rows.Close()

	var workers []Worker
	for rows.Next() {
		var w Worker
		if err := rows.Scan(&w.ID, &w.Hostname, &w.CreatedAt, &w.Updated); err != nil {
			log.Logger.Error("failed to scan worker", zap.String("error", err.Error()))
			return nil, err
		}
		// Split the hostname and port
		parts := strings.Split(w.Hostname, ":")
		if len(parts) == 2 {
			port, err := strconv.Atoi(parts[1])
			if err != nil {
				log.Logger.Error("failed to convert port to int", zap.String("error", err.Error()))
				return nil, err
			}
			w.SyncPort = port
		}
		workers = append(workers, w)
	}

	return workers, nil
}

func (r *defaultRegistry) GetSelfWorker() *Worker {
	return r.self
}

// DeleteFromPool deletes a key from the pool
func (r *defaultRegistry) DeleteFromPool(key string) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.pool) == 0 {
		log.Logger.Error("no workers in the pool")
		return nil
	}

	for _, w := range r.pool {
		log.Logger.Info("writing to worker", zap.String("worker", w.Hostname))
		req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("http://%s/cache/sync?key=%s", w.Hostname, key), nil)
		if err != nil {
			log.Logger.Error("failed to create request", zap.String("worker", w.Hostname), zap.String("error", err.Error()))
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		resp, err := r.client.Do(req)
		if err != nil {
			log.Logger.Error("failed to write to worker", zap.String("worker", w.Hostname), zap.String("error", err.Error()))
			continue
		}
		if resp.StatusCode >= http.StatusBadRequest {
			log.Logger.Error("failed to write to worker", zap.String("worker", w.Hostname), zap.Int("status_code", resp.StatusCode))
			continue
		}
		log.Logger.Info("successfully wrote to worker", zap.String("worker", w.Hostname))
	}
	return nil
}

// WriteToPool writes a key value to the list of workers in the pool
func (r *defaultRegistry) WriteToPool(key string, value map[string]any) error {
	b, err := json.Marshal(value)
	if err != nil {
		log.Logger.Error("failed to marshal value", zap.String("error", err.Error()))
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.pool) == 0 {
		log.Logger.Error("no workers in the pool")
		return nil
	}

	for _, w := range r.pool {
		log.Logger.Info("writing to worker", zap.String("worker", w.Hostname))
		resp, err := r.client.Post(fmt.Sprintf("http://%s/cache/sync?key=%s", w.Hostname, key), "application/json", bytes.NewBuffer(b))
		if err != nil {
			log.Logger.Error("failed to write to worker", zap.String("worker", w.Hostname), zap.String("error", err.Error()))
			continue
		}
		if resp.StatusCode >= http.StatusBadRequest {
			log.Logger.Error("failed to write to worker", zap.String("worker", w.Hostname), zap.Int("status_code", resp.StatusCode))
			continue
		}
		log.Logger.Info("successfully wrote to worker", zap.String("worker", w.Hostname))
	}
	return nil
}

// runRefreshPool runs the refresh pool function
func runRefreshPool() {
	go func() {
		for {
			if err := conf.RefreshPool(); err != nil {
				log.Logger.Error("failed to refresh pool", zap.String("error", err.Error()))
			}
			time.Sleep(REFRESH_POOL_INTERVAL)
		}
	}()
}

// RefreshPool refreshes the pool of workers
func (r *defaultRegistry) RefreshPool() error {
	log.Logger.Info("refreshing pool")
	workers, err := getOtherWorkers(r.self)
	if err != nil {
		log.Logger.Error("failed to get other workers", zap.String("error", err.Error()))
		return err
	}

	r.mu.Lock()
	r.pool = make(map[string]Worker)
	var workerNames []string
	for _, w := range workers {
		r.pool[w.Hostname] = w
		workerNames = append(workerNames, w.String())
	}
	log.Logger.Info("workers in the pool", zap.String("workers", strings.Join(workerNames, ", ")))
	r.mu.Unlock()
	return nil
}

// Heartbeat records the heartbeat of the worker in the workers table
func (w *Worker) Heartbeat(db *sql.DB) error {
	w.Updated = time.Now()
	query := `UPDATE go_cache.workers SET updated_at = $1 WHERE worker = $2`
	_, err := db.Exec(query, w.Updated, w.Hostname)
	if err != nil {
		log.Logger.Error("failed to update worker", zap.String("error", err.Error()))
		return err
	}
	log.Logger.Info("successfully updated worker", zap.String("worker", w.Hostname), zap.Time("updated_at", w.Updated))
	return nil
}

// runHeartbeat runs the heartbeat of the worker
func runHeartbeat(w *Worker, db *sql.DB) {
	go func() {
		for {
			if err := w.Heartbeat(db); err != nil {
				log.Logger.Error("failed to record heartbeat", zap.String("error", err.Error()))
			}
			time.Sleep(HEARTBEAT_INTERVAL)
		}
	}()
}

// Cleanup cleans workers that have not sent a heartbeat in the last 30 seconds
func (r *defaultRegistry) Cleanup() {
	query := `DELETE FROM go_cache.workers WHERE updated_at < $1`
	_, err := r.db.Exec(query, time.Now().Add(-STALE_WORKER_PERIOD))
	if err != nil {
		log.Logger.Warn("failed to cleanup workers", zap.String("error", err.Error()))
		return
	}
	log.Logger.Info("successfully cleaned up workers")
}
