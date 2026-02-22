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

type MockURLRepository struct {
	pingError      error
	getError       error
	getResult      domainurl.URL
	getResultOK    bool
	getAllByUserID []domainurl.URL
	getAllByUserIDError error
	deleteURLsError error
	saveIfNotExistsError error
	saveBatchError error
	getByOriginalURL string
	getByOriginalURLOK bool
}

func (m *MockURLRepository) GetUserURLs(ctx context.Context, userID *uuid.UUID) (string, bool) {
	return "", false
}

func (m *MockURLRepository) DeleteURLs(ctx context.Context, IDs []string, userID *uuid.UUID) error {
	if m.deleteURLsError != nil {
		return m.deleteURLsError
	}
	return nil
}

func (m *MockURLRepository) Get(ctx context.Context, id string) (domainurl.URL, bool) {
	if m.getError != nil {
		return domainurl.URL{}, false
	}
	return m.getResult, m.getResultOK
}

func (m *MockURLRepository) GetAllByUserID(ctx context.Context, userID *uuid.UUID) ([]domainurl.URL, error) {
	if m.getAllByUserIDError != nil {
		return nil, m.getAllByUserIDError
	}
	return m.getAllByUserID, nil
}

func (m *MockURLRepository) SaveIfNotExists(ctx context.Context, row domainurl.URL) error {
	if m.saveIfNotExistsError != nil {
		return m.saveIfNotExistsError
	}
	return nil
}

func (m *MockURLRepository) SaveBatch(ctx context.Context, rows []domainurl.URL) error {
	if m.saveBatchError != nil {
		return m.saveBatchError
	}
	return nil
}

func (m *MockURLRepository) GetByOriginalURL(ctx context.Context, originalURL string) (string, bool) {
	return m.getByOriginalURL, m.getByOriginalURLOK
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
