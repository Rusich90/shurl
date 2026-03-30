// Package postgres содержит тесты для PostgreSQL репозитория URL.
package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	domainurl "github.com/Rusich90/shurl.git/internal/domain/url"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	// Загружаем переменные окружения из .env файла в корне проекта
	// Находим корень проекта по файлу go.mod
	projectRoot := findProjectRoot()
	if projectRoot != "" {
		envPath := filepath.Join(projectRoot, ".env")
		if err := godotenv.Load(envPath); err != nil {
			// Если файл не найден, это не критично - переменные могут быть заданы в системе
			fmt.Printf("Warning: .env file not found at %s or error loading: %v\n", envPath, err)
		}
	}
}

// findProjectRoot находит корень проекта, поднимаясь вверх по директориям
// до тех пор, пока не найдет файл go.mod
func findProjectRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}

	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Достигли корня файловой системы
			return ""
		}
		dir = parent
	}
}

// getTestDSN возвращает DSN для тестовой базы данных из переменной окружения.
// Использует только TEST_DATABASE_DSN.
// Если переменная не задана, возвращает пустую строку.
func getTestDSN() string {
	return os.Getenv("TEST_DATABASE_DSN")
}

// setupTestDB создает тестовое подключение к базе данных и очищает таблицу.
// Пропускает тест, если база данных недоступна.
func setupTestDB(t *testing.T) *sql.DB {
	dsn := getTestDSN()
	if dsn == "" {
		t.Skip("Skipping test: TEST_DATABASE_DSN environment variable not set")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Skip("Skipping test: cannot connect to database:", err)
	}

	// Проверяем подключение
	if err := db.Ping(); err != nil {
		t.Skip("Skipping test: database not available:", err)
	}

	// Очищаем таблицу перед тестом
	_, err = db.Exec("DELETE FROM urls")
	if err != nil {
		t.Fatalf("Failed to clean database: %v", err)
	}

	return db
}

// teardownTestDB закрывает подключение к базе данных.
func teardownTestDB(t *testing.T, db *sql.DB) {
	if err := db.Close(); err != nil {
		t.Logf("Warning: failed to close database: %v", err)
	}
}

func TestNewDBURLRepository(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	repo := NewDBURLRepository(db)
	assert.NotNil(t, repo)
	assert.Equal(t, db, repo.db)
}

func TestDBURLRepository_Get(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	repo := NewDBURLRepository(db)
	ctx := context.Background()

	// Тест 1: Получение существующего URL
	userID := uuid.New()
	_, err := db.ExecContext(ctx, "INSERT INTO urls (short_url, original_url, user_id) VALUES ($1, $2, $3)",
		"test123", "https://example.com", userID)
	require.NoError(t, err)

	url, ok := repo.Get(ctx, "test123")
	require.True(t, ok, "URL should be found")
	assert.Equal(t, "test123", url.ShortURL)
	assert.Equal(t, "https://example.com", url.OriginalURL)
	assert.Equal(t, &userID, url.UserID)
	assert.False(t, url.IsDeleted)

	// Тест 2: Получение несуществующего URL
	url, ok = repo.Get(ctx, "nonexistent")
	assert.False(t, ok, "URL should not be found")
	assert.Empty(t, url.ShortURL)

	// Тест 3: Получение удаленного URL
	_, err = db.ExecContext(ctx, "UPDATE urls SET is_deleted = true WHERE short_url = $1", "test123")
	require.NoError(t, err)

	url, ok = repo.Get(ctx, "test123")
	require.True(t, ok, "Deleted URL should still be found")
	assert.True(t, url.IsDeleted)
}

func TestDBURLRepository_GetAllByUserID(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	repo := NewDBURLRepository(db)
	ctx := context.Background()

	// Создаем двух пользователей
	userID1 := uuid.New()
	userID2 := uuid.New()

	// Добавляем URL для пользователей
	_, err := db.ExecContext(ctx, "INSERT INTO urls (short_url, original_url, user_id) VALUES ($1, $2, $3)",
		"user1_url1", "https://example.com/user1/1", userID1)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, "INSERT INTO urls (short_url, original_url, user_id) VALUES ($1, $2, $3)",
		"user1_url2", "https://example.com/user1/2", userID1)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, "INSERT INTO urls (short_url, original_url, user_id) VALUES ($1, $2, $3)",
		"user2_url1", "https://example.com/user2/1", userID2)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, "INSERT INTO urls (short_url, original_url, user_id) VALUES ($1, $2, NULL)",
		"no_user_url", "https://example.com/no_user")
	require.NoError(t, err)

	// Получаем URL для первого пользователя
	urls, err := repo.GetAllByUserID(ctx, &userID1)
	require.NoError(t, err)
	assert.Len(t, urls, 2, "Should return 2 URLs for user1")

	// Проверяем, что все URL принадлежат правильному пользователю
	for _, url := range urls {
		assert.Equal(t, &userID1, url.UserID)
	}

	// Получаем URL для второго пользователя
	urls, err = repo.GetAllByUserID(ctx, &userID2)
	require.NoError(t, err)
	assert.Len(t, urls, 1, "Should return 1 URL for user2")
	assert.Equal(t, "user2_url1", urls[0].ShortURL)

	// Получаем URL для несуществующего пользователя
	nonExistentUserID := uuid.New()
	urls, err = repo.GetAllByUserID(ctx, &nonExistentUserID)
	require.NoError(t, err)
	assert.Len(t, urls, 0, "Should return 0 URLs for non-existent user")
}

