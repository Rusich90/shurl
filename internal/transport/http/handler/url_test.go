package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/Rusich90/shurl.git/internal/audit"
	"github.com/Rusich90/shurl.git/internal/config"
	domainurl "github.com/Rusich90/shurl.git/internal/domain/url"
	file "github.com/Rusich90/shurl.git/internal/repository/file"
	"github.com/Rusich90/shurl.git/internal/service"
	"github.com/Rusich90/shurl.git/internal/transport/http/authcontext"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type contextKey string

const userIDKey contextKey = "userID"

func TestCreateShortURL(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_urls_*.jsonl")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	cfg := &config.Config{
		ServerAddress:   "localhost:8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: tmpFile.Name(),
		DatabaseDSN:     "",
	}

	logger := zap.NewNop()

	fileStorage, err := file.NewFileStorage(cfg.FileStoragePath)
	if err != nil {
		t.Fatalf("Failed to create file storage: %v", err)
	}

	fileRepo, err := file.NewFileURLRepository(*fileStorage)
	if err != nil {
		t.Fatalf("Failed to create file repository: %v", err)
	}

	auditManager := audit.NewManager(logger)
	urlService := service.NewURLService(fileRepo, cfg, logger, auditManager)
	handler := NewHandler(urlService, cfg, logger)

	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		method         string
		body           string
		expectedStatus int
		checkResponse  bool
		setupFunc      func()
	}{
		{
			name:           "successful creation",
			method:         http.MethodPost,
			body:           "https://example.com",
			expectedStatus: http.StatusCreated,
			checkResponse:  true,
		},
		{
			name:           "successful creation with conflict",
			method:         http.MethodPost,
			body:           "https://example.com/conflict",
			expectedStatus: http.StatusConflict,
			checkResponse:  true,
			setupFunc: func() {
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				var err error
				c.Request, err = http.NewRequest(http.MethodPost, "/", strings.NewReader("https://example.com/conflict"))
				if err != nil {
					t.Fatalf("Failed to create request: %v", err)
				}
				handler.CreateShortURL(c)

				if w.Code != http.StatusCreated {
					t.Fatalf("Expected status %d for setup, got %d", http.StatusCreated, w.Code)
				}
			},
		},
		{
			name:           "wrong method",
			method:         http.MethodGet,
			body:           "",
			expectedStatus: http.StatusMethodNotAllowed,
			checkResponse:  false,
		},
		{
			name:           "empty url",
			method:         http.MethodPost,
			body:           "",
			expectedStatus: http.StatusBadRequest,
			checkResponse:  false,
		},
		{
			name:           "invalid url format",
			method:         http.MethodPost,
			body:           "invalid-url",
			expectedStatus: http.StatusBadRequest,
			checkResponse:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupFunc != nil {
				tt.setupFunc()
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			var err error
			c.Request, err = http.NewRequest(tt.method, "/", strings.NewReader(tt.body))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			handler.CreateShortURL(c)

			if w.Code != tt.expectedStatus {
				t.Fatalf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.checkResponse && (w.Code == http.StatusCreated || w.Code == http.StatusConflict) {
				expectedContentType := "text/plain; charset=utf-8"
				if contentType := w.Header().Get("Content-Type"); contentType != expectedContentType {
					t.Errorf("Expected Content-Type %s, got %s", expectedContentType, contentType)
				}

				responseBody := w.Body.String()
				if responseBody == "" {
					t.Error("Expected non-empty response body")
				}

				if !strings.Contains(responseBody, "http://localhost:8080/") {
					t.Errorf("Expected response to contain short URL, got %s", responseBody)
				}
			}
		})
	}
}

