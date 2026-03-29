// Package server предоставляет функции для настройки и запуска HTTP-сервера.
//
// Содержит функцию SetupServer для настройки маршрутов и middleware
// с использованием переданных зависимостей.
package server

import (
	"github.com/Rusich90/shurl.git/internal/audit"
	"github.com/Rusich90/shurl.git/internal/config"
	domain "github.com/Rusich90/shurl.git/internal/domain/url"
	"github.com/Rusich90/shurl.git/internal/service"
	"github.com/Rusich90/shurl.git/internal/service/auth"
	"github.com/Rusich90/shurl.git/internal/transport/http/handler"
	"github.com/Rusich90/shurl.git/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SetupServer настраивает HTTP-сервер с указанными зависимостями.
//
// Принимает конфигурацию, репозиторий URL, менеджер аудита и логгер.
// Создает сервисы, регистрирует middleware и маршруты.
//
// Пример использования:
//
//	cfg := config.InitConfig()
//	urlRepo, _, err := repository.NewURLRepository(cfg)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer urlRepo.Close()
//
//	log, _ := zap.NewProduction()
//	auditManager, err := audit.NewManagerWithConfig(cfg, log)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer auditManager.Close()
//
//	r := server.SetupServer(cfg, urlRepo, auditManager, log)
//	log.Printf("Starting server on %s", cfg.ServerAddress)
//	r.Run(cfg.ServerAddress)
func SetupServer(cfg *config.Config, urlRepo domain.URLRepository, auditManager *audit.Manager, log *zap.Logger) *gin.Engine {
	authService := auth.NewAuthService(cfg.AuthSecret)

	urlService := service.NewURLService(urlRepo, cfg, log, auditManager)
	healthService := service.NewHealthService(urlRepo)

	urlHandler := handler.NewHandler(urlService, cfg, log)
	healthHandler := handler.NewHealthHandler(healthService, log)

	r := gin.New()
	r.Use(middleware.LoggerMiddleware(log))
	r.Use(middleware.GzipMiddleware())
	r.Use(middleware.AuthMiddleware(authService, log))

	r.GET("/ping", healthHandler.Ping)

	r.POST("/", urlHandler.CreateShortURL)
	r.GET("/:id", urlHandler.GetOriginalURL)

	api := r.Group("/api")
	{
		api.POST("/shorten", urlHandler.JSONCreateShortURL)
		api.POST("/shorten/batch", urlHandler.CreateShortBatchURL)
		api.GET("/user/urls", urlHandler.GetUserOriginalURLs)
		api.DELETE("/user/urls", urlHandler.DeleteURLsByUserID)
	}

	return r
}
