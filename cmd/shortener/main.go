package main

import (
	"fmt"
	"log"

	"github.com/Rusich90/shurl.git/internal/config"
	"github.com/Rusich90/shurl.git/internal/server"
)

var buildVersion string
var buildDate string
var buildCommit string

func main() {
	cfg := config.InitConfig()

	r, urlRepo, err := server.SetupServer(cfg)
	if err != nil {
		log.Fatalf("Failed to setup server: %v", err)
	}
	defer urlRepo.Close()

	printBuildInfo()

	log.Printf("Starting server on %s\n", cfg.ServerAddress)
	if err := r.Run(cfg.ServerAddress); err != nil {
		log.Fatalf("Server failed to start: %v\n", err)
	}
}

func printBuildInfo() {
	fmt.Println("Build version:", buildVersionOrDefault(buildVersion))
	fmt.Println("Build date:", buildDateOrDefault(buildDate))
	fmt.Println("Build commit:", buildCommitOrDefault(buildCommit))
}

func buildVersionOrDefault(version string) string {
	if version == "" {
		return "N/A"
	}
	return version
}

func buildDateOrDefault(date string) string {
	if date == "" {
		return "N/A"
	}
	return date
}

func buildCommitOrDefault(commit string) string {
	if commit == "" {
		return "N/A"
	}
	return commit
}