func TestJsonCreateShortURL(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_urls_*.jsonl")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	cfg := &config.Config{
		ServerAddress:   "localhost:8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: tmpFile.Name(),
		DatabaseDSN:     "",
	}

	logger := zap.NewNop()

	fileStorage, err := file.NewFileStorage(cfg.FileStoragePath)
	if err != nil {
		t.Fatalf("Failed to create file storage: %v", err)
	}

	fileRepo, err := file.NewFileURLRepository(*fileStorage)
	if err != nil {
		t.Fatalf("Failed to create file repository: %v", err)
	}

	auditManager := audit.NewManager(logger)
	urlService := service.NewURLService(fileRepo, cfg, logger, auditManager)
	handler := NewHandler(urlService, cfg, logger)

	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		method         string
		body           string
		contentType    string
		expectedStatus int
		checkResponse  bool
		expectedError  string
		setupFunc      func() // Функция для подготовки тестовых данных
	}{
		{
			name:           "successful creation with valid JSON",
			method:         http.MethodPost,
			body:           `{"url":"https://example.com"}`,
			contentType:    "application/json",
			expectedStatus: http.StatusCreated,
			checkResponse:  true,
		},
		{
			name:           "successful creation with conflict",
			method:         http.MethodPost,
			body:           `{"url":"https://example.com/conflict-json"}`,
			contentType:    "application/json",
			expectedStatus: http.StatusConflict,
			checkResponse:  true,
			setupFunc: func() {
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				var err error
				c.Request, err = http.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBufferString(`{"url":"https://example.com/conflict-json"}`))
				if err != nil {
					t.Fatalf("Failed to create request: %v", err)
				}
				c.Request.Header.Set("Content-Type", "application/json")
				handler.JSONCreateShortURL(c)

				if w.Code != http.StatusCreated {
					t.Fatalf("Expected status %d for setup, got %d", http.StatusCreated, w.Code)
				}
			},
		},
		{
			name:           "wrong method",
			method:         http.MethodGet,
			body:           `{"url":"https://example.com"}`,
			contentType:    "application/json",
			expectedStatus: http.StatusMethodNotAllowed,
			checkResponse:  false,
		},
		{
			name:           "invalid JSON format",
			method:         http.MethodPost,
			body:           `{"url":}`,
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
			checkResponse:  false,
			expectedError:  "Invalid JSON format",
		},
		{
			name:           "missing URL field",
			method:         http.MethodPost,
			body:           `{}`,
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
			checkResponse:  false,
			expectedError:  "url is required",
		},
		{
			name:           "empty URL value",
			method:         http.MethodPost,
			body:           `{"url":""}`,
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
			checkResponse:  false,
			expectedError:  "url is required",
		},
		{
			name:           "invalid URL format",
			method:         http.MethodPost,
			body:           `{"url":"invalid-url"}`,
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
			checkResponse:  false,
			expectedError:  "url must use http or https scheme",
		},
		{
			name:           "FTP scheme not allowed",
			method:         http.MethodPost,
			body:           `{"url":"ftp://example.com"}`,
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
			checkResponse:  false,
			expectedError:  "url must use http or https scheme",
		},
		{
			name:           "HTTP scheme allowed",
			method:         http.MethodPost,
			body:           `{"url":"http://example.com"}`,
			contentType:    "application/json",
			expectedStatus: http.StatusCreated,
			checkResponse:  true,
		},
		{
			name:           "HTTPS scheme allowed",
			method:         http.MethodPost,
			body:           `{"url":"https://example.com"}`,
			contentType:    "application/json",
			expectedStatus: http.StatusConflict,
			checkResponse:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupFunc != nil {
				tt.setupFunc()
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			var err error
			c.Request, err = http.NewRequest(tt.method, "/api/shorten", bytes.NewBufferString(tt.body))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			if tt.contentType != "" {
				c.Request.Header.Set("Content-Type", tt.contentType)
			}

			handler.JSONCreateShortURL(c)

			assert.Equal(t, tt.expectedStatus, w.Code, "Status code mismatch")

			if tt.checkResponse && (w.Code == http.StatusCreated || w.Code == http.StatusConflict) {
				expectedContentType := "application/json"
				contentType := w.Header().Get("Content-Type")
				assert.Contains(t, contentType, expectedContentType, "Content-Type should be application/json")

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err, "Response should be valid JSON")

				result, exists := response["result"]
				assert.True(t, exists, "Response should contain 'result' field")
				assert.NotEmpty(t, result, "Result should not be empty")

				resultStr, ok := result.(string)
				assert.True(t, ok, "Result should be a string")
				assert.Contains(t, resultStr, "http://localhost:8080/", "Result should contain base URL")
			}

			if tt.expectedError != "" && w.Code >= 400 {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err, "Error response should be valid JSON")

				errorMsg, exists := response["error"]
				assert.True(t, exists, "Error response should contain 'error' field")

				errorStr, ok := errorMsg.(string)
				assert.True(t, ok, "Error message should be a string")
				assert.Equal(t, tt.expectedError, errorStr, "Error message mismatch")
			}
		})
	}
}

