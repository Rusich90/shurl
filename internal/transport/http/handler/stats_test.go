package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	domainurl "github.com/Rusich90/shurl.git/internal/domain/url"
	"github.com/Rusich90/shurl.git/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// MockStatsURLRepository - мок-репозиторий для тестирования статистики
type MockStatsURLRepository struct {
	countURLsResult  int
	countURLsError   error
	countUsersResult int
	countUsersError  error
}

func (m *MockStatsURLRepository) Get(ctx context.Context, id string) (domainurl.URL, bool) {
	return domainurl.URL{}, false
}

func (m *MockStatsURLRepository) GetAllByUserID(ctx context.Context, userID *uuid.UUID) ([]domainurl.URL, error) {
	return nil, nil
}

func (m *MockStatsURLRepository) DeleteURLs(ctx context.Context, IDs []string, userID *uuid.UUID) error {
	return nil
}

func (m *MockStatsURLRepository) SaveIfNotExists(ctx context.Context, row domainurl.URL) error {
	return nil
}

func (m *MockStatsURLRepository) SaveBatch(ctx context.Context, rows []domainurl.URL) error {
	return nil
}

func (m *MockStatsURLRepository) GetByOriginalURL(ctx context.Context, originalURL string) (string, bool) {
	return "", false
}

func (m *MockStatsURLRepository) Close() error {
	return nil
}

func (m *MockStatsURLRepository) Ping(ctx context.Context) error {
	return nil
}

func (m *MockStatsURLRepository) CountURLs(ctx context.Context) (int, error) {
	if m.countURLsError != nil {
		return 0, m.countURLsError
	}
	return m.countURLsResult, nil
}

func (m *MockStatsURLRepository) CountUsers(ctx context.Context) (int, error) {
	if m.countUsersError != nil {
		return 0, m.countUsersError
	}
	return m.countUsersResult, nil
}

func TestGetStats_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger := zap.NewNop()
	mockRepo := &MockStatsURLRepository{
		countURLsResult:  100,
		countUsersResult: 50,
	}

	statsService := service.NewStatsService(mockRepo)
	handler := NewStatsHandler(statsService, logger)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/api/stats", nil)

	handler.GetStats(c)

	assert.Equal(t, http.StatusOK, w.Code, "Expected status OK")
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"), "Expected JSON content type")

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err, "Response should be valid JSON")

	urls, exists := response["urls"]
	assert.True(t, exists, "Response should contain 'urls' field")
	assert.Equal(t, float64(100), urls, "URLs count should be 100")

	users, exists := response["users"]
	assert.True(t, exists, "Response should contain 'users' field")
	assert.Equal(t, float64(50), users, "Users count should be 50")
}

func TestGetStats_WrongMethod(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger := zap.NewNop()
	mockRepo := &MockStatsURLRepository{}
	statsService := service.NewStatsService(mockRepo)
	handler := NewStatsHandler(statsService, logger)

	tests := []struct {
		name   string
		method string
	}{
		{
			name:   "POST method",
			method: http.MethodPost,
		},
		{
			name:   "PUT method",
			method: http.MethodPut,
		},
		{
			name:   "DELETE method",
			method: http.MethodDelete,
		},
		{
			name:   "PATCH method",
			method: http.MethodPatch,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest(tt.method, "/api/stats", nil)

			handler.GetStats(c)

			assert.Equal(t, http.StatusMethodNotAllowed, w.Code, "Expected Method Not Allowed")
		})
	}
}

func TestGetStats_CountURLsError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger := zap.NewNop()
	mockRepo := &MockStatsURLRepository{
		countURLsError: fmt.Errorf("database connection error"),
	}

	statsService := service.NewStatsService(mockRepo)
	handler := NewStatsHandler(statsService, logger)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/api/stats", nil)

	handler.GetStats(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code, "Expected Internal Server Error")

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err, "Error response should be valid JSON")

	errorMsg, exists := response["error"]
	assert.True(t, exists, "Error response should contain 'error' field")
	assert.Equal(t, "Internal server error", errorMsg, "Error message should be 'Internal server error'")
}

func TestGetStats_CountUsersError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger := zap.NewNop()
	mockRepo := &MockStatsURLRepository{
		countURLsResult: 100,
		countUsersError: fmt.Errorf("failed to count users"),
	}

	statsService := service.NewStatsService(mockRepo)
	handler := NewStatsHandler(statsService, logger)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/api/stats", nil)

	handler.GetStats(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code, "Expected Internal Server Error")

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err, "Error response should be valid JSON")

	errorMsg, exists := response["error"]
	assert.True(t, exists, "Error response should contain 'error' field")
	assert.Equal(t, "Internal server error", errorMsg, "Error message should be 'Internal server error'")
}

