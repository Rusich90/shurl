package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"context"

	"github.com/Rusich90/shurl.git/internal/model"
	"github.com/Rusich90/shurl.git/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type MockURLRepository struct {
	pingError error
}

func (m *MockURLRepository) Get(ctx context.Context, id string) (string, bool) {
	return "", false
}

func (m *MockURLRepository) SaveIfNotExists(ctx context.Context, row model.URLRow) error {
	return nil
}

func (m *MockURLRepository) SaveBatch(ctx context.Context, rows []model.URLRow) error {
	return nil
}

func (m *MockURLRepository) GetByOriginalURL(ctx context.Context, originalURL string) (string, bool) {
	return "", false
}

func (m *MockURLRepository) Close() error {
	return nil
}

func (m *MockURLRepository) Ping(ctx context.Context) error {
	return m.pingError
}

func TestPing_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger := zap.NewNop()
	mockRepo := &MockURLRepository{}
	healthService := service.NewHealthService(mockRepo)

	handler := NewHealthHandler(healthService, logger)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	handler.Ping(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	status, exists := response["status"]
	assert.True(t, exists)
	assert.Equal(t, "ok", status)
}

func TestPing_WithDatabase_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger := zap.NewNop()
	mockRepo := &MockURLRepository{}

	healthService := service.NewHealthService(mockRepo)

	handler := NewHealthHandler(healthService, logger)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	handler.Ping(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	status, exists := response["status"]
	assert.True(t, exists)
	assert.Equal(t, "ok", status)
}

func TestPing_ErrorResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger := zap.NewNop()
	mockRepo := &MockURLRepository{pingError: fmt.Errorf("database error")}

	healthService := service.NewHealthService(mockRepo)

	handler := NewHealthHandler(healthService, logger)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	handler.Ping(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	_, statusExists := response["status"]
	assert.True(t, statusExists)

	status := response["status"].(string)
	assert.Equal(t, "error", status)
}

func TestPing_ResponseStructure(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger := zap.NewNop()
	mockRepo := &MockURLRepository{}

	healthService := service.NewHealthService(mockRepo)

	handler := NewHealthHandler(healthService, logger)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	handler.Ping(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Contains(t, response, "status")
	assert.IsType(t, "", response["status"])

	status := response["status"].(string)
	assert.Equal(t, "ok", status)
}

func TestNewHealthHandler(t *testing.T) {
	logger := zap.NewNop()
	mockRepo := &MockURLRepository{}
	healthService := service.NewHealthService(mockRepo)

	handler := NewHealthHandler(healthService, logger)

	assert.NotNil(t, handler)
	assert.Equal(t, healthService, handler.healthService)
	assert.Equal(t, logger, handler.logger)
}
