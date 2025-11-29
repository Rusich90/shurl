package repository

import (
	"fmt"
	"sync"

	"github.com/Rusich90/shurl.git/internal/model"
	"github.com/Rusich90/shurl.git/internal/storage"
)

type URLStore struct {
	urls        map[string]string
	fileStorage storage.FileStorage
	mu          sync.Mutex
}

func NewURLStore(fileStorage storage.FileStorage) (*URLStore, error) {
	store := URLStore{
		urls:        make(map[string]string),
		fileStorage: fileStorage,
	}

	err := store.loadFromStorage()
	if err != nil {
		return nil, fmt.Errorf("failed to load store: %w", err)
	}

	return &store, nil
}

func (s *URLStore) Get(id string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	url, ok := s.urls[id]
	return url, ok
}

func (s *URLStore) SaveIfNotExists(row model.URLRow) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.urls[row.ShortURL]; exists {
		return false
	}

	s.urls[row.ShortURL] = row.OriginalURL
	s.fileStorage.SaveRow(row)
	return true
}

func (s *URLStore) loadFromStorage() error {
	urls, err := s.fileStorage.GetURLs()
	if err != nil {
		return fmt.Errorf("failed fileStorage.GetURLs: %w", err)
	}

	for _, url := range urls {
		s.urls[url.ShortURL] = url.OriginalURL
	}

	return nil
}
