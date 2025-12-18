package file

import (
	"context"
	"fmt"
	"sync"

	domainurl "github.com/Rusich90/shurl.git/internal/domain/url"
	"github.com/Rusich90/shurl.git/internal/transport/http/dto"
)

type FileURLRepository struct {
	urls        map[string]string
	fileStorage FileStorage
	mu          sync.Mutex
}

func (r *FileURLRepository) GetAllByUserID(ctx context.Context, userID string) ([]domainurl.URL, error) {
	//TODO implement me
	panic("implement me")
}

func NewFileURLRepository(fileStorage FileStorage) (*FileURLRepository, error) {
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

func (r *FileURLRepository) SaveIfNotExists(ctx context.Context, row dto.URLRow) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if _, exists := r.urls[row.ShortURL]; exists {
		return domainurl.ErrShortURLConflict
	}

	for _, originalURL := range r.urls {
		if originalURL == row.OriginalURL {
			return domainurl.ErrOriginalURLConflict
		}
	}

	r.urls[row.ShortURL] = row.OriginalURL
	err := r.fileStorage.SaveRow(row)
	if err != nil {
		return err
	}
	return nil
}

func (r *FileURLRepository) SaveBatch(ctx context.Context, rows []dto.URLRow) error {
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

func (r *FileURLRepository) GetByOriginalURL(ctx context.Context, originalURL string) (string, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	select {
	case <-ctx.Done():
		return "", false
	default:
	}

	for shortURL, url := range r.urls {
		if url == originalURL {
			return shortURL, true
		}
	}

	return "", false
}

func (r *FileURLRepository) Close() error {
	return nil
}

func (r *FileURLRepository) Ping(ctx context.Context) error {
	return nil
}
