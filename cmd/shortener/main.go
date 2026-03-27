package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/Rusich90/shurl.git/internal/app"
)

var buildVersion string
var buildDate string
var buildCommit string

func main() {
	application, err := app.NewApp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize application: %v\n", err)
		os.Exit(1)
	}

	printBuildInfo()

	// Создаем контекст для graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

	if err := application.Run(ctx); err != nil {
		application.Logger().Error("Application error", zap.Error(err))
		os.Exit(1)
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
