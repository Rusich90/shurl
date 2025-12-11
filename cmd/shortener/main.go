package main

import (
	"log"

	"github.com/Rusich90/shurl.git/internal/config"
	"github.com/Rusich90/shurl.git/internal/server"
)

func main() {
	cfg := config.InitConfig()

	r, urlRepo, err := server.SetupServer(cfg)
	if err != nil {
		log.Fatalf("Failed to setup server: %v", err)
	}
	defer urlRepo.Close()

	log.Printf("Starting server on %s\n", cfg.ServerAddress)
	if err := r.Run(cfg.ServerAddress); err != nil {
		log.Fatalf("Server failed to start: %v\n", err)
	}
}
