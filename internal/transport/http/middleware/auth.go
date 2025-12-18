package middleware

import (
	"net/http"

	"github.com/Rusich90/shurl.git/internal/service/auth"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func AuthMiddleware(authService *auth.AuthService, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenCookie, err := c.Cookie("jwt")

		if err != nil {
			handleMissingToken(c, authService, logger)
			return
		}

		claims, err := authService.ValidateToken(tokenCookie)
		if err != nil {
			c.Set("userID", "")
		} else {
			c.Set("userID", claims.UserID)
		}

		c.Next()
	}
}

func handleMissingToken(c *gin.Context, authService *auth.AuthService, logger *zap.Logger) {
	userID, err := authService.GenerateUserID()
	if err != nil {
		logger.Error("Failed to generate user ID", zap.Error(err))
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	token, err := authService.GenerateToken(userID)
	if err != nil {
		logger.Error("Failed to generate JWT token", zap.Error(err))
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.SetCookie("jwt", token, 24*60*60, "/", "", false, true)

	c.Set("userID", userID)
	c.Next()
}
