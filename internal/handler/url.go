package handler

import (
	"io"
	"log"
	"net/http"
	"net/url"
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

	id, err := h.urlService.CreateShortURL(originallURL)
	if err != nil {
		log.Printf("Failed to create short URL: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	shortURL, err := url.JoinPath(h.cfg.BaseURL, id)
	if err != nil {
		h.logger.Error("Failed to join URL paths: %v", zap.Error(err))
		return
	}

	c.Status(http.StatusCreated)
	c.String(http.StatusCreated, shortURL)
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

	url, ok := h.urlService.GetOriginalURL(id)
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

	id, err := h.urlService.CreateShortURL(req.URL)
	if err != nil {
		h.logger.Error("Failed to create short URL: %v", zap.Error(err))
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	shortURL, err := url.JoinPath(h.cfg.BaseURL, id)
	if err != nil {
		h.logger.Error("Failed to join URL paths: %v", zap.Error(err))
		return
	}

	response := model.CreateURLResponse{
		Result: shortURL,
	}
	respBytes, err := easyjson.Marshal(response)
	if err != nil {
		h.logger.Error("Failed to marshal response: %v", zap.Error(err))
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.Data(http.StatusCreated, "application/json", respBytes)
}
