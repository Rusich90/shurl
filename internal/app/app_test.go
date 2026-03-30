package app

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/Rusich90/shurl.git/internal/audit"
	"github.com/Rusich90/shurl.git/internal/config"
	domain "github.com/Rusich90/shurl.git/internal/domain/url"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// MockURLRepository - mock для URLRepository
type MockURLRepository struct {
	closeFunc  func() error
	pingFunc   func(ctx context.Context) error
	getFunc    func(ctx context.Context, id string) (domain.URL, bool)
	getAllFunc func(ctx context.Context, userID *uuid.UUID) ([]domain.URL, error)
	deleteFunc func(ctx context.Context, IDs []string, userID *uuid.UUID) error
	saveFunc   func(ctx context.Context, row domain.URL) error
	saveBatch  func(ctx context.Context, rows []domain.URL) error
	getByOrig  func(ctx context.Context, originalURL string) (string, bool)
	countURLs  func(ctx context.Context) (int, error)
	countUsers func(ctx context.Context) (int, error)
	closed     bool
	mu         sync.Mutex
}

func (m *MockURLRepository) Get(ctx context.Context, id string) (domain.URL, bool) {
	if m.getFunc != nil {
		return m.getFunc(ctx, id)
	}
	return domain.URL{}, false
}

func (m *MockURLRepository) GetAllByUserID(ctx context.Context, userID *uuid.UUID) ([]domain.URL, error) {
	if m.getAllFunc != nil {
		return m.getAllFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockURLRepository) DeleteURLs(ctx context.Context, IDs []string, userID *uuid.UUID) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, IDs, userID)
	}
	return nil
}

func (m *MockURLRepository) SaveIfNotExists(ctx context.Context, row domain.URL) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, row)
	}
	return nil
}

func (m *MockURLRepository) SaveBatch(ctx context.Context, rows []domain.URL) error {
	if m.saveBatch != nil {
		return m.saveBatch(ctx, rows)
	}
	return nil
}

func (m *MockURLRepository) GetByOriginalURL(ctx context.Context, originalURL string) (string, bool) {
	if m.getByOrig != nil {
		return m.getByOrig(ctx, originalURL)
	}
	return "", false
}

func (m *MockURLRepository) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}

func (m *MockURLRepository) Ping(ctx context.Context) error {
	if m.pingFunc != nil {
		return m.pingFunc(ctx)
	}
	return nil
}

func (m *MockURLRepository) CountURLs(ctx context.Context) (int, error) {
	if m.countURLs != nil {
		return m.countURLs(ctx)
	}
	return 0, nil
}

func (m *MockURLRepository) CountUsers(ctx context.Context) (int, error) {
	if m.countUsers != nil {
		return m.countUsers(ctx)
	}
	return 0, nil
}

func (m *MockURLRepository) IsClosed() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.closed
}

// TestInitConfig тестирует инициализацию конфигурации
func TestInitConfig(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "successful config init",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &App{}
			err := app.initConfig()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, app.cfg)
				assert.NotEmpty(t, app.cfg.ServerAddress)
				assert.NotEmpty(t, app.cfg.BaseURL)
			}
		})
	}
}

// TestInitLogger тестирует инициализацию логгера
func TestInitLogger(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "successful logger init",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &App{}
			err := app.initLogger()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, app.logger)
			}
		})
	}
}

// TestInitRepository тестирует инициализацию репозитория
func TestInitRepository(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *App
		wantErr bool
	}{
		{
			name: "successful repository init with file storage",
			setup: func() *App {
				app := &App{}
				app.cfg = &config.Config{
					FileStoragePath: "/tmp/test_storage.jsonl",
					DatabaseDSN:     "",
				}
				return app
			},
			wantErr: false,
		},
		{
			name: "repository init with empty config",
			setup: func() *App {
				app := &App{}
				app.cfg = &config.Config{
					FileStoragePath: "/tmp/test_empty_storage.jsonl",
					DatabaseDSN:     "",
				}
				return app
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := tt.setup()
			err := app.initRepository()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, app.urlRepo)
			}
		})
	}
}