func TestDBURLRepository_GetAllByUserID_Deleted(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	repo := NewDBURLRepository(db)
	ctx := context.Background()

	userID := uuid.New()

	// Добавляем URL
	_, err := db.ExecContext(ctx, "INSERT INTO urls (short_url, original_url, user_id) VALUES ($1, $2, $3)",
		"test1", "https://example.com/1", userID)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, "INSERT INTO urls (short_url, original_url, user_id) VALUES ($1, $2, $3)",
		"test2", "https://example.com/2", userID)
	require.NoError(t, err)

	// Помечаем один URL как удаленный
	_, err = db.ExecContext(ctx, "UPDATE urls SET is_deleted = true WHERE short_url = $1", "test1")
	require.NoError(t, err)

	// GetAllByUserID не должна возвращать удаленные URL
	urls, err := repo.GetAllByUserID(ctx, &userID)
	require.NoError(t, err)
	assert.Len(t, urls, 1, "Should return only non-deleted URLs")
	assert.Equal(t, "test2", urls[0].ShortURL)
}

func TestDBURLRepository_SaveIfNotExists(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	repo := NewDBURLRepository(db)
	ctx := context.Background()

	userID := uuid.New()

	// Тест 1: Успешное сохранение нового URL
	row := domainurl.URL{
		ShortURL:    "new123",
		OriginalURL: "https://example.com/new",
		UserID:      &userID,
	}

	err := repo.SaveIfNotExists(ctx, row)
	require.NoError(t, err)

	// Проверяем, что URL сохранен
	var count int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM urls WHERE short_url = $1", "new123").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	// Тест 2: Конфликт по короткому URL
	err = repo.SaveIfNotExists(ctx, row)
	assert.Error(t, err)
	assert.ErrorIs(t, err, domainurl.ErrShortURLConflict)

	// Тест 3: Конфликт по оригинальному URL
	row2 := domainurl.URL{
		ShortURL:    "different",
		OriginalURL: "https://example.com/new",
		UserID:      &userID,
	}

	err = repo.SaveIfNotExists(ctx, row2)
	assert.Error(t, err)
	assert.ErrorIs(t, err, domainurl.ErrOriginalURLConflict)

	// Тест 4: Сохранение URL без пользователя
	row3 := domainurl.URL{
		ShortURL:    "nouser",
		OriginalURL: "https://example.com/nouser",
		UserID:      nil,
	}

	err = repo.SaveIfNotExists(ctx, row3)
	require.NoError(t, err)
}

func TestDBURLRepository_DeleteURLs(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	repo := NewDBURLRepository(db)
	ctx := context.Background()

	userID := uuid.New()
	userID2 := uuid.New()

	// Добавляем URL для разных пользователей
	_, err := db.ExecContext(ctx, "INSERT INTO urls (short_url, original_url, user_id) VALUES ($1, $2, $3)",
		"user1_test1", "https://example.com/1", userID)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, "INSERT INTO urls (short_url, original_url, user_id) VALUES ($1, $2, $3)",
		"user1_test2", "https://example.com/2", userID)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, "INSERT INTO urls (short_url, original_url, user_id) VALUES ($1, $2, $3)",
		"user2_test1", "https://example.com/3", userID2)
	require.NoError(t, err)

	// Удаляем URL первого пользователя
	err = repo.DeleteURLs(ctx, []string{"user1_test1", "user1_test2"}, &userID)
	require.NoError(t, err)

	// Проверяем, что URL первого пользователя помечены как удаленные
	var isDeleted bool
	err = db.QueryRowContext(ctx, "SELECT is_deleted FROM urls WHERE short_url = $1", "user1_test1").Scan(&isDeleted)
	require.NoError(t, err)
	assert.True(t, isDeleted)

	err = db.QueryRowContext(ctx, "SELECT is_deleted FROM urls WHERE short_url = $1", "user1_test2").Scan(&isDeleted)
	require.NoError(t, err)
	assert.True(t, isDeleted)

	// Проверяем, что URL второго пользователя не удален
	err = db.QueryRowContext(ctx, "SELECT is_deleted FROM urls WHERE short_url = $1", "user2_test1").Scan(&isDeleted)
	require.NoError(t, err)
	assert.False(t, isDeleted)

	// Тест: Удаление URL другого пользователя (не должно сработать)
	err = repo.DeleteURLs(ctx, []string{"user2_test1"}, &userID)
	require.NoError(t, err)

	err = db.QueryRowContext(ctx, "SELECT is_deleted FROM urls WHERE short_url = $1", "user2_test1").Scan(&isDeleted)
	require.NoError(t, err)
	assert.False(t, isDeleted, "Should not delete URL of another user")

	// Тест: Удаление несуществующих URL (не должно быть ошибки)
	err = repo.DeleteURLs(ctx, []string{"nonexistent1", "nonexistent2"}, &userID)
	require.NoError(t, err)
}

