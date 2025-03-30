package registry

import (
	"os"

	"github.com/vishaldc/go-cache/internal/log"
	"go.uber.org/zap"
)

type Configuration struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	ServerPort string
	SyncPort   string
	Hostname   string
}

// LoadConfiguration loads environment variables into the Configuration struct
func LoadConfiguration() *Configuration {
	config := &Configuration{
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),
		ServerPort: os.Getenv("SERVER_PORT"),
		SyncPort:   os.Getenv("SYNC_PORT"),
		Hostname:   os.Getenv("HOSTNAME"),
	}

	// Validate required environment variables
	if config.DBHost == "" {
		log.Logger.Fatal("DB_HOST environment variable is missing")
	}
	if config.DBPort == "" {
		log.Logger.Fatal("DB_PORT environment variable is missing")
	}
	if config.DBUser == "" {
		log.Logger.Fatal("DB_USER environment variable is missing")
	}
	if config.DBPassword == "" {
		log.Logger.Fatal("DB_PASSWORD environment variable is missing")
	}
	if config.DBName == "" {
		log.Logger.Fatal("DB_NAME environment variable is missing")
	}
	if config.ServerPort == "" {
		log.Logger.Fatal("SERVER_PORT environment variable is missing")
	}
	if config.SyncPort == "" {
		log.Logger.Fatal("SYNC_PORT environment variable is missing")
	}

	if config.Hostname == "" {
		log.Logger.Info("HOSTNAME environment variable is missing")
		h, err := os.Hostname()
		if err != nil {
			log.Logger.Fatal("failed to get hostname", zap.String("error", err.Error()))
		}
		config.Hostname = h
	}
	// Log the loaded configuration
	log.Logger.Info("loaded configuration",
		zap.String("DB_HOST", config.DBHost),
		zap.String("DB_PORT", config.DBPort),
		zap.String("DB_USER", config.DBUser),
		zap.String("DB_NAME", config.DBName),
		zap.String("SERVER_PORT", config.ServerPort),
		zap.String("SYNC_PORT", config.SyncPort),
		zap.String("HOSTNAME", config.Hostname),
	)

	return config
}