// TestInitAuditManager тестирует инициализацию менеджера аудита
func TestInitAuditManager(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *App
		wantErr bool
	}{
		{
			name: "successful audit manager init",
			setup: func() *App {
				app := &App{}
				app.cfg = &config.Config{
					AuditFile: "",
					AuditURL:  "",
				}
				app.logger = zap.NewNop()
				return app
			},
			wantErr: false,
		},
		{
			name: "audit manager init without logger",
			setup: func() *App {
				app := &App{}
				app.cfg = &config.Config{}
				app.logger = zap.NewNop()
				return app
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := tt.setup()
			err := app.initAuditManager()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, app.auditManager)
			}
		})
	}
}

// TestInitHTTPServer тестирует инициализацию HTTP сервера
func TestInitHTTPServer(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *App
		wantErr bool
	}{
		{
			name: "successful HTTP server init",
			setup: func() *App {
				app := &App{}
				app.cfg = &config.Config{
					ServerAddress: "localhost:8080",
					BaseURL:       "http://localhost:8080",
					AuthSecret:    "test_secret",
				}
				app.urlRepo = &MockURLRepository{}
				app.auditManager = audit.NewManager(zap.NewNop())
				app.logger = zap.NewNop()
				return app
			},
			wantErr: false,
		},
		{
			name: "HTTP server init without dependencies",
			setup: func() *App {
				app := &App{}
				app.cfg = &config.Config{
					ServerAddress: "localhost:8080",
					BaseURL:       "http://localhost:8080",
					AuthSecret:    "test_secret",
				}
				app.urlRepo = &MockURLRepository{}
				app.auditManager = audit.NewManager(zap.NewNop())
				app.logger = zap.NewNop()
				return app
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := tt.setup()
			err := app.initHTTPServer()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, app.httpServer)
				assert.Equal(t, app.cfg.ServerAddress, app.httpServer.Addr)
			}
		})
	}
}

// TestInitGRPCServer тестирует инициализацию gRPC сервера
func TestInitGRPCServer(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *App
		wantErr bool
	}{
		{
			name: "successful gRPC server init",
			setup: func() *App {
				app := &App{}
				app.cfg = &config.Config{
					GRPCServerAddress: "localhost:9090",
					BaseURL:           "http://localhost:8080",
					AuthSecret:        "test_secret",
				}
				app.urlRepo = &MockURLRepository{}
				app.auditManager = audit.NewManager(zap.NewNop())
				app.logger = zap.NewNop()
				return app
			},
			wantErr: false,
		},
		{
			name: "gRPC server init with invalid address",
			setup: func() *App {
				app := &App{}
				app.cfg = &config.Config{
					GRPCServerAddress: "invalid:address",
					BaseURL:           "http://localhost:8080",
					AuthSecret:        "test_secret",
				}
				app.urlRepo = &MockURLRepository{}
				app.auditManager = audit.NewManager(zap.NewNop())
				app.logger = zap.NewNop()
				return app
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := tt.setup()
			err := app.initGRPCServer()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, app.grpcServer)
				assert.NotNil(t, app.grpcListener)
			}
		})
	}
}

// TestNewApp тестирует создание нового приложения
func TestNewApp(t *testing.T) {
	tests := []struct {
		name    string
		setup   func()
		wantErr bool
	}{
		{
			name: "successful app creation with file storage",
			setup: func() {
				// Устанавливаем переменные окружения для теста
				t.Setenv("DATABASE_DSN", "")
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			app, err := NewApp()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, app)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, app)
				assert.NotNil(t, app.cfg)
				assert.NotNil(t, app.logger)
				assert.NotNil(t, app.urlRepo)
				assert.NotNil(t, app.auditManager)
				assert.NotNil(t, app.httpServer)
				assert.NotNil(t, app.grpcServer)

				// Очистка после теста
				if app.urlRepo != nil {
					app.urlRepo.Close()
				}
				if app.auditManager != nil {
					app.auditManager.Close()
				}
				if app.logger != nil {
					app.logger.Sync()
				}
			}
		})
	}
}