func TestDBURLRepository_SaveBatch(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	repo := NewDBURLRepository(db)
	ctx := context.Background()

	userID := uuid.New()

	// Тест 1: Успешное сохранение пакета
	rows := []domainurl.URL{
		{
			ShortURL:    "batch1",
			OriginalURL: "https://example.com/batch/1",
			UserID:      &userID,
		},
		{
			ShortURL:    "batch2",
			OriginalURL: "https://example.com/batch/2",
			UserID:      &userID,
		},
		{
			ShortURL:    "batch3",
			OriginalURL: "https://example.com/batch/3",
			UserID:      nil,
		},
	}

	err := repo.SaveBatch(ctx, rows)
	require.NoError(t, err)

	// Проверяем, что все URL сохранены
	for _, row := range rows {
		var count int
		err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM urls WHERE short_url = $1", row.ShortURL).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	}

	// Тест 2: Конфликт в пакете
	conflictRows := []domainurl.URL{
		{
			ShortURL:    "batch1", // Конфликт!
			OriginalURL: "https://example.com/conflict",
			UserID:      &userID,
		},
		{
			ShortURL:    "new1",
			OriginalURL: "https://example.com/new",
			UserID:      &userID,
		},
	}

	err = repo.SaveBatch(ctx, conflictRows)
	assert.Error(t, err, "Should return error on conflict")

	// Проверяем, что ни одна запись не была добавлена
	var count int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM urls WHERE short_url = $1", "new1").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count, "No records should be added on conflict")

	// Тест 3: Пустой пакет
	err = repo.SaveBatch(ctx, []domainurl.URL{})
	require.NoError(t, err)
}

func TestDBURLRepository_GetByOriginalURL(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	repo := NewDBURLRepository(db)
	ctx := context.Background()

	userID := uuid.New()

	// Тест 1: Поиск существующего URL
	_, err := db.ExecContext(ctx, "INSERT INTO urls (short_url, original_url, user_id) VALUES ($1, $2, $3)",
		"test123", "https://example.com", userID)
	require.NoError(t, err)

	shortURL, ok := repo.GetByOriginalURL(ctx, "https://example.com")
	require.True(t, ok, "URL should be found")
	assert.Equal(t, "test123", shortURL)

	// Тест 2: Поиск несуществующего URL
	shortURL, ok = repo.GetByOriginalURL(ctx, "https://nonexistent.com")
	assert.False(t, ok, "URL should not be found")
	assert.Empty(t, shortURL)

	// Тест 3: Поиск удаленного URL (должен быть найден)
	_, err = db.ExecContext(ctx, "UPDATE urls SET is_deleted = true WHERE short_url = $1", "test123")
	require.NoError(t, err)

	shortURL, ok = repo.GetByOriginalURL(ctx, "https://example.com")
	require.True(t, ok, "Deleted URL should still be found")
	assert.Equal(t, "test123", shortURL)
}

func TestDBURLRepository_Close(t *testing.T) {
	db := setupTestDB(t)
	repo := NewDBURLRepository(db)

	err := repo.Close()
	require.NoError(t, err)

	// Проверяем, что соединение закрыто
	err = db.Ping()
	assert.Error(t, err, "Database connection should be closed")
}

func TestDBURLRepository_Ping(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	repo := NewDBURLRepository(db)
	ctx := context.Background()

	err := repo.Ping(ctx)
	require.NoError(t, err, "Should successfully ping database")

	// Тест с отмененным контекстом
	cancelledCtx, cancel := context.WithCancel(ctx)
	cancel()

	err = repo.Ping(cancelledCtx)
	assert.Error(t, err, "Should return error with cancelled context")
}

func TestDBURLRepository_CountURLs(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	repo := NewDBURLRepository(db)
	ctx := context.Background()

	// Тест 1: Пустая база
	count, err := repo.CountURLs(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, count)

	// Тест 2: Добавляем URL
	userID := uuid.New()
	_, err = db.ExecContext(ctx, "INSERT INTO urls (short_url, original_url, user_id) VALUES ($1, $2, $3)",
		"url1", "https://example.com/1", userID)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, "INSERT INTO urls (short_url, original_url, user_id) VALUES ($1, $2, $3)",
		"url2", "https://example.com/2", userID)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, "INSERT INTO urls (short_url, original_url, user_id) VALUES ($1, $2, NULL)",
		"url3", "https://example.com/3")
	require.NoError(t, err)

	count, err = repo.CountURLs(ctx)
	require.NoError(t, err)
	assert.Equal(t, 3, count)

	// Тест 3: Удаленные URL тоже считаются
	_, err = db.ExecContext(ctx, "UPDATE urls SET is_deleted = true WHERE short_url = $1", "url1")
	require.NoError(t, err)

	count, err = repo.CountURLs(ctx)
	require.NoError(t, err)
	assert.Equal(t, 3, count, "Deleted URLs should be counted")
}

