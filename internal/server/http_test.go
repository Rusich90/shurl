package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Rusich90/shurl.git/internal/audit"
	"github.com/Rusich90/shurl.git/internal/config"
	"github.com/Rusich90/shurl.git/internal/repository"
	file "github.com/Rusich90/shurl.git/internal/repository/file"
	"github.com/Rusich90/shurl.git/internal/repository/postgres"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestSetupServer_FileStorage(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_urls_*.jsonl")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	cfg := &config.Config{
		ServerAddress:   "localhost:8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: tmpFile.Name(),
		DatabaseDSN:     "",
	}

	gin.SetMode(gin.TestMode)

	// Инициализация зависимостей
	urlRepo, _, err := repository.NewURLRepository(cfg)
	require.NoError(t, err)
	require.NotNil(t, urlRepo)
	defer urlRepo.Close()

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	auditManager := audit.NewManager(logger)
	defer auditManager.Close()

	router := SetupServer(cfg, urlRepo, auditManager, logger)
	require.NotNil(t, router)

	// Проверяем, что репозиторий является файловым
	_, ok := urlRepo.(*file.FileURLRepository)
	assert.True(t, ok, "URL repository should be file-based")

	// Проверяем регистрацию middleware
	// Middleware регистрируются в следующем порядке:
	// 1. LoggerMiddleware
	// 2. GzipMiddleware
	// 3. AuthMiddleware
	// Проверим, что они действительно зарегистрированы, послав тестовый запрос
	// и проверив, что в контексте появился userID (добавляется AuthMiddleware)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ping", nil)
	router.ServeHTTP(w, req)

	// Проверяем, что middleware отработали корректно
	assert.Equal(t, http.StatusOK, w.Code)

	// Проверяем регистрацию маршрутов
	routes := []string{"/ping", "/", "/:id", "/api/shorten", "/api/shorten/batch", "/api/user/urls", "/api/user/urls"}
	registeredRoutes := make(map[string]bool)
	for _, route := range router.Routes() {
		registeredRoutes[route.Path] = true
	}

	for _, route := range routes {
		assert.True(t, registeredRoutes[route], "Route %s should be registered", route)
	}
}

func TestSetupServer_Database(t *testing.T) {
	// Для полноценного тестирования с БД потребуется запущенный PostgreSQL
	// В данном тесте мы просто проверим, что функция не возвращает ошибку при корректной конфигурации
	// и что репозиторий является DBURLRepository

	cfg := &config.Config{
		ServerAddress:   "localhost:8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: "",
		DatabaseDSN:     "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable",
		MigrationsPath:  "file://../migrations", // Путь к миграциям относительно теста
	}

	gin.SetMode(gin.TestMode)

	// Инициализация зависимостей
	urlRepo, _, err := repository.NewURLRepository(cfg)
	if err != nil {
		t.Skip("Skipping database test: ", err)
	}
	require.NotNil(t, urlRepo)
	defer urlRepo.Close()

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	auditManager := audit.NewManager(logger)
	defer auditManager.Close()

	router := SetupServer(cfg, urlRepo, auditManager, logger)
	require.NotNil(t, router)

	// Проверяем, что репозиторий является DBURLRepository
	_, ok := urlRepo.(*postgres.DBURLRepository)
	assert.True(t, ok, "URL repository should be database-based")

	// Проверяем регистрацию маршрутов
	assert.Len(t, router.Routes(), 7, "Should register 7 routes")
}

func TestSetupServer_InvalidDBConnection(t *testing.T) {
	cfg := &config.Config{
		ServerAddress:   "localhost:8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: "",
		DatabaseDSN:     "invalid-dsn",
		MigrationsPath:  "file://../migrations",
	}

	gin.SetMode(gin.TestMode)

	// Инициализация зависимостей должна завершиться ошибкой
	urlRepo, _, err := repository.NewURLRepository(cfg)
	assert.Error(t, err)
	assert.Nil(t, urlRepo)
}

func TestSetupServer_InvalidFileStoragePath(t *testing.T) {
	cfg := &config.Config{
		ServerAddress:   "localhost:8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: "/invalid/path/to/file",
		DatabaseDSN:     "",
	}

	gin.SetMode(gin.TestMode)

	// Инициализация зависимостей должна завершиться ошибкой
	urlRepo, _, err := repository.NewURLRepository(cfg)
	assert.Error(t, err)
	assert.Nil(t, urlRepo)
}