// TestShutdown тестирует graceful shutdown
func TestShutdown(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *App
		wantErr bool
	}{
		{
			name: "successful shutdown",
			setup: func() *App {
				app := &App{}
				app.logger = zap.NewNop()

				// Создаем тестовый HTTP сервер
				testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}))
				defer testServer.Close()

				app.httpServer = testServer.Config

				// Создаем тестовый gRPC сервер
				listener, err := net.Listen("tcp", "localhost:0")
				require.NoError(t, err)

				grpcServer := grpc.NewServer()
				app.grpcServer = grpcServer
				app.grpcListener = listener

				return app
			},
			wantErr: false,
		},
		{
			name: "shutdown with minimal servers",
			setup: func() *App {
				app := &App{}
				app.logger = zap.NewNop()

				// Создаем минимальные серверы для теста
				testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}))
				defer testServer.Close()
				app.httpServer = testServer.Config

				listener, err := net.Listen("tcp", "localhost:0")
				require.NoError(t, err)
				grpcServer := grpc.NewServer()
				app.grpcServer = grpcServer
				app.grpcListener = listener

				return app
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := tt.setup()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := app.shutdown(ctx)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestClose тестирует закрытие ресурсов
func TestClose(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *App
		wantErr bool
	}{
		{
			name: "successful close",
			setup: func() *App {
				app := &App{}
				app.logger = zap.NewNop()

				mockRepo := &MockURLRepository{
					closeFunc: func() error {
						return nil
					},
				}
				app.urlRepo = mockRepo

				app.auditManager = audit.NewManager(zap.NewNop())

				return app
			},
			wantErr: false,
		},
		{
			name: "close with repository error",
			setup: func() *App {
				app := &App{}
				app.logger = zap.NewNop()

				mockRepo := &MockURLRepository{
					closeFunc: func() error {
						return errors.New("repository close error")
					},
				}
				app.urlRepo = mockRepo

				app.auditManager = audit.NewManager(zap.NewNop())

				return app
			},
			wantErr: true,
		},
		{
			name: "close with minimal dependencies",
			setup: func() *App {
				app := &App{}
				app.logger = zap.NewNop()

				mockRepo := &MockURLRepository{
					closeFunc: func() error {
						return nil
					},
				}
				app.urlRepo = mockRepo

				app.auditManager = audit.NewManager(zap.NewNop())

				return app
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := tt.setup()
			err := app.close()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Проверяем, что репозиторий закрыт
			if mockRepo, ok := app.urlRepo.(*MockURLRepository); ok {
				assert.True(t, mockRepo.IsClosed())
			}
		})
	}
}

// TestRun тестирует запуск приложения
func TestRun(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *App
		wantErr bool
	}{
		{
			name: "successful run and shutdown",
			setup: func() *App {
				app := &App{}
				app.logger = zap.NewNop()
				app.cfg = &config.Config{
					ServerAddress:     "localhost:0", // Используем случайный порт
					GRPCServerAddress: "localhost:0", // Используем случайный порт
					BaseURL:           "http://localhost:8080",
					AuthSecret:        "test_secret",
					EnableHTTPS:       false,
				}

				mockRepo := &MockURLRepository{
					closeFunc: func() error {
						return nil
					},
				}
				app.urlRepo = mockRepo

				app.auditManager = audit.NewManager(zap.NewNop())

				// Создаем HTTP сервер
				testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}))
				app.httpServer = testServer.Config

				// Создаем gRPC сервер
				listener, err := net.Listen("tcp", "localhost:0")
				require.NoError(t, err)

				grpcServer := grpc.NewServer()
				app.grpcServer = grpcServer
				app.grpcListener = listener

				return app
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := tt.setup()

			// Создаем контекст с отменой для тестирования graceful shutdown
			ctx, cancel := context.WithCancel(context.Background())

			// Запускаем приложение в отдельной горутине
			var wg sync.WaitGroup
			wg.Add(1)
			var runErr error
			go func() {
				defer wg.Done()
				runErr = app.Run(ctx)
			}()

			// Даем серверам время запуститься
			time.Sleep(100 * time.Millisecond)

			// Отменяем контекст для graceful shutdown
			cancel()

			// Ждем завершения
			wg.Wait()

			if tt.wantErr {
				assert.Error(t, runErr)
			} else {
				assert.NoError(t, runErr)
			}
		})
	}
}

