package cache

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/vishaldc/go-cache/internal/log"
	"go.uber.org/zap"
)

// create a registry to hold the db connection
type Registry struct {
	db   *sql.DB
	pool map[string]Worker
	self Worker
}

type Worker struct {
	ID        int
	hostname  string
	CreatedAt time.Time
	Updated   time.Time
}

var conf Registry = Registry{
	pool: make(map[string]Worker),
}

// init creates the cache table in the database
func init() {
	db, err := connectToDB()
	if err != nil {
		log.Logger.Fatal("failed to connect to the database", zap.String("error", err.Error()))
	}

	conf.db = db

	h, err := os.Hostname()
	if err != nil {
		log.Logger.Fatal("failed to get hostname", zap.String("error", err.Error()))
	}

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		log.Logger.Fatal("SERVER_PORT environment variable is missing")
	}

	self := Worker{}
	self.hostname = h + ":" + port
	self.CreatedAt = time.Now()
	self.Updated = time.Now()
	workers, err := GetOtherWorkers(h)
	if err != nil {
		log.Logger.Fatal("failed to get other workers", zap.String("error", err.Error()))
	}

	if len(workers) > 0 {
		log.Logger.Info("found other workers", zap.Int("count", len(workers)))
		for _, w := range workers {
			log.Logger.Info("worker", zap.String("worker", w.hostname), zap.Time("created_at", w.CreatedAt), zap.Time("updated_at", w.Updated))
			conf.pool[w.hostname] = w
		}
	} else {
		log.Logger.Info("no other workers found at startup")
	}

	if err := insertWorker(&self); err != nil {
		log.Logger.Fatal("failed to insert worker", zap.String("error", err.Error()))
	}

	conf.self = self
}

func connectToDB() (*sql.DB, error) {
	// Get environment variables
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	if dbHost == "" {
		log.Logger.Fatal("DB_HOST environment variable is missing")
	}
	if dbPort == "" {
		log.Logger.Fatal("DB_PORT environment variable is missing")
	}
	if dbUser == "" {
		log.Logger.Fatal("DB_USER environment variable is missing")
	}
	if dbPassword == "" {
		log.Logger.Fatal("DB_PASSWORD environment variable is missing")
	}
	if dbName == "" {
		log.Logger.Fatal("DB_NAME environment variable is missing")
	}

	// Create the connection string
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)
	log.Logger.Info("Connecting to the database", zap.String("connection_string", connStr))
	// Connect to the database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	// Verify the connection
	if err := db.Ping(); err != nil {
		return nil, err
	}
	log.Logger.Info("Successfully connected to the database")
	return db, nil
}

// InsertWorker inserts a new worker into the workers table
func insertWorker(worker *Worker) error {

	query := `INSERT INTO go_cache.workers (worker, created_at, updated_at) VALUES ($1, $2, $3) RETURNING id`
	var id int
	err := conf.db.QueryRow(query, worker.hostname, worker.CreatedAt, worker.Updated).Scan(&id)
	if err != nil {
		log.Logger.Error("failed to insert worker", zap.String("error", err.Error()))
		return err
	}
	worker.ID = int(id)
	log.Logger.Info("successfully inserted worker", zap.String("worker", worker.hostname), zap.Time("created_at", worker.CreatedAt), zap.Time("updated_at", worker.Updated))
	return nil
}

// GetOtherWorkers returns the workers from the workers table except the current worker
func GetOtherWorkers(worker string) ([]Worker, error) {
	query := `SELECT id, worker, created_at, updated_at FROM go_cache.workers where worker <> $1`

	rows, err := conf.db.Query(query, worker)
	if err != nil {
		log.Logger.Error("failed to query workers", zap.String("error", err.Error()))
		return nil, err
	}
	defer rows.Close()

	var workers []Worker
	for rows.Next() {
		var w Worker
		if err := rows.Scan(&w.ID, &w.hostname, &w.CreatedAt, &w.Updated); err != nil {
			log.Logger.Error("failed to scan worker", zap.String("error", err.Error()))
			return nil, err
		}
		workers = append(workers, w)
	}

	return workers, nil
}
