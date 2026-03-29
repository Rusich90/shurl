// Package repository содержит тесты для фабричных функций репозиториев.
package repository

import (
	"database/sql"
	"os"
	"testing"

	"github.com/Rusich90/shurl.git/internal/config"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewURLRepository_FileStorage(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_urls_*.jsonl")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	cfg := &config.Config{
		FileStoragePath: tmpFile.Name(),
		DatabaseDSN:     "",
	}

	urlRepo, db, err := NewURLRepository(cfg)
	require.NoError(t, err)
	require.NotNil(t, urlRepo)
	assert.Nil(t, db, "DB should be nil for file storage")
	defer urlRepo.Close()
}

func TestNewURLRepository_Database(t *testing.T) {
	cfg := &config.Config{
		DatabaseDSN:    "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable",
		MigrationsPath: "file://../migrations",
	}

	urlRepo, db, err := NewURLRepository(cfg)
	if err != nil {
		t.Skip("Skipping database test: ", err)
	}
	require.NotNil(t, urlRepo)
	require.NotNil(t, db, "DB should not be nil for database storage")
	defer urlRepo.Close()
}

func TestNewURLRepository_InvalidDBConnection(t *testing.T) {
	cfg := &config.Config{
		DatabaseDSN:    "invalid-dsn",
		MigrationsPath: "file://../migrations",
	}

	urlRepo, db, err := NewURLRepository(cfg)
	assert.Error(t, err)
	assert.Nil(t, urlRepo)
	assert.Nil(t, db)
}

func TestNewURLRepository_InvalidFileStoragePath(t *testing.T) {
	cfg := &config.Config{
		FileStoragePath: "/invalid/path/to/file",
		DatabaseDSN:     "",
	}

	urlRepo, db, err := NewURLRepository(cfg)
	assert.Error(t, err)
	assert.Nil(t, urlRepo)
	assert.Nil(t, db)
}

func TestNewURLRepository_EmptyConfig(t *testing.T) {
	cfg := &config.Config{
		FileStoragePath: "",
		DatabaseDSN:     "",
	}

	urlRepo, db, err := NewURLRepository(cfg)
	assert.Error(t, err)
	assert.Nil(t, urlRepo)
	assert.Nil(t, db)
}

func TestNewURLRepository_DatabasePingError(t *testing.T) {
	cfg := &config.Config{
		DatabaseDSN:    "postgres://invalid:invalid@invalid-host:9999/invalid?sslmode=disable",
		MigrationsPath: "file://../migrations",
	}

	urlRepo, db, err := NewURLRepository(cfg)
	assert.Error(t, err)
	assert.Nil(t, urlRepo)
	assert.Nil(t, db)
}

func TestRunMigrations_InvalidDB(t *testing.T) {
	db, err := sql.Open("pgx", "invalid-dsn")
	require.NoError(t, err)
	defer db.Close()

	cfg := &config.Config{
		MigrationsPath: "file://../migrations",
	}

	err = runMigrations(db, cfg)
	assert.Error(t, err)
}

func TestRunMigrations_InvalidMigrationsPath(t *testing.T) {
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
		MigrationsPath: "file://invalid/path",
	}

	err = runMigrations(db, cfg)
	assert.Error(t, err)
}
