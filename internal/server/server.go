package server

import (
	"github.com/Rusich90/shurl.git/internal/config"
	"github.com/Rusich90/shurl.git/internal/handler"
	"github.com/Rusich90/shurl.git/internal/logger"
	"github.com/Rusich90/shurl.git/internal/repository"
	"github.com/Rusich90/shurl.git/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func SetupRouter(cfg *config.Config) *gin.Engine {
	store := repository.NewURLStore()
	urlService := service.NewURLService(store, cfg)
	urlHandler := handler.NewHandler(urlService, cfg)

	log, _ := zap.NewProduction()
	defer log.Sync()

	r := gin.New()
	r.Use(logger.LoggerMiddleware(log))

	r.POST("/", urlHandler.CreateShortURL)
	r.GET("/:id", urlHandler.GetOriginalURL)

	return r
}