func TestGetStats_EmptyStats(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger := zap.NewNop()
	mockRepo := &MockStatsURLRepository{
		countURLsResult:  0,
		countUsersResult: 0,
	}

	statsService := service.NewStatsService(mockRepo)
	handler := NewStatsHandler(statsService, logger)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/api/stats", nil)

	handler.GetStats(c)

	assert.Equal(t, http.StatusOK, w.Code, "Expected status OK")

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err, "Response should be valid JSON")

	urls, exists := response["urls"]
	assert.True(t, exists, "Response should contain 'urls' field")
	assert.Equal(t, float64(0), urls, "URLs count should be 0")

	users, exists := response["users"]
	assert.True(t, exists, "Response should contain 'users' field")
	assert.Equal(t, float64(0), users, "Users count should be 0")
}

func TestGetStats_LargeNumbers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger := zap.NewNop()
	mockRepo := &MockStatsURLRepository{
		countURLsResult:  1000000,
		countUsersResult: 500000,
	}

	statsService := service.NewStatsService(mockRepo)
	handler := NewStatsHandler(statsService, logger)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/api/stats", nil)

	handler.GetStats(c)

	assert.Equal(t, http.StatusOK, w.Code, "Expected status OK")

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err, "Response should be valid JSON")

	urls, exists := response["urls"]
	assert.True(t, exists, "Response should contain 'urls' field")
	assert.Equal(t, float64(1000000), urls, "URLs count should be 1000000")

	users, exists := response["users"]
	assert.True(t, exists, "Response should contain 'users' field")
	assert.Equal(t, float64(500000), users, "Users count should be 500000")
}

func TestGetStats_ResponseStructure(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger := zap.NewNop()
	mockRepo := &MockStatsURLRepository{
		countURLsResult:  42,
		countUsersResult: 17,
	}

	statsService := service.NewStatsService(mockRepo)
	handler := NewStatsHandler(statsService, logger)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/api/stats", nil)

	handler.GetStats(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Проверяем наличие всех полей
	assert.Contains(t, response, "urls", "Response should contain 'urls' field")
	assert.Contains(t, response, "users", "Response should contain 'users' field")

	// Проверяем типы полей
	assert.IsType(t, float64(0), response["urls"], "URLs should be a number")
	assert.IsType(t, float64(0), response["users"], "Users should be a number")

	// Проверяем отсутствие лишних полей
	assert.Equal(t, 2, len(response), "Response should contain exactly 2 fields")
}

func TestNewStatsHandler(t *testing.T) {
	logger := zap.NewNop()
	mockRepo := &MockStatsURLRepository{}
	statsService := service.NewStatsService(mockRepo)

	handler := NewStatsHandler(statsService, logger)

	assert.NotNil(t, handler, "Handler should not be nil")
	assert.Equal(t, statsService, handler.statsService, "StatsService should be set correctly")
	assert.Equal(t, logger, handler.logger, "Logger should be set correctly")
}

func TestGetStats_ContentType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger := zap.NewNop()
	mockRepo := &MockStatsURLRepository{
		countURLsResult:  10,
		countUsersResult: 5,
	}

	statsService := service.NewStatsService(mockRepo)
	handler := NewStatsHandler(statsService, logger)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/api/stats", nil)

	handler.GetStats(c)

	contentType := w.Header().Get("Content-Type")
	assert.Equal(t, "application/json", contentType, "Content-Type should be application/json")
}

func TestGetStats_WithUsersButNoURLs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger := zap.NewNop()
	mockRepo := &MockStatsURLRepository{
		countURLsResult:  0,
		countUsersResult: 10,
	}

	statsService := service.NewStatsService(mockRepo)
	handler := NewStatsHandler(statsService, logger)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/api/stats", nil)

	handler.GetStats(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	urls, _ := response["urls"]
	users, _ := response["users"]

	assert.Equal(t, float64(0), urls, "URLs count should be 0")
	assert.Equal(t, float64(10), users, "Users count should be 10")
}

func TestGetStats_WithURLsButNoUsers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger := zap.NewNop()
	mockRepo := &MockStatsURLRepository{
		countURLsResult:  25,
		countUsersResult: 0,
	}

	statsService := service.NewStatsService(mockRepo)
	handler := NewStatsHandler(statsService, logger)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/api/stats", nil)

	handler.GetStats(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	urls, _ := response["urls"]
	users, _ := response["users"]

	assert.Equal(t, float64(25), urls, "URLs count should be 25")
	assert.Equal(t, float64(0), users, "Users count should be 0")
}