func TestGetOriginalURL(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_urls_*.jsonl")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	cfg := &config.Config{
		ServerAddress:   "localhost:8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: tmpFile.Name(),
		DatabaseDSN:     "",
	}

	logger := zap.NewNop()

	fileStorage, err := file.NewFileStorage(cfg.FileStoragePath)
	if err != nil {
		t.Fatalf("Failed to create file storage: %v", err)
	}

	fileRepo, err := file.NewFileURLRepository(*fileStorage)
	if err != nil {
		t.Fatalf("Failed to create file repository: %v", err)
	}

	auditManager := audit.NewManager(logger)
	urlService := service.NewURLService(fileRepo, cfg, logger, auditManager)
	handler := NewHandler(urlService, cfg, logger)

	w1 := httptest.NewRecorder()
	c1, _ := gin.CreateTestContext(w1)

	c1.Request, err = http.NewRequest(http.MethodPost, "/", strings.NewReader("https://example.com"))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	handler.CreateShortURL(c1)

	shortURL := w1.Body.String()
	id := shortURL[strings.LastIndex(shortURL, "/")+1:]

	gin.SetMode(gin.TestMode)

	tests := []struct {
		name             string
		method           string
		path             string
		param            string
		expectedStatus   int
		expectedLocation string
	}{
		{
			name:             "successful retrieval",
			method:           http.MethodGet,
			path:             "/",
			param:            id,
			expectedStatus:   http.StatusTemporaryRedirect,
			expectedLocation: "https://example.com",
		},
		{
			name:             "wrong method",
			method:           http.MethodPost,
			path:             "/",
			param:            id,
			expectedStatus:   http.StatusMethodNotAllowed,
			expectedLocation: "",
		},
		{
			name:             "empty id",
			method:           http.MethodGet,
			path:             "/",
			param:            "",
			expectedStatus:   http.StatusBadRequest,
			expectedLocation: "",
		},
		{
			name:             "non-existent id",
			method:           http.MethodGet,
			path:             "/",
			param:            "nonexistent",
			expectedStatus:   http.StatusNotFound,
			expectedLocation: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			var err error
			c.Request, err = http.NewRequest(tt.method, tt.path, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			c.AddParam("id", tt.param)

			handler.GetOriginalURL(c)

			if w.Code != tt.expectedStatus {
				t.Fatalf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedLocation != "" {
				location := w.Header().Get("Location")
				if location != tt.expectedLocation {
					t.Errorf("Expected Location header to be %s, got %s", tt.expectedLocation, location)
				}
			}
		})
	}
}

