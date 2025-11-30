package repository

import "github.com/Rusich90/shurl.git/internal/model"

type URLRepository interface {
	Get(id string) (string, bool)
	SaveIfNotExists(row model.URLRow) bool
}