func TestDBURLRepository_CountUsers(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	repo := NewDBURLRepository(db)
	ctx := context.Background()

	// Тест 1: Пустая база
	count, err := repo.CountUsers(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, count)

	// Тест 2: Добавляем URL с пользователями
	userID1 := uuid.New()
	userID2 := uuid.New()

	_, err = db.ExecContext(ctx, "INSERT INTO urls (short_url, original_url, user_id) VALUES ($1, $2, $3)",
		"url1", "https://example.com/1", userID1)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, "INSERT INTO urls (short_url, original_url, user_id) VALUES ($1, $2, $3)",
		"url2", "https://example.com/2", userID1)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, "INSERT INTO urls (short_url, original_url, user_id) VALUES ($1, $2, $3)",
		"url3", "https://example.com/3", userID2)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, "INSERT INTO urls (short_url, original_url, user_id) VALUES ($1, $2, NULL)",
		"url4", "https://example.com/4")
	require.NoError(t, err)

	count, err = repo.CountUsers(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, count, "Should count only unique users with non-nil user_id")

	// Тест 3: Удаленные URL пользователя тоже считаются
	_, err = db.ExecContext(ctx, "UPDATE urls SET is_deleted = true WHERE short_url = $1", "url1")
	require.NoError(t, err)

	count, err = repo.CountUsers(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, count, "Deleted URLs should still count users")
}

func TestDBURLRepository_Integration(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	repo := NewDBURLRepository(db)
	ctx := context.Background()

	// Интеграционный тест: полный цикл работы с URL
	userID := uuid.New()

	// 1. Сохраняем URL
	row := domainurl.URL{
		ShortURL:    "integration",
		OriginalURL: "https://integration.example.com",
		UserID:      &userID,
	}

	err := repo.SaveIfNotExists(ctx, row)
	require.NoError(t, err)

	// 2. Получаем URL по короткому идентификатору
	url, ok := repo.Get(ctx, "integration")
	require.True(t, ok)
	assert.Equal(t, "integration", url.ShortURL)
	assert.Equal(t, "https://integration.example.com", url.OriginalURL)

	// 3. Получаем URL по оригинальному URL
	shortURL, ok := repo.GetByOriginalURL(ctx, "https://integration.example.com")
	require.True(t, ok)
	assert.Equal(t, "integration", shortURL)

	// 4. Получаем все URL пользователя
	urls, err := repo.GetAllByUserID(ctx, &userID)
	require.NoError(t, err)
	assert.Len(t, urls, 1)
	assert.Equal(t, "integration", urls[0].ShortURL)

	// 5. Проверяем счетчики
	count, err := repo.CountURLs(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	count, err = repo.CountUsers(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	// 6. Удаляем URL
	err = repo.DeleteURLs(ctx, []string{"integration"}, &userID)
	require.NoError(t, err)

	// 7. Проверяем, что URL помечен как удаленный
	url, ok = repo.Get(ctx, "integration")
	require.True(t, ok)
	assert.True(t, url.IsDeleted)

	// 8. GetAllByUserID не должен возвращать удаленные URL
	urls, err = repo.GetAllByUserID(ctx, &userID)
	require.NoError(t, err)
	assert.Len(t, urls, 0)
}
