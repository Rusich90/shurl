package repository

import (
	"context"

	"github.com/Rusich90/shurl.git/internal/model"
)

type URLRepository interface {
	Get(ctx context.Context, id string) (string, bool)
	SaveIfNotExists(ctx context.Context, row model.URLRow) bool
	SaveBatch(ctx context.Context, rows []model.URLRow) error
}
