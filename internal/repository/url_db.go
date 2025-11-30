package repository

import (
	"context"
	"database/sql"
	"sync"

	"github.com/Rusich90/shurl.git/internal/model"
)

type DBURLRepository struct {
	db *sql.DB
	mu sync.Mutex
}

func NewDBURLRepository(db *sql.DB) *DBURLRepository {
	return &DBURLRepository{
		db: db,
	}
}

func (r *DBURLRepository) Get(id string) (string, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var originalURL string
	query := `SELECT original_url FROM urls WHERE short_url = $1`
	err := r.db.QueryRowContext(context.Background(), query, id).Scan(&originalURL)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", false
		}
		return "", false
	}

	return originalURL, true
}

func (r *DBURLRepository) SaveIfNotExists(row model.URLRow) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM urls WHERE short_url = $1)`
	err := r.db.QueryRowContext(context.Background(), checkQuery, row.ShortURL).Scan(&exists)
	if err != nil {
		return false
	}

	if exists {
		return false
	}

	insertQuery := `INSERT INTO urls (short_url, original_url) VALUES ($1, $2)`
	_, err = r.db.ExecContext(context.Background(), insertQuery, row.ShortURL, row.OriginalURL)
	return err == nil
}
