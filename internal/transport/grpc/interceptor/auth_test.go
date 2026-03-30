package interceptor

import (
	"context"
	"errors"
	"testing"

	"github.com/Rusich90/shurl.git/internal/service/auth"
	"github.com/Rusich90/shurl.git/internal/transport/grpc/authcontext"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestAuthInterceptor_MissingMetadata(t *testing.T) {
	authService := auth.NewAuthService("test-secret-key-12345678901234567890")
	logger := zap.NewNop()

	interceptor := AuthInterceptor(authService, logger)

	ctx := context.Background()
	req := "test request"
	info := &grpc.UnaryServerInfo{
		FullMethod: "/test.Service/Method",
	}

	handlerCalled := false
	var capturedUserID *uuid.UUID
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		handlerCalled = true
		capturedUserID, _ = authcontext.GetUserID(ctx)
		return "response", nil
	}

	resp, err := interceptor(ctx, req, info, handler)

	require.NoError(t, err)
	assert.Equal(t, "response", resp)
	assert.True(t, handlerCalled)

	// Проверяем, что userID был установлен в контексте
	assert.NotNil(t, capturedUserID)
}

func TestAuthInterceptor_MissingToken(t *testing.T) {
	authService := auth.NewAuthService("test-secret-key-12345678901234567890")
	logger := zap.NewNop()

	interceptor := AuthInterceptor(authService, logger)

	ctx := metadata.NewIncomingContext(context.Background(), metadata.MD{})
	req := "test request"
	info := &grpc.UnaryServerInfo{
		FullMethod: "/test.Service/Method",
	}

	handlerCalled := false
	var capturedUserID *uuid.UUID
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		handlerCalled = true
		capturedUserID, _ = authcontext.GetUserID(ctx)
		return "response", nil
	}

	resp, err := interceptor(ctx, req, info, handler)

	require.NoError(t, err)
	assert.Equal(t, "response", resp)
	assert.True(t, handlerCalled)

	// Проверяем, что userID был установлен в контексте
	assert.NotNil(t, capturedUserID)
}

func TestAuthInterceptor_ValidToken(t *testing.T) {
	authService := auth.NewAuthService("test-secret-key-12345678901234567890")
	logger := zap.NewNop()

	interceptor := AuthInterceptor(authService, logger)

	// Создаем валидный токен
	userID := uuid.New()
	token, err := authService.GenerateToken(userID)
	require.NoError(t, err)

	ctx := metadata.NewIncomingContext(context.Background(), metadata.MD{
		"authorization": []string{token},
	})
	req := "test request"
	info := &grpc.UnaryServerInfo{
		FullMethod: "/test.Service/Method",
	}

	handlerCalled := false
	var capturedUserID *uuid.UUID
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		handlerCalled = true
		capturedUserID, _ = authcontext.GetUserID(ctx)
		return "response", nil
	}

	resp, err := interceptor(ctx, req, info, handler)

	require.NoError(t, err)
	assert.Equal(t, "response", resp)
	assert.True(t, handlerCalled)
	assert.Equal(t, userID, *capturedUserID)
}

func TestAuthInterceptor_InvalidToken(t *testing.T) {
	authService := auth.NewAuthService("test-secret-key-12345678901234567890")
	logger := zap.NewNop()

	interceptor := AuthInterceptor(authService, logger)

	ctx := metadata.NewIncomingContext(context.Background(), metadata.MD{
		"authorization": []string{"invalid-token"},
	})
	req := "test request"
	info := &grpc.UnaryServerInfo{
		FullMethod: "/test.Service/Method",
	}

	handlerCalled := false
	var capturedUserID *uuid.UUID
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		handlerCalled = true
		capturedUserID, _ = authcontext.GetUserID(ctx)
		return "response", nil
	}

	resp, err := interceptor(ctx, req, info, handler)

	require.NoError(t, err)
	assert.Equal(t, "response", resp)
	assert.True(t, handlerCalled)

	// Проверяем, что userID был установлен в контексте (новый анонимный пользователь)
	assert.NotNil(t, capturedUserID)
}

