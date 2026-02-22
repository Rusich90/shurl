// Package middleware предоставляет функции-обработчики Gin middleware.
//
// Содержит middleware для логирования, сжатия и аутентификации запросов.
package middleware

import (
	"net/http"

	"github.com/Rusich90/shurl.git/internal/service/auth"
	"github.com/Rusich90/shurl.git/internal/transport/http/authcontext"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AuthMiddleware проверяет JWT-токен в куки и устанавливает пользователя в контекст.
//
// Если токен отсутствует или невалиден, генерирует новый токен для анонимного пользователя.
func AuthMiddleware(authService *auth.AuthService, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenCookie, err := c.Cookie("jwt")

		if err != nil {
			authcontext.SetUserID(c, nil)
			handleMissingToken(c, authService, logger)
			return
		}

		claims, err := authService.ValidateToken(tokenCookie)
		if err != nil {
			authcontext.SetUserID(c, nil)
		} else {
			authcontext.SetUserID(c, &claims.UserID)
		}

		c.Next()
	}
}

func handleMissingToken(c *gin.Context, authService *auth.AuthService, logger *zap.Logger) {
	userID := authService.GenerateUserID()

	token, err := authService.GenerateToken(userID)
	if err != nil {
		logger.Error("Failed to generate JWT token", zap.Error(err))
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.SetCookie("jwt", token, 24*60*60, "/", "", false, true)

	authcontext.SetUserID(c, &userID)
	c.Next()
}
