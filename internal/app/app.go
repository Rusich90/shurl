package app

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/acme/autocert"
	"google.golang.org/grpc"

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
	grpcServer   *grpc.Server
	grpcListener net.Listener
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

	if err := app.initGRPCServer(); err != nil {
		return nil, fmt.Errorf("failed to init gRPC server: %w", err)
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

// initGRPCServer инициализирует gRPC-сервер
func (a *App) initGRPCServer() error {
	a.grpcServer = server.SetupGRPCServer(a.cfg, a.urlRepo, a.auditManager, a.logger)

	listener, err := net.Listen("tcp", a.cfg.GRPCServerAddress)
	if err != nil {
		return fmt.Errorf("failed to create gRPC listener: %w", err)
	}
	a.grpcListener = listener

	return nil
}

// Run запускает приложение
func (a *App) Run(ctx context.Context) error {
	// Канал для передачи ошибок из горутин серверов
	serverErr := make(chan error, 2)

	// WaitGroup для ожидания завершения горутин серверов
	var wg sync.WaitGroup

	// Запуск HTTP сервера в отдельной горутине
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := a.runHTTPServer(); err != nil {
			serverErr <- fmt.Errorf("HTTP server error: %w", err)
		}
	}()

	// Запуск gRPC сервера в отдельной горутине
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := a.runGRPCServer(); err != nil {
			serverErr <- fmt.Errorf("gRPC server error: %w", err)
		}
	}()

	// Ожидание сигнала для завершения или ошибки серверов
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

	// Запрашиваем корректное завершение работы серверов
	if err := a.shutdown(shutdownCtx); err != nil {
		a.logger.Error("Server shutdown error", zap.Error(err))
		return err
	}

	// Ждем завершения горутин серверов
	wg.Wait()
	a.logger.Info("Servers gracefully stopped")

	// Закрываем ресурсы
	if err := a.close(); err != nil {
		a.logger.Error("Error closing resources", zap.Error(err))
		return err
	}

	a.logger.Info("All resources closed successfully")
	return nil
}

// runHTTPServer запускает HTTP/HTTPS сервер
func (a *App) runHTTPServer() error {
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
		a.logger.Info("Starting HTTP server", zap.String("address", a.cfg.ServerAddress))
		if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("HTTP server failed to start: %w", err)
		}
	}
	return nil
}

// runGRPCServer запускает gRPC сервер
func (a *App) runGRPCServer() error {
	a.logger.Info("Starting gRPC server", zap.String("address", a.cfg.GRPCServerAddress))
	if err := a.grpcServer.Serve(a.grpcListener); err != nil {
		return fmt.Errorf("gRPC server failed to start: %w", err)
	}
	return nil
}

// shutdown выполняет graceful shutdown серверов
func (a *App) shutdown(ctx context.Context) error {
	var errs []error

	// Shutdown HTTP сервера
	a.logger.Info("Shutting down HTTP server")
	if err := a.httpServer.Shutdown(ctx); err != nil {
		a.logger.Warn("HTTP server forced to shutdown", zap.Error(err))
		errs = append(errs, fmt.Errorf("HTTP server shutdown error: %w", err))
	}

	// Graceful stop gRPC сервера
	a.logger.Info("Shutting down gRPC server")
	a.grpcServer.GracefulStop()

	if len(errs) > 0 {
		return fmt.Errorf("errors during shutdown: %v", errs)
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
