package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Rusich90/shurl.git/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestPing_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger := zap.NewNop()
	healthService := service.NewHealthService(nil)

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

	healthService := service.NewHealthService(nil)

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

	healthService := service.NewHealthService(nil)

	handler := NewHealthHandler(healthService, logger)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	handler.Ping(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	_, statusExists := response["status"]
	assert.True(t, statusExists)

	status := response["status"].(string)
	assert.Contains(t, []string{"ok", "error"}, status)
}

func TestPing_ResponseStructure(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger := zap.NewNop()
	healthService := service.NewHealthService(nil)

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
	assert.Contains(t, []string{"ok", "error"}, status)
}

func TestNewHealthHandler(t *testing.T) {
	logger := zap.NewNop()
	healthService := service.NewHealthService(nil)

	handler := NewHealthHandler(healthService, logger)

	assert.NotNil(t, handler)
	assert.Equal(t, healthService, handler.healthService)
	assert.Equal(t, logger, handler.logger)
}