// TestCreateShortBatchURLWithDB тестирует CreateShortBatchURL с использованием мок-репозитория
func TestCreateShortBatchURLWithDB(t *testing.T) {
	// Создаем мок-репозиторий
	mockRepo := &MockURLRepository{
		saveBatchError: nil,
	}

	cfg := &config.Config{
		ServerAddress: "localhost:8080",
		BaseURL:       "http://localhost:8080",
		DatabaseDSN:   "postgres://localhost:5432/testdb",
	}

	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)
	urlService := service.NewURLService(mockRepo, cfg, logger, auditManager)
	handler := NewHandler(urlService, cfg, logger)

	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		method         string
		body           string
		expectedStatus int
		checkResponse  bool
		expectedItems  int
	}{
		{
			name:           "successful batch creation",
			method:         http.MethodPost,
			body:           `[{"correlation_id":"1","original_url":"https://example.com/1"},{"correlation_id":"2","original_url":"https://example.com/2"}]`,
			expectedStatus: http.StatusCreated,
			checkResponse:  true,
			expectedItems:  2,
		},
		{
			name:           "wrong method",
			method:         http.MethodGet,
			body:           `[{"correlation_id":"1","original_url":"https://example.com/1"}]`,
			expectedStatus: http.StatusMethodNotAllowed,
			checkResponse:  false,
			expectedItems:  0,
		},
		{
			name:           "invalid JSON format",
			method:         http.MethodPost,
			body:           `[{"correlation_id":"1","original_url":}]`,
			expectedStatus: http.StatusBadRequest,
			checkResponse:  false,
			expectedItems:  0,
		},
		{
			name:           "missing correlation_id",
			method:         http.MethodPost,
			body:           `[{"original_url":"https://example.com/1"}]`,
			expectedStatus: http.StatusBadRequest,
			checkResponse:  false,
			expectedItems:  0,
		},
		{
			name:           "missing original_url",
			method:         http.MethodPost,
			body:           `[{"correlation_id":"1"}]`,
			expectedStatus: http.StatusBadRequest,
			checkResponse:  false,
			expectedItems:  0,
		},
		{
			name:           "empty original_url",
			method:         http.MethodPost,
			body:           `[{"correlation_id":"1","original_url":""}]`,
			expectedStatus: http.StatusBadRequest,
			checkResponse:  false,
			expectedItems:  0,
		},
		{
			name:           "invalid URL format",
			method:         http.MethodPost,
			body:           `[{"correlation_id":"1","original_url":"invalid-url"}]`,
			expectedStatus: http.StatusBadRequest,
			checkResponse:  false,
			expectedItems:  0,
		},
		{
			name:           "FTP scheme not allowed",
			method:         http.MethodPost,
			body:           `[{"correlation_id":"1","original_url":"ftp://example.com"}]`,
			expectedStatus: http.StatusBadRequest,
			checkResponse:  false,
			expectedItems:  0,
		},
		{
			name:           "HTTP scheme allowed",
			method:         http.MethodPost,
			body:           `[{"correlation_id":"1","original_url":"http://example.com"}]`,
			expectedStatus: http.StatusCreated,
			checkResponse:  true,
			expectedItems:  1,
		},
		{
			name:           "HTTPS scheme allowed",
			method:         http.MethodPost,
			body:           `[{"correlation_id":"1","original_url":"https://example.com"}]`,
			expectedStatus: http.StatusCreated,
			checkResponse:  true,
			expectedItems:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.saveBatchError = nil

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			var err error
			c.Request, err = http.NewRequest(tt.method, "/api/shorten/batch", bytes.NewBufferString(tt.body))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			c.Request.Header.Set("Content-Type", "application/json")

			handler.CreateShortBatchURL(c)

			if w.Code != tt.expectedStatus {
				t.Fatalf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.checkResponse && w.Code == http.StatusCreated {
				expectedContentType := "application/json"
				contentType := w.Header().Get("Content-Type")
				assert.Contains(t, contentType, expectedContentType, "Content-Type should be application/json")

				var response []map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err, "Response should be valid JSON")

				assert.Equal(t, tt.expectedItems, len(response), "Response should contain expected number of items")

				for _, item := range response {
					result, exists := item["short_url"]
					assert.True(t, exists, "Response item should contain 'short_url' field")
					assert.NotEmpty(t, result, "Short URL should not be empty")

					resultStr, ok := result.(string)
					assert.True(t, ok, "Short URL should be a string")
					assert.Contains(t, resultStr, "http://localhost:8080/", "Short URL should contain base URL")

					correlationID, exists := item["correlation_id"]
					assert.True(t, exists, "Response item should contain 'correlation_id' field")
					assert.NotEmpty(t, correlationID, "Correlation ID should not be empty")
				}
			}
		})
	}
}

