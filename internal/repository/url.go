// Package repository предоставляет фабричные функции для создания репозиториев.
//
// Пакет содержит функции для инициализации репозиториев URL в зависимости
// от конфигурации приложения (файловое хранилище или PostgreSQL).
package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Rusich90/shurl.git/internal/config"
	domain "github.com/Rusich90/shurl.git/internal/domain/url"
	"github.com/Rusich90/shurl.git/internal/repository/file"
	postgresrepo "github.com/Rusich90/shurl.git/internal/repository/postgres"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// NewURLRepository создает репозиторий URL на основе конфигурации.
//
// Если указан DatabaseDSN, создается PostgreSQL репозиторий с миграциями.
// Иначе создается файловый репозиторий.
//
// Возвращает репозиторий, соединение с БД (если используется PostgreSQL) и ошибку.
func NewURLRepository(cfg *config.Config) (domain.URLRepository, *sql.DB, error) {
	if cfg.DatabaseDSN != "" {
		return newPostgresRepository(cfg)
	}
	return newFileRepository(cfg)
}

// newPostgresRepository создает PostgreSQL репозиторий с миграциями.
func newPostgresRepository(cfg *config.Config) (domain.URLRepository, *sql.DB, error) {
	db, err := sql.Open("pgx", cfg.DatabaseDSN)
	if err != nil {
		return nil, nil, fmt.Errorf("failed DB connect open: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		return nil, nil, fmt.Errorf("failed DB ping: %w", err)
	}

	if err = runMigrations(db, cfg); err != nil {
		return nil, nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	urlRepo := postgresrepo.NewDBURLRepository(db)
	return urlRepo, db, nil
}

// newFileRepository создает файловый репозиторий.
func newFileRepository(cfg *config.Config) (domain.URLRepository, *sql.DB, error) {
	fileStorage, err := file.NewFileStorage(cfg.FileStoragePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize file storage: %w", err)
	}

	fileRepo, err := file.NewFileURLRepository(*fileStorage)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize file repository: %w", err)
	}

	return fileRepo, nil, nil
}

// runMigrations выполняет миграции базы данных.
//
// Использует golang-migrate для применения всех доступных миграций.
func runMigrations(db *sql.DB, cfg *config.Config) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migrate driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		cfg.MigrationsPath,
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
