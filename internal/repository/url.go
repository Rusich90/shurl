package repository

import (
	"context"
	"fmt"
	"sync"

	"github.com/Rusich90/shurl.git/internal/model"
	"github.com/Rusich90/shurl.git/internal/storage"
)

type FileURLRepository struct {
	urls        map[string]string
	fileStorage storage.FileStorage
	mu          sync.Mutex
}

func NewFileURLRepository(fileStorage storage.FileStorage) (*FileURLRepository, error) {
	repo := &FileURLRepository{
		urls:        make(map[string]string),
		fileStorage: fileStorage,
	}

	err := repo.loadFromStorage()
	if err != nil {
		return nil, fmt.Errorf("failed to load store: %w", err)
	}

	return repo, nil
}

func (r *FileURLRepository) Get(ctx context.Context, id string) (string, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	select {
	case <-ctx.Done():
		return "", false
	default:
	}

	url, ok := r.urls[id]
	return url, ok
}

func (r *FileURLRepository) SaveIfNotExists(ctx context.Context, row model.URLRow) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	select {
	case <-ctx.Done():
		return false
	default:
	}

	if _, exists := r.urls[row.ShortURL]; exists {
		return false
	}

	r.urls[row.ShortURL] = row.OriginalURL
	err := r.fileStorage.SaveRow(row)
	return err == nil
}

func (r *FileURLRepository) SaveBatch(ctx context.Context, rows []model.URLRow) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, row := range rows {
		if _, exists := r.urls[row.ShortURL]; exists {
			return fmt.Errorf("conflict: short URL %s already exists", row.ShortURL)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}

	for _, row := range rows {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err := r.fileStorage.SaveRow(row)
		if err != nil {
			return fmt.Errorf("failed to save row %s: %w", row.ShortURL, err)
		}
		r.urls[row.ShortURL] = row.OriginalURL
	}

	return nil
}

func (r *FileURLRepository) loadFromStorage() error {
	urls, err := r.fileStorage.GetURLs()
	if err != nil {
		return fmt.Errorf("failed fileStorage.GetURLs: %w", err)
	}

	for _, url := range urls {
		r.urls[url.ShortURL] = url.OriginalURL
	}

	return nil
}
