package service

import (
	"github.com/Rusich90/shurl.git/internal/idgen"
	"github.com/Rusich90/shurl.git/internal/repository"
)

type URLService struct {
	repo *repository.URLStore
}

func NewURLService(repo *repository.URLStore) *URLService {
	return &URLService{
		repo: repo,
	}
}

func (s *URLService) CreateShortURL(originalURL string) string {
	id := idgen.GenerateID()
	s.repo.SaveWithID(id, originalURL)
	return id
}

func (s *URLService) GetOriginalURL(id string) (string, bool) {
	return s.repo.Get(id)
}
