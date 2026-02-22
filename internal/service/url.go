// Package service предоставляет бизнес-логику приложения.
//
// Содержит сервисы для работы с URL, аутентификацией и состоянием сервиса.
package service

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/Rusich90/shurl.git/internal/audit"
	"github.com/Rusich90/shurl.git/internal/batch"
	"github.com/Rusich90/shurl.git/internal/config"
	domainurl "github.com/Rusich90/shurl.git/internal/domain/url"
	"github.com/Rusich90/shurl.git/internal/idgen"
	"github.com/Rusich90/shurl.git/internal/transport/http/dto"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// URLService предоставляет операции для управления короткими URL.
//
// Обрабатывает создание, получение и удаление URL, а также отправку
// аудит-событий.
type URLService struct {
	repo         domainurl.URLRepository
	cfg          *config.Config
	logger       *zap.Logger
	auditManager *audit.Manager
}

// NewURLService создает новый URLService с указанными зависимостями.
func NewURLService(repo domainurl.URLRepository, cfg *config.Config, logger *zap.Logger, auditManager *audit.Manager) *URLService {
	return &URLService{
		repo:         repo,
		cfg:          cfg,
		logger:       logger,
		auditManager: auditManager,
	}
}

// CreateShortURLResult содержит результат операции создания короткой URL.
type CreateShortURLResult struct {
	// URL — сгенерированная или существующая короткая ссылка.
	URL string
	// IsNew — флаг, указывающий, что ссылка была создана в этой операции.
	IsNew bool
}

// CreateShortURL создает новую короткую URL для указанного исходного URL.
//
// Если URL уже существует, возвращает существующую ссылку с IsNew=false.
// Генерирует уникальный идентификатор, сохраняет URL и отправляет
// аудит-событие.
//
// Пример использования:
//
//	result, err := urlService.CreateShortURL(ctx, "https://example.com", userID)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("Short URL:", result.URL)
func (s *URLService) CreateShortURL(ctx context.Context, originalURL string, userID *uuid.UUID) (*CreateShortURLResult, error) {
	for {
		id, err := idgen.GenerateID()
		if err != nil {
			return nil, err
		}

		row := domainurl.URL{ShortURL: id, OriginalURL: originalURL, UserID: userID}

		err = s.repo.SaveIfNotExists(ctx, row)
		if err != nil {
			if domainurl.IsErrOriginalURLConflict(err) {
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

		auditEvent := audit.NewAuditEvent(audit.ActionShorten, userID, originalURL)
		go s.auditManager.NotifyAll(ctx, auditEvent)

		return &CreateShortURLResult{
			URL:   shortURL,
			IsNew: true,
		}, nil
	}
}

// CreateShortBatchURL создает несколько коротких URL за одну операцию.
//
// Принимает пакет запросов и возвращает пакет ответов с корреляционными
// идентификаторами для сопоставления.
func (s *URLService) CreateShortBatchURL(ctx context.Context, request dto.CreateBatchURLRequest, userID *uuid.UUID) (dto.CreateBatchURLResponse, error) {
	var urlRows []domainurl.URL
	var responses dto.CreateBatchURLResponse

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

		row := domainurl.URL{
			ShortURL:    id,
			OriginalURL: req.OriginalURL,
			UserID:      userID,
		}
		urlRows = append(urlRows, row)

		shortURL, err := url.JoinPath(s.cfg.BaseURL, id)
		if err != nil {
			return nil, fmt.Errorf("failed to create short URL: %w", err)
		}

		responses = append(responses, dto.BatchURLResponseItem{
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

// GetOriginalURL возвращает исходный URL по короткому идентификатору.
//
// Возвращает URL и true, если найден, или пустой URL и false, если не найден.
func (s *URLService) GetOriginalURL(ctx context.Context, id string) (domainurl.URL, bool) {
	URL, err := s.repo.Get(ctx, id)

	auditEvent := audit.NewAuditEvent(audit.ActionShorten, nil, URL.OriginalURL)
	go s.auditManager.NotifyAll(ctx, auditEvent)

	return URL, err
}

// GetUserOriginalURLs возвращает все URL, принадлежащие указанному пользователю.
func (s *URLService) GetUserOriginalURLs(ctx context.Context, userID *uuid.UUID) ([]domainurl.URL, error) {
	urls, err := s.repo.GetAllByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("GetUserOriginalURLs.repo.GetAllByUserID: %w", err)
	}

	return urls, nil
}

// DeleteURLsByUserID помечает указанные URL как удаленные для пользователя.
//
// Операция выполняется асинхронно с использованием пакетной обработки.
func (s *URLService) DeleteURLsByUserID(ctx context.Context, IDs []string, userID *uuid.UUID) error {
	start := time.Now()

	s.logger.Info("Starting DeleteURLsByUserID",
		zap.Int("url_count", len(IDs)),
		zap.Any("user_id", userID),
	)

	processor, err := batch.NewBatchProcessor(10, 50000, s.logger)
	if err != nil {
		return fmt.Errorf("batch.NewBatchProcessor: %w", err)
	}

	processFunc := func(ctx context.Context, items []string, uid *uuid.UUID) error {
		return s.repo.DeleteURLs(ctx, items, uid)
	}

	stats, err := processor.ProcessBatch(ctx, IDs, userID, processFunc)
	if err != nil {
		return fmt.Errorf("failed to process batch: %w", err)
	}

	duration := time.Since(start)
	s.logger.Info("DeleteURLsByUserID completed",
		zap.Duration("duration", duration),
		zap.Int64("successful_deletes", stats.SuccessfulDeletes),
		zap.Int64("failed_deletes", stats.FailedDeletes),
		zap.Int64("total_processed", stats.SuccessfulDeletes+stats.FailedDeletes),
		zap.Int64("processed_chunks", stats.ProcessedChunks),
		zap.Int("url_count", len(IDs)),
	)

	return nil
}
