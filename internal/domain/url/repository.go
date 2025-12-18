package domain

import (
	"context"

	"github.com/Rusich90/shurl.git/internal/transport/http/dto"
)

type URLRepository interface {
	Get(ctx context.Context, id string) (string, bool)
	GetAllByUserID(ctx context.Context, userID string) ([]URL, error)
	SaveIfNotExists(ctx context.Context, row dto.URLRow) error
	SaveBatch(ctx context.Context, rows []dto.URLRow) error
	GetByOriginalURL(ctx context.Context, originalURL string) (string, bool)
	Close() error
	Ping(ctx context.Context) error
}
