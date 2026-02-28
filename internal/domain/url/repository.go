package domain

import (
	"context"

	"github.com/google/uuid"
)

// URLRepository определяет интерфейс для работы с хранилищем URL.
//
// Реализовано в file/url.go и postgres/url.go.
type URLRepository interface {
	// Get возвращает URL по короткому идентификатору.
	// Возвращает false, если URL не найден.
	Get(ctx context.Context, id string) (URL, bool)

	// GetAllByUserID возвращает все URL, принадлежащие указанному пользователю.
	GetAllByUserID(ctx context.Context, userID *uuid.UUID) ([]URL, error)

	// DeleteURLs помечает указанные URL как удаленные для указанного пользователя.
	DeleteURLs(ctx context.Context, IDs []string, userID *uuid.UUID) error

	// SaveIfNotExists сохраняет URL, если короткий идентификатор еще не занят.
	// Возвращает ErrShortURLConflict или ErrOriginalURLConflict при конфликте.
	SaveIfNotExists(ctx context.Context, row URL) error

	// SaveBatch сохраняет несколько URL в одной транзакции.
	SaveBatch(ctx context.Context, rows []URL) error

	// GetByOriginalURL возвращает короткий идентификатор по исходному URL.
	GetByOriginalURL(ctx context.Context, originalURL string) (string, bool)

	// Close закрывает соединение с хранилищем (если применимо).
	Close() error

	// Ping проверяет доступность хранилища.
	Ping(ctx context.Context) error
}
