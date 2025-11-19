package server

import (
	"fmt"

	"github.com/Rusich90/shurl.git/internal/config"
	"github.com/Rusich90/shurl.git/internal/handler"
	"github.com/Rusich90/shurl.git/internal/middleware"
	"github.com/Rusich90/shurl.git/internal/repository"
	"github.com/Rusich90/shurl.git/internal/service"
	"github.com/Rusich90/shurl.git/internal/storage"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func SetupServer(cfg *config.Config) (*gin.Engine, error) {
	fileStorage, err := storage.NewFileStorage(cfg.FileStoragePath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize file storage: %w", err)
	}

	store := repository.NewURLStore(*fileStorage)
	err = store.LoadFromStorage()
	if err != nil {
		return nil, fmt.Errorf("failed to load store: %w", err)
	}

	log, _ := zap.NewProduction()
	defer log.Sync()

	urlService := service.NewURLService(store, cfg)
	urlHandler := handler.NewHandler(urlService, cfg, log)

	r := gin.New()
	r.Use(middleware.LoggerMiddleware(log))
	r.Use(middleware.GzipMiddleware())

	r.POST("/", urlHandler.CreateShortURL)
	r.GET("/:id", urlHandler.GetOriginalURL)

	api := r.Group("/api")
	{
		api.POST("/shorten", urlHandler.JSONCreateShortURL)
	}

	return r, nil
}
