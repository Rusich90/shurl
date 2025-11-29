package config

import (
	"flag"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerAddress   string
	BaseURL         string
	FileStoragePath string
	DatabaseDSN     string
}

func InitConfig() *Config {
	_ = godotenv.Load()

	config := &Config{}

	flag.StringVar(&config.ServerAddress, "a", "localhost:8080", "HTTP server address")
	flag.StringVar(&config.BaseURL, "b", "http://localhost:8080", "Base URL for shortened URLs")
	flag.StringVar(&config.FileStoragePath, "f", "file_storage.jsonl", "Path to file storage")
	flag.StringVar(&config.DatabaseDSN, "d", "postgres://user:password@localhost:5432/shurl", "Database DSN")
	flag.Parse()

	if envServAddr, exists := os.LookupEnv("SERVER_ADDRESS"); exists {
		config.ServerAddress = envServAddr
	}

	if envBaseURL, exists := os.LookupEnv("BASE_URL"); exists {
		config.BaseURL = envBaseURL
	}

	if envFileStoragePath, exists := os.LookupEnv("FILE_STORAGE_PATH"); exists {
		config.FileStoragePath = envFileStoragePath
	}

	if envDatabaseDSN, exists := os.LookupEnv("DATABASE_DSN"); exists {
		config.DatabaseDSN = envDatabaseDSN
	}

	return config
}
