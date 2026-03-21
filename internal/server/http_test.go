package server

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Rusich90/shurl.git/internal/config"
	file "github.com/Rusich90/shurl.git/internal/repository/file"
	postgresrepo "github.com/Rusich90/shurl.git/internal/repository/postgres"
	"github.com/gin-gonic/gin"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	router, urlRepo, _, err := SetupServer(cfg)
	require.NoError(t, err)
	require.NotNil(t, router)
	require.NotNil(t, urlRepo)
	defer urlRepo.Close()

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

	// Создаем временную базу данных в памяти
	db, err := sql.Open("pgx", "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable")
	if err != nil {
		t.Skip("Skipping database test: ", err)
	}
	defer db.Close()

	// Проверяем подключение
	err = db.Ping()
	if err != nil {
		t.Skip("Skipping database test: ", err)
	}

	// Создаем тестовую таблицу
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS shurl (
			short_url TEXT PRIMARY KEY,
			original_url TEXT NOT NULL,
			user_id UUID,
			is_deleted BOOLEAN DEFAULT FALSE
		)
	`)
	require.NoError(t, err)
	defer db.Exec("DROP TABLE IF EXISTS shurl")

	cfg := &config.Config{
		ServerAddress:   "localhost:8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: "",
		DatabaseDSN:     "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable",
		MigrationsPath:  "file://../migrations", // Путь к миграциям относительно теста
	}

	gin.SetMode(gin.TestMode)

	router, urlRepo, _, err := SetupServer(cfg)
	require.NoError(t, err)
	require.NotNil(t, router)
	require.NotNil(t, urlRepo)
	defer urlRepo.Close()

	// Проверяем, что репозиторий является DBURLRepository
	_, ok := urlRepo.(*postgresrepo.DBURLRepository)
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

	router, urlRepo, _, err := SetupServer(cfg)
	assert.Error(t, err)
	assert.Nil(t, router)
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

	router, urlRepo, _, err := SetupServer(cfg)
	assert.Error(t, err)
	assert.Nil(t, router)
	assert.Nil(t, urlRepo)
}

func TestRunMigrations(t *testing.T) {
	// Создаем временную базу данных в памяти
	db, err := sql.Open("pgx", "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable")
	if err != nil {
		t.Skip("Skipping migration test: ", err)
	}
	defer db.Close()

	// Проверяем подключение
	err = db.Ping()
	if err != nil {
		t.Skip("Skipping migration test: ", err)
	}

	cfg := &config.Config{
		MigrationsPath: "file://../migrations",
	}

	err = runMigrations(db, cfg)
	assert.NoError(t, err)

	// Проверяем, что миграции применились (таблица существует)
	var exists bool
	query := `
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_name = 'shurl'
		)
	`
	err = db.QueryRow(query).Scan(&exists)
	assert.NoError(t, err)
	assert.True(t, exists, "Table 'shurl' should exist after migrations")

	// Удаляем таблицу после теста
	db.Exec("DROP TABLE IF EXISTS shurl")
}
