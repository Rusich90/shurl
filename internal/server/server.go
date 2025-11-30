package server

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Rusich90/shurl.git/internal/config"
	"github.com/Rusich90/shurl.git/internal/handler"
	"github.com/Rusich90/shurl.git/internal/middleware"
	"github.com/Rusich90/shurl.git/internal/repository"
	"github.com/Rusich90/shurl.git/internal/service"
	"github.com/Rusich90/shurl.git/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

func SetupServer(cfg *config.Config) (*gin.Engine, *sql.DB, error) {
	var db *sql.DB
	var urlRepo repository.URLRepository

	if cfg.DatabaseDSN != "" {
		var err error
		db, err = sql.Open("pgx", cfg.DatabaseDSN)
		if err != nil {
			return nil, nil, fmt.Errorf("failed DB connect open: %w", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err = db.PingContext(ctx); err != nil {
			return nil, nil, fmt.Errorf("failed DB ping: %w", err)
		}

		if err = runMigrations(db); err != nil {
			return nil, nil, fmt.Errorf("failed to run migrations: %w", err)
		}

		urlRepo = repository.NewDBURLRepository(db)
	} else {
		fileStorage, err := storage.NewFileStorage(cfg.FileStoragePath)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to initialize file storage: %w", err)
		}

		fileRepo, err := repository.NewFileURLRepository(*fileStorage)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to initialize file repository: %w", err)
		}
		urlRepo = fileRepo
	}

	log, err := zap.NewProduction()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create logger: %w", err)
	}
	defer log.Sync()

	urlService := service.NewURLService(urlRepo, cfg)
	healthService := service.NewHealthService(db)

	urlHandler := handler.NewHandler(urlService, cfg, log)
	healthHandler := handler.NewHealthHandler(healthService, log)

	r := gin.New()
	r.Use(middleware.LoggerMiddleware(log))
	r.Use(middleware.GzipMiddleware())

	r.GET("/ping", healthHandler.Ping)

	r.POST("/", urlHandler.CreateShortURL)
	r.GET("/:id", urlHandler.GetOriginalURL)

	api := r.Group("/api")
	{
		api.POST("/shorten", urlHandler.JSONCreateShortURL)
	}

	return r, db, nil
}

func runMigrations(db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migrate driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
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
