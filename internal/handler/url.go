package handler

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/Rusich90/shurl.git/internal/config"
	"github.com/Rusich90/shurl.git/internal/httpctx"
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

	originalURL := strings.TrimSpace(string(body))
	if originalURL == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "URL is required"})
		return
	}

	if !strings.HasPrefix(originalURL, "http://") && !strings.HasPrefix(originalURL, "https://") {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid URL format"})
		return
	}

	userID, _ := httpctx.GetUserID(c)
	result, err := h.urlService.CreateShortURL(c.Request.Context(), originalURL, userID)
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

func (h *Handler) GetUserOriginalURLs(c *gin.Context) {
	if c.Request.Method != http.MethodGet {
		c.AbortWithStatus(http.StatusMethodNotAllowed)
		return
	}

	userID, _ := httpctx.GetUserID(c)
	if userID == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	urls, err := h.urlService.GetUserOriginalURLs(c.Request.Context(), userID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "URL not found"})
		return
	}

	if len(urls) == 0 {
		c.AbortWithStatusJSON(http.StatusNoContent, gin.H{})
		return
	}

	var response model.UserURLsResponse
	for _, domainURL := range urls {
		shortURL, _ := url.JoinPath(h.cfg.BaseURL, domainURL.ShortURL) // TODO: Переделать и вынести сборку в сервис
		response = append(response, model.URLResponse{
			ShortURL:    shortURL,
			OriginalURL: domainURL.OriginalURL,
		})
	}

	respBytes, err := easyjson.Marshal(response)
	if err != nil {
		h.logger.Error("Failed to marshal response", zap.Error(err))
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.Data(http.StatusOK, "application/json", respBytes)

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

	userID, _ := httpctx.GetUserID(c)
	result, err := h.urlService.CreateShortURL(c.Request.Context(), req.URL, userID)
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

	userID, _ := httpctx.GetUserID(c)
	results, err := h.urlService.CreateShortBatchURL(c.Request.Context(), req, userID)
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
