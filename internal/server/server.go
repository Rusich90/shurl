package server

import (
	"database/sql"
	"fmt"

	"github.com/Rusich90/shurl.git/internal/config"
	"github.com/Rusich90/shurl.git/internal/handler"
	"github.com/Rusich90/shurl.git/internal/middleware"
	"github.com/Rusich90/shurl.git/internal/repository"
	"github.com/Rusich90/shurl.git/internal/service"
	"github.com/Rusich90/shurl.git/internal/storage"
	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

func SetupServer(cfg *config.Config) (*gin.Engine, *sql.DB, error) {
	db, err := sql.Open("pgx", cfg.DatabaseDSN)
	if err != nil {
		fmt.Println(err)
		//return nil, nil, fmt.Errorf("failed DB connect: %w", err)
	}

	//ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	//defer cancel()
	//if err = db.PingContext(ctx); err != nil {
	//	panic(err)
	//}

	fileStorage, err := storage.NewFileStorage(cfg.FileStoragePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize file storage: %w", err)
	}

	store, err := repository.NewURLStore(*fileStorage)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize urlStore: %w", err)
	}

	log, err := zap.NewProduction()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create logger: %w", err)
	}
	defer log.Sync()

	urlService := service.NewURLService(store, cfg)
	urlHandler := handler.NewHandler(urlService, cfg, log, db)

	r := gin.New()
	r.Use(middleware.LoggerMiddleware(log))
	r.Use(middleware.GzipMiddleware())

	r.GET("/ping", urlHandler.Ping)

	r.POST("/", urlHandler.CreateShortURL)
	r.GET("/:id", urlHandler.GetOriginalURL)

	api := r.Group("/api")
	{
		api.POST("/shorten", urlHandler.JSONCreateShortURL)
	}

	return r, db, nil
}
