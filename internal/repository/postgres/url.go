// Package postgres предоставляет реализацию репозитория URL на основе PostgreSQL.
//
// Использует pgx для подключения к базе данных и поддерживает миграции.
package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	domainurl "github.com/Rusich90/shurl.git/internal/domain/url"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/lib/pq"
)

// DBURLRepository реализует репозиторий URL для PostgreSQL.
type DBURLRepository struct {
	db *sql.DB
}

// NewDBURLRepository создает новый репозиторий с указанным подключением к БД.
func NewDBURLRepository(db *sql.DB) *DBURLRepository {
	return &DBURLRepository{
		db: db,
	}
}

// Get возвращает URL по короткому идентификатору.
//
// Возвращает URL и true, если найден, или пустой URL и false, если не найден.
func (r *DBURLRepository) Get(ctx context.Context, id string) (domainurl.URL, bool) {
	var url domainurl.URL
	query := `SELECT short_url, original_url, user_id, is_deleted FROM urls WHERE short_url = $1`
	err := r.db.QueryRowContext(ctx, query, id).Scan(&url.ShortURL, &url.OriginalURL, &url.UserID, &url.IsDeleted)
	if err != nil {
		return domainurl.URL{}, false
	}

	return url, true
}

// GetAllByUserID возвращает все URL, принадлежащие указанному пользователю.
//
// Возвращает срез URL в порядке убывания времени создания.
func (r *DBURLRepository) GetAllByUserID(ctx context.Context, userID *uuid.UUID) ([]domainurl.URL, error) {
	query := `
		SELECT short_url, original_url, user_id, is_deleted
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
		if err := rows.Scan(&u.ShortURL, &u.OriginalURL, &u.UserID, &u.IsDeleted); err != nil {
			return nil, err
		}
		urls = append(urls, u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return urls, nil
}

// SaveIfNotExists сохраняет URL, если короткий идентификатор еще не занят.
//
// Возвращает ErrShortURLConflict или ErrOriginalURLConflict при конфликте.
func (r *DBURLRepository) SaveIfNotExists(ctx context.Context, row domainurl.URL) error {
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

// DeleteURLs помечает указанные URL как удаленные для указанного пользователя.
func (r *DBURLRepository) DeleteURLs(ctx context.Context, IDs []string, userID *uuid.UUID) error {
	query := `
		UPDATE urls
		SET is_deleted = true
		WHERE short_url = ANY($1) AND user_id = $2
	`
	_, err := r.db.ExecContext(ctx, query, pq.Array(IDs), userID)
	if err != nil {
		return fmt.Errorf("failed to delete URLs: %w", err)
	}

	return nil
}

// SaveBatch сохраняет несколько URL в одной транзакции.
func (r *DBURLRepository) SaveBatch(ctx context.Context, rows []domainurl.URL) error {
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

// GetByOriginalURL возвращает короткий идентификатор по исходному URL.
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

// Close закрывает соединение с базой данных.
func (r *DBURLRepository) Close() error {
	return r.db.Close()
}

// Ping проверяет доступность базы данных.
func (r *DBURLRepository) Ping(ctx context.Context) error {
	return r.db.PingContext(ctx)
}

// CountURLs возвращает количество URL в хранилище.
func (r *DBURLRepository) CountURLs(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM urls`
	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count URLs: %w", err)
	}
	return count, nil
}

// CountUsers возвращает количество уникальных пользователей в хранилище.
func (r *DBURLRepository) CountUsers(ctx context.Context) (int, error) {
	query := `SELECT COUNT(DISTINCT user_id) FROM urls WHERE user_id IS NOT NULL`
	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}
	return count, nil
}
