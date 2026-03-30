// Package authcontext предоставляет функции для работы с пользователем в контексте gRPC.
//
// Используется для хранения и извлечения идентификатора пользователя из контекста запроса.
package authcontext

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

type contextKey string

const userIDKey contextKey = "userID"

// SetUserID устанавливает идентификатор пользователя в контексте запроса.
func SetUserID(ctx context.Context, id *uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDKey, id)
}

// GetUserID извлекает идентификатор пользователя из контекста запроса.
//
// Возвращает ошибку, если идентификатор не найден или имеет неверный тип.
func GetUserID(ctx context.Context) (*uuid.UUID, error) {
	val := ctx.Value(userIDKey)
	if val == nil {
		return nil, errors.New("userID not found in context")
	}

	id, ok := val.(*uuid.UUID)
	if !ok {
		return nil, errors.New("userID has invalid type in context")
	}

	return id, nil
}
