package repository

import (
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

func (r *FileURLRepository) Get(id string) (string, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	url, ok := r.urls[id]
	return url, ok
}

func (r *FileURLRepository) SaveIfNotExists(row model.URLRow) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.urls[row.ShortURL]; exists {
		return false
	}

	r.urls[row.ShortURL] = row.OriginalURL
	r.fileStorage.SaveRow(row)
	return true
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
