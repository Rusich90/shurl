package app

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/acme/autocert"

	"github.com/Rusich90/shurl.git/internal/audit"
	"github.com/Rusich90/shurl.git/internal/config"
	domain "github.com/Rusich90/shurl.git/internal/domain/url"
	"github.com/Rusich90/shurl.git/internal/repository"
	"github.com/Rusich90/shurl.git/internal/server"
)

// App представляет приложение и управляет его жизненным циклом
type App struct {
	cfg          *config.Config
	logger       *zap.Logger
	urlRepo      domain.URLRepository
	auditManager *audit.Manager
	httpServer   *http.Server
}

// NewApp создает новое приложение
func NewApp() (*App, error) {
	app := &App{}

	if err := app.initConfig(); err != nil {
		return nil, fmt.Errorf("failed to init config: %w", err)
	}

	if err := app.initLogger(); err != nil {
		return nil, fmt.Errorf("failed to init logger: %w", err)
	}

	if err := app.initRepository(); err != nil {
		return nil, fmt.Errorf("failed to init repository: %w", err)
	}

	if err := app.initAuditManager(); err != nil {
		return nil, fmt.Errorf("failed to init audit manager: %w", err)
	}

	if err := app.initHTTPServer(); err != nil {
		return nil, fmt.Errorf("failed to init HTTP server: %w", err)
	}

	return app, nil
}

// initConfig инициализирует конфигурацию
func (a *App) initConfig() error {
	a.cfg = config.InitConfig()
	return nil
}

// initLogger инициализирует структурированный логгер
func (a *App) initLogger() error {
	logger, err := zap.NewProduction()
	if err != nil {
		return err
	}
	a.logger = logger
	return nil
}

// initRepository инициализирует репозиторий URL
func (a *App) initRepository() error {
	urlRepo, _, err := repository.NewURLRepository(a.cfg)
	if err != nil {
		return err
	}
	a.urlRepo = urlRepo
	return nil
}

// initAuditManager инициализирует менеджер аудита
func (a *App) initAuditManager() error {
	auditManager, err := audit.NewManagerWithConfig(a.cfg, a.logger)
	if err != nil {
		return err
	}
	a.auditManager = auditManager
	return nil
}

// initHTTPServer инициализирует HTTP-сервер
func (a *App) initHTTPServer() error {
	r := server.SetupServer(a.cfg, a.urlRepo, a.auditManager, a.logger)

	a.httpServer = &http.Server{
		Addr:    a.cfg.ServerAddress,
		Handler: r,
	}
	return nil
}

// Run запускает приложение
func (a *App) Run(ctx context.Context) error {
	// Канал для передачи ошибок из горутины сервера
	serverErr := make(chan error, 1)

	// WaitGroup для ожидания завершения горутины сервера
	var wg sync.WaitGroup

	// Запуск сервера в отдельной горутине
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := a.runServer(); err != nil {
			serverErr <- err
		}
	}()

	// Ожидание сигнала для завершения или ошибки сервера
	select {
	case <-ctx.Done():
		a.logger.Info("Received shutdown signal, starting graceful shutdown")
	case err := <-serverErr:
		a.logger.Error("Server error", zap.Error(err))
		return err
	}

	// Создаем контекст с таймаутом для завершения
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Запрашиваем корректное завершение работы сервера
	if err := a.shutdown(shutdownCtx); err != nil {
		a.logger.Error("Server shutdown error", zap.Error(err))
		return err
	}

	// Ждем завершения горутины сервера
	wg.Wait()
	a.logger.Info("Server gracefully stopped")

	// Закрываем ресурсы
	if err := a.close(); err != nil {
		a.logger.Error("Error closing resources", zap.Error(err))
		return err
	}

	a.logger.Info("All resources closed successfully")
	return nil
}

// runServer запускает HTTP/HTTPS сервер
func (a *App) runServer() error {
	if a.cfg.EnableHTTPS {
		a.logger.Info("Starting HTTPS server", zap.String("address", a.cfg.ServerAddress))
		// Используем autocert для автоматического получения сертификатов Let's Encrypt
		m := &autocert.Manager{
			Prompt: autocert.AcceptTOS,
		}

		a.httpServer.TLSConfig = m.TLSConfig()
		if err := a.httpServer.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("HTTPS server failed to start: %w", err)
		}
	} else {
		a.logger.Info("Starting server", zap.String("address", a.cfg.ServerAddress))
		if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("server failed to start: %w", err)
		}
	}
	return nil
}

// shutdown выполняет graceful shutdown сервера
func (a *App) shutdown(ctx context.Context) error {
	if err := a.httpServer.Shutdown(ctx); err != nil {
		a.logger.Warn("Server forced to shutdown", zap.Error(err))
		// Принудительное завершение через 5 секунд
		<-time.After(5 * time.Second)
		return fmt.Errorf("server forced to shutdown: %w", err)
	}
	return nil
}

// Logger возвращает логгер приложения
func (a *App) Logger() *zap.Logger {
	return a.logger
}

// close закрывает все ресурсы приложения
func (a *App) close() error {
	var errs []error

	// Закрываем соединение с БД
	a.logger.Info("Closing database connection")
	if err := a.urlRepo.Close(); err != nil {
		errs = append(errs, fmt.Errorf("error closing URL repository: %w", err))
	}

	// Закрываем audit manager
	a.logger.Info("Closing audit manager")
	a.auditManager.Close()

	// Синхронизируем логгер
	a.logger.Sync()

	if len(errs) > 0 {
		return fmt.Errorf("errors during close: %v", errs)
	}
	return nil
}
