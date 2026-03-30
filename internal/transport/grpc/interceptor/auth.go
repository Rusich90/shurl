// Package interceptor предоставляет gRPC-интерцепторы для обработки запросов.
//
// Содержит интерцептор для аутентификации пользователей.
package interceptor

import (
	"context"

	"github.com/Rusich90/shurl.git/internal/service/auth"
	"github.com/Rusich90/shurl.git/internal/transport/grpc/authcontext"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthInterceptor проверяет JWT-токен в метаданных и устанавливает пользователя в контекст.
//
// Если токен отсутствует или невалиден, генерирует новый токен для анонимного пользователя.
func AuthInterceptor(authService *auth.AuthService, logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			authcontext.SetUserID(ctx, nil)
			return handleMissingToken(ctx, authService, logger, handler, req, info)
		}

		tokens := md.Get("authorization")
		if len(tokens) == 0 {
			authcontext.SetUserID(ctx, nil)
			return handleMissingToken(ctx, authService, logger, handler, req, info)
		}

		token := tokens[0]
		claims, err := authService.ValidateToken(token)
		if err != nil {
			authcontext.SetUserID(ctx, nil)
			return handleMissingToken(ctx, authService, logger, handler, req, info)
		}

		ctx = authcontext.SetUserID(ctx, &claims.UserID)
		return handler(ctx, req)
	}
}

func handleMissingToken(ctx context.Context, authService *auth.AuthService, logger *zap.Logger, handler grpc.UnaryHandler, req interface{}, info *grpc.UnaryServerInfo) (interface{}, error) {
	userID := authService.GenerateUserID()

	token, err := authService.GenerateToken(userID)
	if err != nil {
		logger.Error("Failed to generate JWT token", zap.Error(err))
		return nil, status.Error(codes.Internal, "Internal server error")
	}

	// В gRPC мы не можем установить куки, поэтому просто устанавливаем userID в контекст
	// Клиент должен будет использовать возвращенный токен в последующих запросах
	ctx = authcontext.SetUserID(ctx, &userID)

	// Добавляем токен в trailer metadata, чтобы клиент мог его получить
	// Это альтернатива установке куки в HTTP
	trailer := metadata.Pairs("authorization", token)
	grpc.SetTrailer(ctx, trailer)

	return handler(ctx, req)
}
