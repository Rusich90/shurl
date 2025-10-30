package main

import (
	"log"

	"github.com/Rusich90/shurl.git/internal/config"
	"github.com/Rusich90/shurl.git/internal/server"
)

func main() {
	cfg := config.InitConfig()

	r := server.SetupRouter(cfg)

	log.Printf("Starting server on %s\n", cfg.ServerAddress)
	if err := r.Run(cfg.ServerAddress); err != nil {
		log.Fatalf("Server failed to start: %v\n", err)
	}
}
