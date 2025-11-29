package config

import (
	"flag"
	"os"
)

type Config struct {
	ServerAddress   string
	BaseURL         string
	FileStoragePath string
}

func InitConfig() *Config {
	config := &Config{}

	flag.StringVar(&config.ServerAddress, "a", "localhost:8080", "HTTP server address")
	flag.StringVar(&config.BaseURL, "b", "http://localhost:8080", "Base URL for shortened URLs")
	flag.StringVar(&config.FileStoragePath, "f", "file_storage.jsonl", "Path to file storage")
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

	return config
}
