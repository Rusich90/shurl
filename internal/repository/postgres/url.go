package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	domainurl "github.com/Rusich90/shurl.git/internal/domain/url"
	"github.com/Rusich90/shurl.git/internal/transport/http/dto"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

type DBURLRepository struct {
	db *sql.DB
}

func NewDBURLRepository(db *sql.DB) *DBURLRepository {
	return &DBURLRepository{
		db: db,
	}
}

func (r *DBURLRepository) Get(ctx context.Context, id string) (string, bool) {
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

func (r *DBURLRepository) GetAllByUserID(ctx context.Context, userID string) ([]domainurl.URL, error) {
	query := `
		SELECT short_url, original_url
		FROM urls 
		WHERE user_id = $1 
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	urls := make([]domainurl.URL, 0)

	for rows.Next() {
		var u domainurl.URL
		if err := rows.Scan(&u.ShortURL, &u.OriginalURL); err != nil {
			return nil, err
		}
		urls = append(urls, u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return urls, nil
}

func (r *DBURLRepository) SaveIfNotExists(ctx context.Context, row dto.URLRow) error {
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM urls WHERE short_url = $1)`
	err := r.db.QueryRowContext(ctx, checkQuery, row.ShortURL).Scan(&exists)
	if err != nil {
		return err
	}

	if exists {
		return domainurl.ErrShortURLConflict
	}

	insertQuery := `INSERT INTO urls (short_url, original_url, user_id) VALUES ($1, $2, $3)`
	_, err = r.db.ExecContext(ctx, insertQuery, row.ShortURL, row.OriginalURL, row.UserID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return domainurl.ErrOriginalURLConflict
		}
		return err
	}
	return nil
}

func (r *DBURLRepository) SaveBatch(ctx context.Context, rows []dto.URLRow) error {
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

	insertStmt, err := tx.PrepareContext(ctx, `INSERT INTO urls (short_url, original_url, user_id) VALUES ($1, $2, $3)`)
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

		_, err = insertStmt.ExecContext(ctx, row.ShortURL, row.OriginalURL, row.UserID)
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

func (r *DBURLRepository) Close() error {
	return r.db.Close()
}

func (r *DBURLRepository) Ping(ctx context.Context) error {
	return r.db.PingContext(ctx)
}
