// Package handler предоставляет HTTP-обработчики (хендлеры) для API-эндпоинтов.
//
// Содержит обработчики для создания, получения и удаления коротких URL.
package handler

import (
	"net/http"

	"github.com/Rusich90/shurl.git/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/mailru/easyjson"
	"go.uber.org/zap"
)

// HealthHandler обрабатывает HTTP-запросы для проверки состояния сервиса.
type HealthHandler struct {
	healthService *service.HealthService
	logger        *zap.Logger
}

// NewHealthHandler создает новый HealthHandler с указанными зависимостями.
func NewHealthHandler(healthService *service.HealthService, logger *zap.Logger) *HealthHandler {
	return &HealthHandler{
		healthService: healthService,
		logger:        logger,
	}
}

// Ping обрабатывает GET-запрос для проверки состояния сервиса.
//
// Возвращает JSON-объект с информацией о статусе и типе хранилища.
func (h *HealthHandler) Ping(c *gin.Context) {
	response := h.healthService.Ping()

	respBytes, err := easyjson.Marshal(response)
	if err != nil {
		h.logger.Error("Failed to marshal ping response", zap.Error(err))
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	statusCode := http.StatusOK
	if response.Status == "error" {
		statusCode = http.StatusInternalServerError
	}

	c.Data(statusCode, "application/json", respBytes)
}
