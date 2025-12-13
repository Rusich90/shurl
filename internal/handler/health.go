package handler

import (
	"net/http"

	"github.com/Rusich90/shurl.git/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/mailru/easyjson"
	"go.uber.org/zap"
)

type HealthHandler struct {
	healthService *service.HealthService
	logger        *zap.Logger
}

func NewHealthHandler(healthService *service.HealthService, logger *zap.Logger) *HealthHandler {
	return &HealthHandler{
		healthService: healthService,
		logger:        logger,
	}
}

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
