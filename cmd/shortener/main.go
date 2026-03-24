package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/acme/autocert"

	"github.com/Rusich90/shurl.git/internal/audit"
	"github.com/Rusich90/shurl.git/internal/config"
	"github.com/Rusich90/shurl.git/internal/repository"
	"github.com/Rusich90/shurl.git/internal/server"
)

var buildVersion string
var buildDate string
var buildCommit string

func main() {
	cfg := config.InitConfig()

	// Инициализация репозитория URL
	urlRepo, _, err := repository.NewURLRepository(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize URL repository: %v", err)
	}
	defer urlRepo.Close()

	// Инициализация логгера
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	// Инициализация менеджера аудита
	auditManager, err := audit.NewManagerWithConfig(cfg, logger)
	if err != nil {
		log.Fatalf("Failed to initialize audit manager: %v", err)
	}

	// Настройка HTTP-сервера
	r := server.SetupServer(cfg, urlRepo, auditManager, logger)

	printBuildInfo()

	// Создаем HTTP-сервер
	httpServer := &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: r,
	}

	// Создаем контекст для graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

	// WaitGroup для ожидания завершения горутины сервера
	var wg sync.WaitGroup

	// Запуск сервера в отдельной горутине
	wg.Add(1)
	go func() {
		defer wg.Done()
		if cfg.EnableHTTPS {
			log.Printf("Starting HTTPS server on %s\n", cfg.ServerAddress)
			// Используем autocert для автоматического получения сертификатов Let's Encrypt
			m := &autocert.Manager{
				Prompt: autocert.AcceptTOS,
				// Можно добавить HostPolicy для ограничения доменов, если нужно
				// HostPolicy: autocert.HostWhitelist("yourdomain.com"),
			}

			httpServer.TLSConfig = m.TLSConfig()
			if err := httpServer.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
				log.Fatalf("HTTPS Server failed to start: %v\n", err)
			}
		} else {
			log.Printf("Starting server on %s\n", cfg.ServerAddress)
			if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("Server failed to start: %v\n", err)
			}
		}
	}()

	// Ожидание сигнала для завершения
	<-ctx.Done()
	log.Printf("Received shutdown signal, starting graceful shutdown...\n")

	// Создаем контекст с таймаутом для завершения
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Запрашиваем корректное завершение работы сервера
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
		// Принудительное завершение через 5 секунд
		<-time.After(5 * time.Second)
	}

	// Ждем завершения горутины сервера
	wg.Wait()
	log.Println("Server gracefully stopped")

	// Закрываем соединение с БД
	log.Println("Closing database connection...")
	if err := urlRepo.Close(); err != nil {
		log.Printf("Error closing URL repository: %v", err)
	}

	// Закрываем audit manager
	log.Println("Closing audit manager...")
	auditManager.Close()

	log.Println("All resources closed successfully")
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
