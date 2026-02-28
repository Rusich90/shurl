// Package authcontext предоставляет функции для работы с пользователем в контексте Gin.
//
// Используется для хранения и извлечения идентификатора пользователя из контекста запроса.
package authcontext

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type contextKey string

const userIDKey contextKey = "userID"

// SetUserID устанавливает идентификатор пользователя в контексте запроса.
func SetUserID(c *gin.Context, id *uuid.UUID) {
	c.Set(string(userIDKey), id)
}

// GetUserID извлекает идентификатор пользователя из контекста запроса.
//
// Возвращает ошибку, если идентификатор не найден или имеет неверный тип.
func GetUserID(c *gin.Context) (*uuid.UUID, error) {
	val, exists := c.Get(string(userIDKey))
	if !exists || val == nil {
		return nil, errors.New("userID not found in context")
	}

	id, ok := val.(*uuid.UUID)
	if !ok {
		return nil, errors.New("userID has invalid type in context")
	}

	return id, nil
}
