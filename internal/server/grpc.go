// Package server предоставляет функции для настройки и запуска gRPC-сервера.
//
// Содержит функцию SetupGRPCServer для настройки gRPC-сервиса
// с использованием переданных зависимостей.
package server

import (
	"github.com/Rusich90/shurl.git/internal/audit"
	"github.com/Rusich90/shurl.git/internal/config"
	domain "github.com/Rusich90/shurl.git/internal/domain/url"
	"github.com/Rusich90/shurl.git/internal/service"
	"github.com/Rusich90/shurl.git/internal/service/auth"
	grpchandler "github.com/Rusich90/shurl.git/internal/transport/grpc/handler"
	grpcinterceptor "github.com/Rusich90/shurl.git/internal/transport/grpc/interceptor"
	"github.com/Rusich90/shurl.git/pkg"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// SetupGRPCServer настраивает gRPC-сервер с указанными зависимостями.
//
// Принимает конфигурацию, репозиторий URL, менеджер аудита и логгер.
// Создает сервисы и регистрирует gRPC-хендлеры.
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
//	grpcServer := server.SetupGRPCServer(cfg, urlRepo, auditManager, log)
//	lis, _ := net.Listen("tcp", cfg.GRPCServerAddress)
//	log.Printf("Starting gRPC server on %s", cfg.GRPCServerAddress)
//	grpcServer.Serve(lis)
func SetupGRPCServer(cfg *config.Config, urlRepo domain.URLRepository, auditManager *audit.Manager, log *zap.Logger) *grpc.Server {
	urlService := service.NewURLService(urlRepo, cfg, log, auditManager)

	authService := auth.NewAuthService(cfg.AuthSecret)

	urlHandler := grpchandler.NewHandler(urlService, cfg, log)

	// Создаем интерцептор аутентификации
	authInterceptor := grpcinterceptor.AuthInterceptor(authService, log)

	// Создаем gRPC сервер с интерцептором
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(authInterceptor),
	)
	pkg.RegisterShortenerServiceServer(grpcServer, urlHandler)

	// Регистрируем Server Reflection для поддержки Postman и других инструментов
	reflection.Register(grpcServer)

	return grpcServer
}
