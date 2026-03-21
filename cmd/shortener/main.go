package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/crypto/acme/autocert"

	"github.com/Rusich90/shurl.git/internal/config"
	"github.com/Rusich90/shurl.git/internal/server"
)

var buildVersion string
var buildDate string
var buildCommit string

func main() {
	cfg := config.InitConfig()

	r, urlRepo, auditManager, err := server.SetupServer(cfg)
	if err != nil {
		log.Fatalf("Failed to setup server: %v", err)
	}
	defer urlRepo.Close()
	defer auditManager.Close()

	printBuildInfo()

	// Создаем HTTP-сервер
	httpServer := &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: r,
	}

	// Канал для получения сигналов
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	// Запуск сервера в отдельной горутине
	go func() {
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
	sig := <-shutdownChan
	log.Printf("Received signal %v, starting graceful shutdown...\n", sig)

	// Создаем контекст с таймаутом для завершения
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Запрашиваем корректное завершение работы сервера
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
		// Принудительное завершение через 5 секунд
		<-time.After(5 * time.Second)
	}

	log.Println("Server gracefully stopped")
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
