package repository

import (
	"context"

	domain "github.com/Rusich90/shurl.git/internal/domain/url"
	"github.com/Rusich90/shurl.git/internal/model"
)

type URLRepository interface {
	Get(ctx context.Context, id string) (string, bool)
	GetAllByUserID(ctx context.Context, userID string) ([]domain.URL, error)
	SaveIfNotExists(ctx context.Context, row model.URLRow) error
	SaveBatch(ctx context.Context, rows []model.URLRow) error
	GetByOriginalURL(ctx context.Context, originalURL string) (string, bool)
	Close() error
	Ping(ctx context.Context) error
}
