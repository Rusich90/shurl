// Package service предоставляет бизнес-логику приложения.
//
// Содержит сервисы для работы с URL, аутентификацией и состоянием сервиса.
package service

import (
	"context"
	"fmt"

	domainurl "github.com/Rusich90/shurl.git/internal/domain/url"
	"github.com/Rusich90/shurl.git/internal/transport/http/dto"
)

// StatsService предоставляет операции для получения статистики сервиса.
type StatsService struct {
	repo domainurl.URLRepository
}

// NewStatsService создает новый StatsService с указанными зависимостями.
func NewStatsService(repo domainurl.URLRepository) *StatsService {
	return &StatsService{
		repo: repo,
	}
}

// GetStats возвращает статистику по URL и пользователям.
func (s *StatsService) GetStats(ctx context.Context) (*dto.StatsResponse, error) {
	urls, err := s.repo.CountURLs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count URLs: %w", err)
	}

	users, err := s.repo.CountUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}

	return &dto.StatsResponse{
		URLs:  urls,
		Users: users,
	}, nil
}
