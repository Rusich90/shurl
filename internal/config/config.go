package config

import (
	"flag"
)

type Config struct {
	ServerAddress string
	BaseURL       string
}

var AppConfig *Config

const (
	Charset  = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	IDLength = 5
)

func InitConfig() *Config {
	config := &Config{}

	flag.StringVar(&config.ServerAddress, "a", "localhost:8080", "HTTP server address")
	flag.StringVar(&config.BaseURL, "b", "http://localhost:8080", "Base URL for shortened URLs")

	flag.Parse()

	return config
}

func GetConfig() *Config {
	if AppConfig == nil {
		AppConfig = &Config{
			ServerAddress: "localhost:8080",
			BaseURL:       "http://localhost:8080",
		}
	}
	return AppConfig
}
