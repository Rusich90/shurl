package service

import (
	"context"
	"fmt"
	"log"
	"net/url"

	"github.com/Rusich90/shurl.git/internal/config"
	internalErrors "github.com/Rusich90/shurl.git/internal/errors"
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

type CreateShortURLResult struct {
	URL   string
	IsNew bool
}

func (s *URLService) CreateShortURL(ctx context.Context, originalURL string) (*CreateShortURLResult, error) {
	for {
		id, err := idgen.GenerateID()
		if err != nil {
			return nil, err
		}

		row := model.URLRow{ShortURL: id, OriginalURL: originalURL}

		err = s.repo.SaveIfNotExists(ctx, row)
		if err != nil {
			if internalErrors.IsErrOriginalURLConflict(err) {
				if shortURL, exists := s.repo.GetByOriginalURL(ctx, originalURL); exists {
					resultURL, err := url.JoinPath(s.cfg.BaseURL, shortURL)
					if err != nil {
						return nil, fmt.Errorf("failed to create short URL: %w", err)
					}
					return &CreateShortURLResult{
						URL:   resultURL,
						IsNew: false,
					}, nil
				}
			} else {
				log.Printf("Error saving URL: %v, generating new ID", err)
				continue
			}
		}

		shortURL, err := url.JoinPath(s.cfg.BaseURL, id)
		if err != nil {
			return nil, fmt.Errorf("failed to create short URL: %w", err)
		}

		return &CreateShortURLResult{
			URL:   shortURL,
			IsNew: true,
		}, nil
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