// TestCreateShortBatchURLWithDB_SaveBatchError тестирует обработку ошибок при сохранении пакета
func TestCreateShortBatchURLWithDB_SaveBatchError(t *testing.T) {
	// Создаем мок-репозиторий
	mockRepo := &MockURLRepository{
		saveBatchError: fmt.Errorf("database error"),
	}

	cfg := &config.Config{
		ServerAddress: "localhost:8080",
		BaseURL:       "http://localhost:8080",
		DatabaseDSN:   "postgres://localhost:5432/testdb",
	}

	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)
	urlService := service.NewURLService(mockRepo, cfg, logger, auditManager)
	handler := NewHandler(urlService, cfg, logger)

	gin.SetMode(gin.TestMode)

	// Тестируем обработку ошибки
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewBufferString(`[{"correlation_id":"1","original_url":"https://example.com/1"}]`))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.CreateShortBatchURL(c)

	// Должен вернуться InternalServerError
	assert.Equal(t, http.StatusInternalServerError, w.Code, "Expected InternalServerError for save batch error")

	// Проверяем, что ответ содержит ошибку
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err, "Response should be valid JSON")

	errorMsg, exists := response["error"]
	assert.True(t, exists, "Error response should contain 'error' field")
	assert.NotEmpty(t, errorMsg, "Error message should not be empty")
}

