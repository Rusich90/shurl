package config

import (
	"flag"
)

type Config struct {
	ServerAddress string
	BaseURL       string
}

func InitConfig() *Config {
	config := &Config{}

	flag.StringVar(&config.ServerAddress, "a", "localhost:8080", "HTTP server address")
	flag.StringVar(&config.BaseURL, "b", "http://localhost:8080", "Base URL for shortened URLs")

	flag.Parse()

	return config
}
