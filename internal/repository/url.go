package repository

import (
	"github.com/Rusich90/shurl.git/internal/model"
	"github.com/Rusich90/shurl.git/internal/storage"
)

type URLStore struct {
	urls       map[string]string
	fileStorge storage.FileStorage
}

func NewURLStore(fileStorage storage.FileStorage) *URLStore {
	return &URLStore{
		urls:       make(map[string]string),
		fileStorge: fileStorage,
	}
}

func (s *URLStore) SaveWithID(row model.URLRow) {
	s.urls[row.ShortURL] = row.OriginalURL
	s.fileStorge.SaveRow(row)
}

func (s *URLStore) Get(id string) (string, bool) {
	url, ok := s.urls[id]
	return url, ok
}

func (s *URLStore) LoadFromStorage() error {
	urls, err := s.fileStorge.GetURLs()
	if err != nil {
		return err
	}

	for _, url := range urls {
		s.urls[url.ShortURL] = url.OriginalURL
	}

	return nil
}