// TestCreateShortBatchURLWithDB_EmptyBatch тестирует обработку пустого пакета
func TestCreateShortBatchURLWithDB_EmptyBatch(t *testing.T) {
	// Создаем мок-репозиторий
	mockRepo := &MockURLRepository{
		saveBatchError: nil,
	}

	cfg := &config.Config{
		ServerAddress: "localhost:8080",
		BaseURL:       "http://localhost:8080",
		DatabaseDSN:   "postgres://localhost:5432/testdb",
	}

	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)
	urlService := service.NewURLService(mockRepo, cfg, logger, auditManager)
	handler := NewHandler(urlService, cfg, logger)

	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		body           string
		expectedStatus int
	}{
		{
			name:           "empty array",
			body:           `[]`,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "single item",
			body:           `[{"correlation_id":"1","original_url":"https://example.com/1"}]`,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "multiple items",
			body:           `[{"correlation_id":"1","original_url":"https://example.com/1"},{"correlation_id":"2","original_url":"https://example.com/2"},{"correlation_id":"3","original_url":"https://example.com/3"}]`,
			expectedStatus: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.saveBatchError = nil

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewBufferString(tt.body))
			c.Request.Header.Set("Content-Type", "application/json")

			handler.CreateShortBatchURL(c)

			if w.Code != tt.expectedStatus {
				t.Fatalf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

// TestCreateShortBatchURLWithDB_InvalidURLFormat тестирует обработку невалидного URL формата
func TestCreateShortBatchURLWithDB_InvalidURLFormat(t *testing.T) {
	// Создаем мок-репозиторий
	mockRepo := &MockURLRepository{
		saveBatchError: nil,
	}

	cfg := &config.Config{
		ServerAddress: "localhost:8080",
		BaseURL:       "http://localhost:8080",
		DatabaseDSN:   "postgres://localhost:5432/testdb",
	}

	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)
	urlService := service.NewURLService(mockRepo, cfg, logger, auditManager)
	handler := NewHandler(urlService, cfg, logger)

	gin.SetMode(gin.TestMode)

	tests := []struct {
		name string
		body string
	}{
		{
			name: "invalid URL in batch",
			body: `[{"correlation_id":"1","original_url":"invalid-url"}]`,
		},
		{
			name: "FTP scheme not allowed",
			body: `[{"correlation_id":"1","original_url":"ftp://example.com"}]`,
		},
		{
			name: "empty URL",
			body: `[{"correlation_id":"1","original_url":""}]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewBufferString(tt.body))
			c.Request.Header.Set("Content-Type", "application/json")

			handler.CreateShortBatchURL(c)

			// Должен вернуться BadRequest
			assert.Equal(t, http.StatusBadRequest, w.Code, "Expected BadRequest for invalid URL format")
		})
	}
}

// TestCreateShortBatchURLWithDB_MultipleItems тестирует создание нескольких URL
func TestCreateShortBatchURLWithDB_MultipleItems(t *testing.T) {
	// Создаем мок-репозиторий
	mockRepo := &MockURLRepository{
		saveBatchError: nil,
	}

	cfg := &config.Config{
		ServerAddress: "localhost:8080",
		BaseURL:       "http://localhost:8080",
		DatabaseDSN:   "postgres://localhost:5432/testdb",
	}

	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)
	urlService := service.NewURLService(mockRepo, cfg, logger, auditManager)
	handler := NewHandler(urlService, cfg, logger)

	gin.SetMode(gin.TestMode)

	// Тестируем создание 5 URL
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewBufferString(`[
		{"correlation_id":"1","original_url":"https://example.com/1"},
		{"correlation_id":"2","original_url":"https://example.com/2"},
		{"correlation_id":"3","original_url":"https://example.com/3"},
		{"correlation_id":"4","original_url":"https://example.com/4"},
		{"correlation_id":"5","original_url":"https://example.com/5"}
	]`))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.CreateShortBatchURL(c)

	// Должен вернуться StatusCreated
	assert.Equal(t, http.StatusCreated, w.Code, "Expected StatusCreated for successful batch creation")

	// Проверяем, что ответ содержит 5 элементов
	var response []map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err, "Response should be valid JSON")
	assert.Equal(t, 5, len(response), "Response should contain 5 items")

	// Проверяем, что каждый элемент содержит short_url и correlation_id
	for i, item := range response {
		result, exists := item["short_url"]
		assert.True(t, exists, "Response item should contain 'short_url' field")
		assert.NotEmpty(t, result, "Short URL should not be empty")

		resultStr, ok := result.(string)
		assert.True(t, ok, "Short URL should be a string")
		assert.Contains(t, resultStr, "http://localhost:8080/", "Short URL should contain base URL")

		correlationID, exists := item["correlation_id"]
		assert.True(t, exists, "Response item should contain 'correlation_id' field")
		assert.NotEmpty(t, correlationID, "Correlation ID should not be empty")
		assert.Equal(t, fmt.Sprintf("%d", i+1), correlationID, "Correlation ID should match")
	}
}

// TestCreateShortBatchURLWithDB_UserID тестирует передачу user ID
func TestCreateShortBatchURLWithDB_UserID(t *testing.T) {
	// Создаем мок-репозиторий
	mockRepo := &MockURLRepository{
		saveBatchError: nil,
	}

	cfg := &config.Config{
		ServerAddress: "localhost:8080",
		BaseURL:       "http://localhost:8080",
		DatabaseDSN:   "postgres://localhost:5432/testdb",
	}

	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)
	urlService := service.NewURLService(mockRepo, cfg, logger, auditManager)
	handler := NewHandler(urlService, cfg, logger)

	gin.SetMode(gin.TestMode)

	// Создаем тестовый user ID
	userID := uuid.New()

	// Тестируем создание URL с user ID
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewBufferString(`[{"correlation_id":"1","original_url":"https://example.com/1"}]`))
	c.Request.Header.Set("Content-Type", "application/json")

	// Добавляем user ID в контекст
	ctx := context.WithValue(c.Request.Context(), userIDKey, &userID)
	c.Request = c.Request.WithContext(ctx)

	handler.CreateShortBatchURL(c)

	// Должен вернуться StatusCreated
	assert.Equal(t, http.StatusCreated, w.Code, "Expected StatusCreated for successful batch creation")
}

// TestGetUserOriginalURLs тестирует GetUserOriginalURLs с использованием мок-репозитория
func TestGetUserOriginalURLs(t *testing.T) {
	userID := uuid.New()
	mockRepo := &MockURLRepository{
		getAllByUserID: []domainurl.URL{
			{ShortURL: "abc123", OriginalURL: "https://example.com/1", UserID: &userID},
			{ShortURL: "def456", OriginalURL: "https://example.com/2", UserID: &userID},
			{ShortURL: "ghi789", OriginalURL: "https://example.com/3", UserID: &userID},
		},
	}

	cfg := &config.Config{
		ServerAddress: "localhost:8080",
		BaseURL:       "http://localhost:8080",
		DatabaseDSN:   "postgres://localhost:5432/testdb",
	}

	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)
	urlService := service.NewURLService(mockRepo, cfg, logger, auditManager)
	handler := NewHandler(urlService, cfg, logger)

	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		method         string
		expectedStatus int
		expectedItems  int
	}{
		{
			name:           "successful retrieval",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
			expectedItems:  3,
		},
		{
			name:           "wrong method",
			method:         http.MethodPost,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedItems:  0,
		},
		{
			name:           "no content",
			method:         http.MethodGet,
			expectedStatus: http.StatusNoContent,
			expectedItems:  0,
		},
		{
			name:           "repository error",
			method:         http.MethodGet,
			expectedStatus: http.StatusInternalServerError,
			expectedItems:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.getAllByUserID = []domainurl.URL{
				{ShortURL: "abc123", OriginalURL: "https://example.com/1", UserID: &userID},
				{ShortURL: "def456", OriginalURL: "https://example.com/2", UserID: &userID},
				{ShortURL: "ghi789", OriginalURL: "https://example.com/3", UserID: &userID},
			}
			mockRepo.getAllByUserIDError = nil

			if tt.name == "no content" {
				mockRepo.getAllByUserID = []domainurl.URL{}
			}
			if tt.name == "repository error" {
				mockRepo.getAllByUserIDError = fmt.Errorf("database error")
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest(tt.method, "/api/urls", nil)

			authcontext.SetUserID(c, &userID)

			handler.GetUserOriginalURLs(c)

			assert.Equal(t, tt.expectedStatus, w.Code, "Status code mismatch")

			if tt.expectedStatus == http.StatusOK {
				expectedContentType := "application/json"
				contentType := w.Header().Get("Content-Type")
				assert.Contains(t, contentType, expectedContentType, "Content-Type should be application/json")

				var response []map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err, "Response should be valid JSON")
				assert.Equal(t, tt.expectedItems, len(response), "Response should contain expected number of items")

				for _, item := range response {
					shortURL, exists := item["short_url"]
					assert.True(t, exists, "Response item should contain 'short_url' field")
					assert.NotEmpty(t, shortURL, "Short URL should not be empty")

					originalURL, exists := item["original_url"]
					assert.True(t, exists, "Response item should contain 'original_url' field")
					assert.NotEmpty(t, originalURL, "Original URL should not be empty")
				}
			}
		})
	}
}

// TestDeleteURLsByUserID тестирует DeleteURLsByUserID с использованием мок-репозитория
func TestDeleteURLsByUserID(t *testing.T) {
	mockRepo := &MockURLRepository{
		deleteURLsError: nil,
	}

	cfg := &config.Config{
		ServerAddress: "localhost:8080",
		BaseURL:       "http://localhost:8080",
		DatabaseDSN:   "postgres://localhost:5432/testdb",
	}

	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)
	urlService := service.NewURLService(mockRepo, cfg, logger, auditManager)
	handler := NewHandler(urlService, cfg, logger)

	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		method         string
		body           string
		expectedStatus int
	}{
		{
			name:           "successful deletion",
			method:         http.MethodDelete,
			body:           `["abc123","def456"]`,
			expectedStatus: http.StatusAccepted,
		},
		{
			name:           "wrong method",
			method:         http.MethodGet,
			body:           `["abc123"]`,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "invalid JSON format",
			method:         http.MethodDelete,
			body:           `["abc123",]`,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.deleteURLsError = nil

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest(tt.method, "/api/urls", bytes.NewBufferString(tt.body))
			c.Request.Header.Set("Content-Type", "application/json")

			userID := uuid.New()
			authcontext.SetUserID(c, &userID)

			handler.DeleteURLsByUserID(c)

			assert.Equal(t, tt.expectedStatus, w.Code, "Status code mismatch")
		})
	}
}

// TestDeleteURLsByUserID_Unauthorized тестирует обработку отсутствия user ID
func TestDeleteURLsByUserID_Unauthorized(t *testing.T) {
	mockRepo := &MockURLRepository{
		deleteURLsError: nil,
	}

	cfg := &config.Config{
		ServerAddress: "localhost:8080",
		BaseURL:       "http://localhost:8080",
		DatabaseDSN:   "postgres://localhost:5432/testdb",
	}

	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)
	urlService := service.NewURLService(mockRepo, cfg, logger, auditManager)
	handler := NewHandler(urlService, cfg, logger)

	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodDelete, "/api/urls", bytes.NewBufferString(`["abc123"]`))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.DeleteURLsByUserID(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code, "Expected Unauthorized for missing user ID")

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err, "Error response should be valid JSON")

	errorMsg, exists := response["error"]
	assert.True(t, exists, "Error response should contain 'error' field")
	assert.Equal(t, "unauthorized", errorMsg, "Error message should be 'unauthorized'")
}