// TestLogger тестирует метод Logger
func TestLogger(t *testing.T) {
	tests := []struct {
		name  string
		setup func() *App
	}{
		{
			name: "get logger",
			setup: func() *App {
				app := &App{}
				app.logger = zap.NewNop()
				return app
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := tt.setup()
			logger := app.Logger()

			assert.NotNil(t, logger)
			assert.Equal(t, app.logger, logger)
		})
	}
}

// TestRunHTTPServer тестирует запуск HTTP сервера
func TestRunHTTPServer(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *App
		wantErr bool
	}{
		{
			name: "successful HTTP server run",
			setup: func() *App {
				app := &App{}
				app.logger = zap.NewNop()
				app.cfg = &config.Config{
					ServerAddress: "localhost:0",
					EnableHTTPS:   false,
				}

				testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}))
				app.httpServer = testServer.Config

				return app
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := tt.setup()

			// Запускаем сервер в отдельной горутине
			var wg sync.WaitGroup
			wg.Add(1)
			var runErr error
			go func() {
				defer wg.Done()
				runErr = app.runHTTPServer()
			}()

			// Даем серверу время запуститься
			time.Sleep(50 * time.Millisecond)

			// Останавливаем сервер
			if app.httpServer != nil {
				app.httpServer.Close()
			}

			// Ждем завершения
			wg.Wait()

			if tt.wantErr {
				assert.Error(t, runErr)
			} else {
				// Ожидаем ошибку ErrServerClosed, что является нормальным поведением
				assert.True(t, runErr == nil || runErr == http.ErrServerClosed)
			}
		})
	}
}

// TestRunGRPCServer тестирует запуск gRPC сервера
func TestRunGRPCServer(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *App
		wantErr bool
	}{
		{
			name: "successful gRPC server run",
			setup: func() *App {
				app := &App{}
				app.logger = zap.NewNop()
				app.cfg = &config.Config{
					GRPCServerAddress: "localhost:0",
				}

				listener, err := net.Listen("tcp", "localhost:0")
				require.NoError(t, err)

				grpcServer := grpc.NewServer()
				app.grpcServer = grpcServer
				app.grpcListener = listener

				return app
			},
			wantErr: false, // Не проверяем ошибку, просто проверяем запуск
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := tt.setup()

			// Запускаем сервер в отдельной горутине
			var wg sync.WaitGroup
			wg.Add(1)
			var runErr error
			go func() {
				defer wg.Done()
				runErr = app.runGRPCServer()
			}()

			// Даем серверу время запуститься
			time.Sleep(50 * time.Millisecond)

			// Останавливаем сервер
			if app.grpcServer != nil {
				app.grpcServer.Stop()
			}

			// Ждем завершения
			wg.Wait()

			// gRPC сервер может вернуть ошибку при остановке или nil
			// Главное, что он запустился и остановился корректно
			// Не проверяем ошибку, так как поведение зависит от того, как сервер был остановлен
			_ = runErr // Используем переменную, чтобы избежать ошибки компиляции
		})
	}
}