func TestAuthInterceptor_TokenFromDifferentSecret(t *testing.T) {
	authService1 := auth.NewAuthService("test-secret-key-12345678901234567890")
	authService2 := auth.NewAuthService("different-secret-key-1234567890123456")
	logger := zap.NewNop()

	interceptor := AuthInterceptor(authService2, logger)

	// Создаем токен с другим секретом
	userID := uuid.New()
	token, err := authService1.GenerateToken(userID)
	require.NoError(t, err)

	ctx := metadata.NewIncomingContext(context.Background(), metadata.MD{
		"authorization": []string{token},
	})
	req := "test request"
	info := &grpc.UnaryServerInfo{
		FullMethod: "/test.Service/Method",
	}

	handlerCalled := false
	var capturedUserID *uuid.UUID
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		handlerCalled = true
		capturedUserID, _ = authcontext.GetUserID(ctx)
		return "response", nil
	}

	resp, err := interceptor(ctx, req, info, handler)

	require.NoError(t, err)
	assert.Equal(t, "response", resp)
	assert.True(t, handlerCalled)

	// Проверяем, что userID был установлен в контексте (новый анонимный пользователь)
	assert.NotNil(t, capturedUserID)
	// ID должен отличаться от оригинального, так как токен невалидный
	assert.NotEqual(t, userID, *capturedUserID)
}

func TestAuthInterceptor_MultipleAuthorizationHeaders(t *testing.T) {
	authService := auth.NewAuthService("test-secret-key-12345678901234567890")
	logger := zap.NewNop()

	interceptor := AuthInterceptor(authService, logger)

	// Создаем валидный токен
	userID := uuid.New()
	token, err := authService.GenerateToken(userID)
	require.NoError(t, err)

	ctx := metadata.NewIncomingContext(context.Background(), metadata.MD{
		"authorization": []string{token, "another-token"},
	})
	req := "test request"
	info := &grpc.UnaryServerInfo{
		FullMethod: "/test.Service/Method",
	}

	handlerCalled := false
	var capturedUserID *uuid.UUID
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		handlerCalled = true
		capturedUserID, _ = authcontext.GetUserID(ctx)
		return "response", nil
	}

	resp, err := interceptor(ctx, req, info, handler)

	require.NoError(t, err)
	assert.Equal(t, "response", resp)
	assert.True(t, handlerCalled)
	assert.Equal(t, userID, *capturedUserID)
}

func TestHandleMissingToken_Success(t *testing.T) {
	authService := auth.NewAuthService("test-secret-key-12345678901234567890")
	logger := zap.NewNop()

	ctx := context.Background()
	req := "test request"
	info := &grpc.UnaryServerInfo{
		FullMethod: "/test.Service/Method",
	}

	handlerCalled := false
	var capturedUserID *uuid.UUID
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		handlerCalled = true
		capturedUserID, _ = authcontext.GetUserID(ctx)
		return "response", nil
	}

	resp, err := handleMissingToken(ctx, authService, logger, handler, req, info)

	require.NoError(t, err)
	assert.Equal(t, "response", resp)
	assert.True(t, handlerCalled)
	assert.NotNil(t, capturedUserID)
}

func TestAuthInterceptor_HandlerError(t *testing.T) {
	authService := auth.NewAuthService("test-secret-key-12345678901234567890")
	logger := zap.NewNop()

	interceptor := AuthInterceptor(authService, logger)

	ctx := context.Background()
	req := "test request"
	info := &grpc.UnaryServerInfo{
		FullMethod: "/test.Service/Method",
	}

	expectedError := errors.New("handler error")
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, expectedError
	}

	resp, err := interceptor(ctx, req, info, handler)

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, expectedError, err)
}

func TestAuthInterceptor_HandlerPanic(t *testing.T) {
	authService := auth.NewAuthService("test-secret-key-12345678901234567890")
	logger := zap.NewNop()

	interceptor := AuthInterceptor(authService, logger)

	ctx := context.Background()
	req := "test request"
	info := &grpc.UnaryServerInfo{
		FullMethod: "/test.Service/Method",
	}

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		panic("unexpected panic")
	}

	assert.Panics(t, func() {
		interceptor(ctx, req, info, handler)
	})
}
