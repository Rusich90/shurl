package handler

import (
	"context"
	"io"
	"log"
	"net/http"
	neturl "net/url"
	"strings"
	"time"

	"github.com/Rusich90/shurl.git/internal/config"
	"github.com/Rusich90/shurl.git/internal/service"
	"github.com/Rusich90/shurl.git/internal/transport/http/authcontext"
	"github.com/Rusich90/shurl.git/internal/transport/http/dto"
	"github.com/Rusich90/shurl.git/internal/transport/http/validator"
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

	userID, err := authcontext.GetUserID(c)
	if err != nil {
		h.logger.Info("Failed to get user ID: ", zap.Error(err))
	}

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

	if url.IsDeleted {
		c.Status(http.StatusGone)
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, url.OriginalURL)
}

func (h *Handler) GetUserOriginalURLs(c *gin.Context) {
	if c.Request.Method != http.MethodGet {
		c.AbortWithStatus(http.StatusMethodNotAllowed)
		return
	}

	userID, err := authcontext.GetUserID(c)
	if err != nil {
		h.logger.Info("Failed to get user ID: ", zap.Error(err))
	}

	if userID == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	urls, err := h.urlService.GetUserOriginalURLs(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get user original URLs: %v", zap.Error(err))
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "unknown error"})
		return
	}

	if len(urls) == 0 {
		c.AbortWithStatusJSON(http.StatusNoContent, gin.H{})
		return
	}

	var response dto.UserURLsResponse
	for _, domainURL := range urls {
		shortURL, _ := neturl.JoinPath(h.cfg.BaseURL, domainURL.ShortURL) // TODO: Переделать и вынести сборку в сервис
		response = append(response, dto.URLResponse{
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

func (h *Handler) DeleteURLsByUserID(c *gin.Context) {
	if c.Request.Method != http.MethodDelete {
		c.AbortWithStatus(http.StatusMethodNotAllowed)
		return
	}

	userID, err := authcontext.GetUserID(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}
	defer c.Request.Body.Close()

	var req dto.DeleteURLsRequest
	if err := easyjson.Unmarshal(body, &req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	if err := validators.ValidateDeleteURLsRequest(req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c, 1*time.Minute)
	go func() {
		defer cancel()
		if err := h.urlService.DeleteURLsByUserID(ctx, req, userID); err != nil {
			h.logger.Error("Failed to delete URLs", zap.Error(err))
		}
	}()

	c.Status(http.StatusAccepted)
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

	var req dto.CreateURLRequest
	if err := easyjson.Unmarshal(body, &req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	if err := validators.ValidateCreateURLRequest(req.URL); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := authcontext.GetUserID(c)
	if err != nil {
		h.logger.Info("Failed to get user ID: ", zap.Error(err))
	}

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

	response := dto.CreateURLResponse{
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

	var req dto.CreateBatchURLRequest
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

	userID, err := authcontext.GetUserID(c)
	if err != nil {
		h.logger.Info("Failed to get user ID: ", zap.Error(err))
	}

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
