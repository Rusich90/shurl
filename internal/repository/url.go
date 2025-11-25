package repository

import (
	"fmt"

	"github.com/Rusich90/shurl.git/internal/model"
	"github.com/Rusich90/shurl.git/internal/storage"
)

type URLStore struct {
	urls        map[string]string
	fileStorage storage.FileStorage
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

func (s *URLStore) SaveWithID(row model.URLRow) {
	s.urls[row.ShortURL] = row.OriginalURL
	s.fileStorage.SaveRow(row)
}

func (s *URLStore) Get(id string) (string, bool) {
	url, ok := s.urls[id]
	return url, ok
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
