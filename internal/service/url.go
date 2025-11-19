package service

import (
	"log"

	"github.com/Rusich90/shurl.git/internal/config"
	"github.com/Rusich90/shurl.git/internal/idgen"
	"github.com/Rusich90/shurl.git/internal/model"
	"github.com/Rusich90/shurl.git/internal/repository"
)

type URLService struct {
	repo *repository.URLStore
	cfg  *config.Config
}

func NewURLService(repo *repository.URLStore, cfg *config.Config) *URLService {
	return &URLService{
		repo: repo,
		cfg:  cfg,
	}
}

func (s *URLService) CreateShortURL(originalURL string) (string, error) {
	for {
		id, err := idgen.GenerateID()
		if err != nil {
			return "", err
		}

		row := model.URLRow{ShortURL: id, OriginalURL: originalURL}
		if _, ok := s.repo.Get(id); !ok {
			s.repo.SaveWithID(row)
			return id, nil
		}

		log.Printf("Collision detected for ID: %s, generating new ID", id)
	}
}

func (s *URLService) GetOriginalURL(id string) (string, bool) {
	return s.repo.Get(id)
}
