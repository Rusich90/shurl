package domain

import (
	"context"

	"github.com/google/uuid"
)

type URLRepository interface {
	Get(ctx context.Context, id string) (string, bool)
	GetAllByUserID(ctx context.Context, userID *uuid.UUID) ([]URL, error)
	SaveIfNotExists(ctx context.Context, row URL) error
	SaveBatch(ctx context.Context, rows []URL) error
	GetByOriginalURL(ctx context.Context, originalURL string) (string, bool)
	Close() error
	Ping(ctx context.Context) error
}
