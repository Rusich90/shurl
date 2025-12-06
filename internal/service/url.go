package service

import (
	"context"
	"fmt"
	"log"
	"net/url"

	"github.com/Rusich90/shurl.git/internal/config"
	"github.com/Rusich90/shurl.git/internal/idgen"
	"github.com/Rusich90/shurl.git/internal/model"
	"github.com/Rusich90/shurl.git/internal/repository"
)

type URLService struct {
	repo repository.URLRepository
	cfg  *config.Config
}

func NewURLService(repo repository.URLRepository, cfg *config.Config) *URLService {
	return &URLService{
		repo: repo,
		cfg:  cfg,
	}
}

func (s *URLService) CreateShortURL(ctx context.Context, originalURL string) (string, error) {
	for {
		id, err := idgen.GenerateID()
		if err != nil {
			return "", err
		}

		row := model.URLRow{ShortURL: id, OriginalURL: originalURL}

		if s.repo.SaveIfNotExists(ctx, row) {
			shortURL, err := url.JoinPath(s.cfg.BaseURL, id)
			if err != nil {
				return "", fmt.Errorf("failed to create short URL: %w", err)
			}
			return shortURL, nil
		}

		log.Printf("Collision detected for ID: %s, generating new ID", id)
	}
}

func (s *URLService) CreateShortBatchURL(ctx context.Context, request model.CreateBatchURLRequest) (model.CreateBatchURLResponse, error) {
	var urlRows []model.URLRow
	var responses model.CreateBatchURLResponse

	for _, req := range request {
		var id string
		var err error

		maxAttempts := 10
		attempts := 0
		for attempts < maxAttempts {
			id, err = idgen.GenerateID()
			if err != nil {
				return nil, fmt.Errorf("failed to generate ID: %w", err)
			}

			if _, exists := s.repo.Get(ctx, id); !exists {
				break
			}
			attempts++
		}

		if attempts >= maxAttempts {
			return nil, fmt.Errorf("failed to generate unique ID after %d attempts", maxAttempts)
		}

		row := model.URLRow{
			ShortURL:    id,
			OriginalURL: req.OriginalURL,
		}
		urlRows = append(urlRows, row)

		shortURL, err := url.JoinPath(s.cfg.BaseURL, id)
		if err != nil {
			return nil, fmt.Errorf("failed to create short URL: %w", err)
		}

		responses = append(responses, model.BatchURLResponseItem{
			CorrelationID: req.CorrelationID,
			ShortURL:      shortURL,
		})
	}

	err := s.repo.SaveBatch(ctx, urlRows)
	if err != nil {
		return nil, fmt.Errorf("failed to save batch URLs: %w", err)
	}

	return responses, nil
}

func (s *URLService) GetOriginalURL(ctx context.Context, id string) (string, bool) {
	return s.repo.Get(ctx, id)
}
