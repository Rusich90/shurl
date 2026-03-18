package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestLoggerMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Создаем ядро логгера, которое записывает в буфер
	core, observedLogs := observer.New(zap.InfoLevel)
	logger := zap.New(core)

	middleware := LoggerMiddleware(logger)

	// Тестируемые сценарии
	tests := []struct {
		name           string
		method         string
		path           string
		responseBody   string
		expectedStatus int
	}{
		{
			name:           "GET request",
			method:         http.MethodGet,
			path:           "/test/path",
			responseBody:   "test response",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "POST request",
			method:         http.MethodPost,
			path:           "/api/data",
			responseBody:   "created",
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "PUT request",
			method:         http.MethodPut,
			path:           "/api/update",
			responseBody:   "updated",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "DELETE request",
			method:         http.MethodDelete,
			path:           "/api/delete",
			responseBody:   "",
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "Request with error status",
			method:         http.MethodGet,
			path:           "/error",
			responseBody:   "not found",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Очищаем логи перед каждым тестом
			observedLogs.TakeAll()

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest(tt.method, tt.path, nil)
			c.Writer.WriteHeader(tt.expectedStatus)
			c.Writer.Write([]byte(tt.responseBody))

			middleware(c)

			// Получаем логи
			logEntries := observedLogs.All()
			assert.Len(t, logEntries, 1, "Should have exactly one log entry")

			entry := logEntries[0]
			assert.Equal(t, "HTTP request", entry.Message)

			// Проверяем, что в логе есть все необходимые поля
			assert.Contains(t, entry.Context, zap.String("uri", tt.path))
			assert.Contains(t, entry.Context, zap.String("method", tt.method))
			assert.Contains(t, entry.Context, zap.Int("status_code", tt.expectedStatus))
			// response_size может быть -1, если WriteHeader не был вызван до Write
			// или если размер неизвестен (например, при chunked encoding)
			// В тестах WriteHeader вызывается до Write, поэтому response_size будет равен размеру
			assert.Contains(t, entry.Context, zap.Int("response_size", len(tt.responseBody)))
		})
	}
}

func TestLoggerMiddleware_Duration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	core, observedLogs := observer.New(zap.InfoLevel)
	logger := zap.New(core)

	middleware := LoggerMiddleware(logger)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/slow", nil)

	middleware(c)

	// Проверяем, что middleware завершился
	logEntries := observedLogs.All()
	assert.Len(t, logEntries, 1, "Should have exactly one log entry")

	entry := logEntries[0]
	assert.Equal(t, "HTTP request", entry.Message)

	// Проверяем, что duration присутствует в логе (как Duration)
	foundDuration := false
	for _, field := range entry.Context {
		if field.Key == "duration" {
			foundDuration = true
			break
		}
	}
	assert.True(t, foundDuration, "duration field should be present in log")
}

func TestLoggerMiddleware_DifferentPaths(t *testing.T) {
	gin.SetMode(gin.TestMode)

	core, observedLogs := observer.New(zap.InfoLevel)
	logger := zap.New(core)

	middleware := LoggerMiddleware(logger)

	paths := []string{
		"/",
		"/api/v1/users",
		"/api/v1/products/123",
		"/health",
		"/metrics",
	}

	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			observedLogs.TakeAll()

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest(http.MethodGet, path, nil)

			middleware(c)

			logEntries := observedLogs.All()
			assert.Len(t, logEntries, 1, "Should have exactly one log entry for path %s", path)

			entry := logEntries[0]
			assert.Contains(t, entry.Context, zap.String("uri", path))
		})
	}
}

func TestLoggerMiddleware_DifferentMethods(t *testing.T) {
	gin.SetMode(gin.TestMode)

	core, observedLogs := observer.New(zap.InfoLevel)
	logger := zap.New(core)

	middleware := LoggerMiddleware(logger)

	methods := []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodDelete,
		http.MethodPatch,
		http.MethodHead,
		http.MethodOptions,
	}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			observedLogs.TakeAll()

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest(method, "/test", nil)

			middleware(c)

			logEntries := observedLogs.All()
			assert.Len(t, logEntries, 1, "Should have exactly one log entry for method %s", method)

			entry := logEntries[0]
			assert.Contains(t, entry.Context, zap.String("method", method))
		})
	}
}

func TestLoggerMiddleware_StatusCodes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	core, observedLogs := observer.New(zap.InfoLevel)
	logger := zap.New(core)

	middleware := LoggerMiddleware(logger)

	statusCodes := []int{
		http.StatusOK,
		http.StatusCreated,
		http.StatusAccepted,
		http.StatusNoContent,
		http.StatusMovedPermanently,
		http.StatusFound,
		http.StatusBadRequest,
		http.StatusUnauthorized,
		http.StatusForbidden,
		http.StatusNotFound,
		http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
	}

	for _, statusCode := range statusCodes {
		t.Run(http.StatusText(statusCode), func(t *testing.T) {
			observedLogs.TakeAll()

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest(http.MethodGet, "/test", nil)
			c.Writer.WriteHeader(statusCode)

			middleware(c)

			logEntries := observedLogs.All()
			assert.Len(t, logEntries, 1, "Should have exactly one log entry for status %d", statusCode)

			entry := logEntries[0]
			assert.Contains(t, entry.Context, zap.Int("status_code", statusCode))
		})
	}
}

func TestLoggerMiddleware_ResponseSize(t *testing.T) {
	gin.SetMode(gin.TestMode)

	core, observedLogs := observer.New(zap.InfoLevel)
	logger := zap.New(core)

	middleware := LoggerMiddleware(logger)

	tests := []struct {
		name       string
		response   string
		expectedSz int
	}{
		{
			name:       "Empty response",
			response:   "",
			expectedSz: 0,
		},
		{
			name:       "Small response",
			response:   "ok",
			expectedSz: 2,
		},
		{
			name:       "Large response",
			response:   strings.Repeat("x", 1000),
			expectedSz: 1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			observedLogs.TakeAll()

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest(http.MethodGet, "/test", nil)
			c.Writer.WriteHeader(http.StatusOK)
			c.Writer.Write([]byte(tt.response))

			middleware(c)

			logEntries := observedLogs.All()
			assert.Len(t, logEntries, 1, "Should have exactly one log entry")

			entry := logEntries[0]
			// response_size может быть -1, если WriteHeader не был вызван до Write
			// В тестах WriteHeader вызывается до Write, поэтому response_size будет равен размеру
			assert.Contains(t, entry.Context, zap.Int("response_size", len(tt.response)))
		})
	}
}

func TestLoggerMiddleware_MiddlewareChain(t *testing.T) {
	gin.SetMode(gin.TestMode)

	core, observedLogs := observer.New(zap.InfoLevel)
	logger := zap.New(core)

	middleware := LoggerMiddleware(logger)

	// Создаем цепочку middleware
	handler := gin.HandlerFunc(func(c *gin.Context) {
		c.String(http.StatusOK, "handler response")
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/chain", nil)

	// Вызываем middleware и handler
	middleware(c)
	handler(c)

	logEntries := observedLogs.All()
	assert.Len(t, logEntries, 1, "Should have exactly one log entry")

	entry := logEntries[0]
	assert.Equal(t, "HTTP request", entry.Message)
	assert.Contains(t, entry.Context, zap.String("uri", "/chain"))
	assert.Contains(t, entry.Context, zap.String("method", http.MethodGet))
	assert.Contains(t, entry.Context, zap.Int("status_code", http.StatusOK))
	// response_size может быть -1, если WriteHeader не был вызван до Write
	assert.Contains(t, entry.Context, zap.Int("response_size", -1))
}
