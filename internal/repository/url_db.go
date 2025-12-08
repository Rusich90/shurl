package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"

	internalErrors "github.com/Rusich90/shurl.git/internal/errors"
	"github.com/Rusich90/shurl.git/internal/model"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
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

func (r *DBURLRepository) Get(ctx context.Context, id string) (string, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var originalURL string
	query := `SELECT original_url FROM urls WHERE short_url = $1`
	err := r.db.QueryRowContext(ctx, query, id).Scan(&originalURL)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", false
		}
		return "", false
	}

	return originalURL, true
}

func (r *DBURLRepository) SaveIfNotExists(ctx context.Context, row model.URLRow) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM urls WHERE short_url = $1)`
	err := r.db.QueryRowContext(ctx, checkQuery, row.ShortURL).Scan(&exists)
	if err != nil {
		return err
	}

	if exists {
		return internalErrors.ErrShortURLConflict
	}

	insertQuery := `INSERT INTO urls (short_url, original_url) VALUES ($1, $2)`
	_, err = r.db.ExecContext(ctx, insertQuery, row.ShortURL, row.OriginalURL)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return internalErrors.ErrOriginalURLConflict
		}
		return err
	}
	return nil
}

func (r *DBURLRepository) SaveBatch(ctx context.Context, rows []model.URLRow) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	checkStmt, err := tx.PrepareContext(ctx, `SELECT EXISTS(SELECT 1 FROM urls WHERE short_url = $1)`)
	if err != nil {
		return fmt.Errorf("failed to prepare check statement: %w", err)
	}
	defer checkStmt.Close()

	insertStmt, err := tx.PrepareContext(ctx, `INSERT INTO urls (short_url, original_url) VALUES ($1, $2)`)
	if err != nil {
		return fmt.Errorf("failed to prepare insert statement: %w", err)
	}
	defer insertStmt.Close()

	for _, row := range rows {
		var exists bool
		err := checkStmt.QueryRowContext(ctx, row.ShortURL).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check existence of %s: %w", row.ShortURL, err)
		}

		if exists {
			return fmt.Errorf("conflict: short URL %s already exists", row.ShortURL)
		}

		_, err = insertStmt.ExecContext(ctx, row.ShortURL, row.OriginalURL)
		if err != nil {
			return fmt.Errorf("failed to insert row %s: %w", row.ShortURL, err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *DBURLRepository) GetByOriginalURL(ctx context.Context, originalURL string) (string, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var shortURL string
	query := `SELECT short_url FROM urls WHERE original_url = $1`
	err := r.db.QueryRowContext(ctx, query, originalURL).Scan(&shortURL)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", false
		}
		return "", false
	}

	return shortURL, true
}
