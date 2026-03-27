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

// StatsHandler обрабатывает HTTP-запросы для получения статистики сервиса.
type StatsHandler struct {
	statsService *service.StatsService
	logger       *zap.Logger
}

// NewStatsHandler создает новый StatsHandler с указанными зависимостями.
func NewStatsHandler(statsService *service.StatsService, logger *zap.Logger) *StatsHandler {
	return &StatsHandler{
		statsService: statsService,
		logger:       logger,
	}
}

// GetStats обрабатывает GET-запрос для получения статистики сервиса.
//
// Возвращает JSON-объект с количеством URL и пользователей.
func (h *StatsHandler) GetStats(c *gin.Context) {
	if c.Request.Method != http.MethodGet {
		c.AbortWithStatus(http.StatusMethodNotAllowed)
		return
	}

	response, err := h.statsService.GetStats(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get stats", zap.Error(err))
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	respBytes, err := easyjson.Marshal(response)
	if err != nil {
		h.logger.Error("Failed to marshal stats response", zap.Error(err))
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.Data(http.StatusOK, "application/json", respBytes)
}