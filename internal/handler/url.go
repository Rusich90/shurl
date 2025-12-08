package handler

import (
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/Rusich90/shurl.git/internal/config"
	"github.com/Rusich90/shurl.git/internal/model"
	"github.com/Rusich90/shurl.git/internal/service"
	validators "github.com/Rusich90/shurl.git/internal/validator"
	"github.com/gin-gonic/gin"
	"github.com/mailru/easyjson"
	"go.uber.org/zap"
)

type Handler struct {
	urlService *service.URLService
	cfg        *config.Config
	logger     *zap.Logger
}

func NewHandler(urlService *service.URLService, cfg *config.Config, logger *zap.Logger) *Handler {
	return &Handler{
		urlService: urlService,
		cfg:        cfg,
		logger:     logger,
	}
}

func (h *Handler) CreateShortURL(c *gin.Context) {
	if c.Request.Method != http.MethodPost {
		c.AbortWithStatus(http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}
	defer c.Request.Body.Close()

	originallURL := strings.TrimSpace(string(body))
	if originallURL == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "URL is required"})
		return
	}

	if !strings.HasPrefix(originallURL, "http://") && !strings.HasPrefix(originallURL, "https://") {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid URL format"})
		return
	}

	result, err := h.urlService.CreateShortURL(c.Request.Context(), originallURL)
	if err != nil {
		log.Printf("Failed to create short URL: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	statusCode := http.StatusCreated
	if !result.IsNew {
		statusCode = http.StatusConflict
	}
	
	c.Status(statusCode)
	c.String(statusCode, result.URL)
}

func (h *Handler) GetOriginalURL(c *gin.Context) {
	if c.Request.Method != http.MethodGet {
		c.AbortWithStatus(http.StatusMethodNotAllowed)
		return
	}

	id := c.Param("id")
	if id == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "ID is required"})
		return
	}

	url, ok := h.urlService.GetOriginalURL(c.Request.Context(), id)
	if !ok {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "URL not found"})
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, url)
}

func (h *Handler) JSONCreateShortURL(c *gin.Context) {
	if c.Request.Method != http.MethodPost {
		c.AbortWithStatus(http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}
	defer c.Request.Body.Close()

	var req model.CreateURLRequest
	if err := easyjson.Unmarshal(body, &req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	if err := validators.ValidateCreateURLRequest(req.URL); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.urlService.CreateShortURL(c.Request.Context(), req.URL)
	if err != nil {
		h.logger.Error("Failed to create short URL: %v", zap.Error(err))
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	statusCode := http.StatusCreated
	if !result.IsNew {
		statusCode = http.StatusConflict
	}

	response := model.CreateURLResponse{
		Result: result.URL,
	}
	respBytes, err := easyjson.Marshal(response)
	if err != nil {
		h.logger.Error("Failed to marshal response: %v", zap.Error(err))
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.Data(statusCode, "application/json", respBytes)
}

func (h *Handler) CreateShortBatchURL(c *gin.Context) {
	if c.Request.Method != http.MethodPost {
		c.AbortWithStatus(http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}
	defer c.Request.Body.Close()

	var req model.CreateBatchURLRequest
	if err := easyjson.Unmarshal(body, &req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	for _, item := range req {
		if err := validators.ValidateCreateBatchURLRequest(item.CorrelationID, item.OriginalURL); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	results, err := h.urlService.CreateShortBatchURL(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to create batch short URLs", zap.Error(err))
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	respBytes, err := easyjson.Marshal(results)
	if err != nil {
		h.logger.Error("Failed to marshal response", zap.Error(err))
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.Data(http.StatusCreated, "application/json", respBytes)
}
