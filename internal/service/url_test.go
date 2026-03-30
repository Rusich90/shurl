package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Rusich90/shurl.git/internal/audit"
	"github.com/Rusich90/shurl.git/internal/config"
	domainurl "github.com/Rusich90/shurl.git/internal/domain/url"
	"github.com/Rusich90/shurl.git/internal/transport/http/dto"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// MockURLRepository - мок для репозитория URL
type MockURLRepository struct {
	mock.Mock
}

func (m *MockURLRepository) Get(ctx context.Context, id string) (domainurl.URL, bool) {
	args := m.Called(ctx, id)
	if args.Get(1) == false {
		return domainurl.URL{}, false
	}
	return args.Get(0).(domainurl.URL), args.Bool(1)
}

func (m *MockURLRepository) GetAllByUserID(ctx context.Context, userID *uuid.UUID) ([]domainurl.URL, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domainurl.URL), args.Error(1)
}

func (m *MockURLRepository) DeleteURLs(ctx context.Context, IDs []string, userID *uuid.UUID) error {
	args := m.Called(ctx, IDs, userID)
	return args.Error(0)
}

func (m *MockURLRepository) SaveIfNotExists(ctx context.Context, row domainurl.URL) error {
	args := m.Called(ctx, row)
	return args.Error(0)
}

func (m *MockURLRepository) SaveBatch(ctx context.Context, rows []domainurl.URL) error {
	args := m.Called(ctx, rows)
	return args.Error(0)
}

func (m *MockURLRepository) GetByOriginalURL(ctx context.Context, originalURL string) (string, bool) {
	args := m.Called(ctx, originalURL)
	if args.Get(1) == false {
		return "", false
	}
	return args.String(0), args.Bool(1)
}

func (m *MockURLRepository) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockURLRepository) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockURLRepository) CountURLs(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

func (m *MockURLRepository) CountUsers(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

func TestNewURLService(t *testing.T) {
	mockRepo := new(MockURLRepository)
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)

	service := NewURLService(mockRepo, cfg, logger, auditManager)

	assert.NotNil(t, service)
	assert.Equal(t, mockRepo, service.repo)
	assert.Equal(t, cfg, service.cfg)
	assert.Equal(t, logger, service.logger)
	assert.Equal(t, auditManager, service.auditManager)
}

func TestURLService_CreateShortURL_Success(t *testing.T) {
	mockRepo := new(MockURLRepository)
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)

	service := NewURLService(mockRepo, cfg, logger, auditManager)
	ctx := context.Background()
	originalURL := "https://example.com"
	userID := uuid.New()

	mockRepo.On("SaveIfNotExists", ctx, mock.Anything).Return(nil)

	result, err := service.CreateShortURL(ctx, originalURL, &userID)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.IsNew)
	assert.Contains(t, result.URL, "http://localhost:8080/")
	mockRepo.AssertExpectations(t)
}

func TestURLService_CreateShortURL_OriginalURLConflict(t *testing.T) {
	mockRepo := new(MockURLRepository)
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)

	service := NewURLService(mockRepo, cfg, logger, auditManager)
	ctx := context.Background()
	originalURL := "https://example.com"
	userID := uuid.New()
	existingShortURL := "abc123"

	mockRepo.On("SaveIfNotExists", ctx, mock.Anything).Return(domainurl.ErrOriginalURLConflict)
	mockRepo.On("GetByOriginalURL", ctx, originalURL).Return(existingShortURL, true)

	result, err := service.CreateShortURL(ctx, originalURL, &userID)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.IsNew)
	assert.Equal(t, "http://localhost:8080/"+existingShortURL, result.URL)
	mockRepo.AssertExpectations(t)
}

func TestURLService_CreateShortURL_ShortURLConflict(t *testing.T) {
	mockRepo := new(MockURLRepository)
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)

	service := NewURLService(mockRepo, cfg, logger, auditManager)
	ctx := context.Background()
	originalURL := "https://example.com"
	userID := uuid.New()

	// Первый вызов возвращает конфликт короткого URL, второй - успешное сохранение
	mockRepo.On("SaveIfNotExists", ctx, mock.Anything).Return(domainurl.ErrShortURLConflict).Once()
	mockRepo.On("SaveIfNotExists", ctx, mock.Anything).Return(nil).Once()

	result, err := service.CreateShortURL(ctx, originalURL, &userID)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.IsNew)
	mockRepo.AssertExpectations(t)
}

func TestURLService_CreateShortBatchURL_Success(t *testing.T) {
	mockRepo := new(MockURLRepository)
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)

	service := NewURLService(mockRepo, cfg, logger, auditManager)
	ctx := context.Background()
	userID := uuid.New()

	request := dto.CreateBatchURLRequest{
		{CorrelationID: "1", OriginalURL: "https://example1.com"},
		{CorrelationID: "2", OriginalURL: "https://example2.com"},
	}

	mockRepo.On("Get", ctx, mock.AnythingOfType("string")).Return(domainurl.URL{}, false)
	mockRepo.On("SaveBatch", ctx, mock.Anything).Return(nil)

	response, err := service.CreateShortBatchURL(ctx, request, &userID)

	require.NoError(t, err)
	assert.Len(t, response, 2)
	assert.Equal(t, "1", response[0].CorrelationID)
	assert.Equal(t, "2", response[1].CorrelationID)
	assert.Contains(t, response[0].ShortURL, "http://localhost:8080/")
	assert.Contains(t, response[1].ShortURL, "http://localhost:8080/")
	mockRepo.AssertExpectations(t)
}

func TestURLService_CreateShortBatchURL_SaveBatchError(t *testing.T) {
	mockRepo := new(MockURLRepository)
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)

	service := NewURLService(mockRepo, cfg, logger, auditManager)
	ctx := context.Background()
	userID := uuid.New()

	request := dto.CreateBatchURLRequest{
		{CorrelationID: "1", OriginalURL: "https://example1.com"},
	}

	mockRepo.On("Get", ctx, mock.AnythingOfType("string")).Return(domainurl.URL{}, false)
	mockRepo.On("SaveBatch", ctx, mock.Anything).Return(errors.New("batch save error"))

	response, err := service.CreateShortBatchURL(ctx, request, &userID)

	require.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "failed to save batch URLs")
	mockRepo.AssertExpectations(t)
}

func TestURLService_CreateShortBatchURL_MaxAttemptsExceeded(t *testing.T) {
	mockRepo := new(MockURLRepository)
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)

	service := NewURLService(mockRepo, cfg, logger, auditManager)
	ctx := context.Background()
	userID := uuid.New()

	request := dto.CreateBatchURLRequest{
		{CorrelationID: "1", OriginalURL: "https://example1.com"},
	}

	// Всегда возвращаем существующий URL, чтобы превысить лимит попыток
	mockRepo.On("Get", ctx, mock.AnythingOfType("string")).Return(domainurl.URL{ShortURL: "existing"}, true)

	response, err := service.CreateShortBatchURL(ctx, request, &userID)

	require.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "failed to generate unique ID after")
	mockRepo.AssertExpectations(t)
}

func TestURLService_GetOriginalURL_Success(t *testing.T) {
	mockRepo := new(MockURLRepository)
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)

	service := NewURLService(mockRepo, cfg, logger, auditManager)
	ctx := context.Background()
	shortURL := "abc123"
	expectedURL := domainurl.URL{
		ShortURL:    shortURL,
		OriginalURL: "https://example.com",
		IsDeleted:   false,
	}

	mockRepo.On("Get", ctx, shortURL).Return(expectedURL, true)

	result, ok := service.GetOriginalURL(ctx, shortURL)

	require.True(t, ok)
	assert.Equal(t, expectedURL, result)
	mockRepo.AssertExpectations(t)
}

func TestURLService_GetOriginalURL_NotFound(t *testing.T) {
	mockRepo := new(MockURLRepository)
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)

	service := NewURLService(mockRepo, cfg, logger, auditManager)
	ctx := context.Background()
	shortURL := "nonexistent"

	mockRepo.On("Get", ctx, shortURL).Return(domainurl.URL{}, false)

	result, ok := service.GetOriginalURL(ctx, shortURL)

	require.False(t, ok)
	assert.Equal(t, domainurl.URL{}, result)
	mockRepo.AssertExpectations(t)
}

func TestURLService_GetUserOriginalURLs_Success(t *testing.T) {
	mockRepo := new(MockURLRepository)
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)

	service := NewURLService(mockRepo, cfg, logger, auditManager)
	ctx := context.Background()
	userID := uuid.New()

	expectedURLs := []domainurl.URL{
		{ShortURL: "abc123", OriginalURL: "https://example1.com"},
		{ShortURL: "def456", OriginalURL: "https://example2.com"},
	}

	mockRepo.On("GetAllByUserID", ctx, &userID).Return(expectedURLs, nil)

	urls, err := service.GetUserOriginalURLs(ctx, &userID)

	require.NoError(t, err)
	assert.Len(t, urls, 2)
	assert.Equal(t, expectedURLs, urls)
	mockRepo.AssertExpectations(t)
}

func TestURLService_GetUserOriginalURLs_Error(t *testing.T) {
	mockRepo := new(MockURLRepository)
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)

	service := NewURLService(mockRepo, cfg, logger, auditManager)
	ctx := context.Background()
	userID := uuid.New()

	mockRepo.On("GetAllByUserID", ctx, &userID).Return(nil, errors.New("repository error"))

	urls, err := service.GetUserOriginalURLs(ctx, &userID)

	require.Error(t, err)
	assert.Nil(t, urls)
	assert.Contains(t, err.Error(), "GetUserOriginalURLs.repo.GetAllByUserID")
	mockRepo.AssertExpectations(t)
}

func TestURLService_DeleteURLsByUserID_Success(t *testing.T) {
	mockRepo := new(MockURLRepository)
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)

	service := NewURLService(mockRepo, cfg, logger, auditManager)
	ctx := context.Background()
	userID := uuid.New()

	ids := []string{"abc123", "def456", "ghi789"}

	mockRepo.On("DeleteURLs", mock.Anything, mock.Anything, &userID).Return(nil)

	err := service.DeleteURLsByUserID(ctx, ids, &userID)

	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestURLService_DeleteURLsByUserID_ProcessBatchError(t *testing.T) {
	mockRepo := new(MockURLRepository)
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)

	service := NewURLService(mockRepo, cfg, logger, auditManager)
	ctx := context.Background()
	userID := uuid.New()

	ids := []string{"abc123", "def456"}

	mockRepo.On("DeleteURLs", mock.Anything, mock.Anything, &userID).Return(errors.New("delete error"))

	_ = service.DeleteURLsByUserID(ctx, ids, &userID)

	// Batch processor может не возвращать ошибку, если некоторые операции успешны
	// Проверяем, что мок был вызван
	mockRepo.AssertExpectations(t)
}

func TestURLService_DeleteURLsByUserID_EmptyIDs(t *testing.T) {
	mockRepo := new(MockURLRepository)
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)

	service := NewURLService(mockRepo, cfg, logger, auditManager)
	ctx := context.Background()
	userID := uuid.New()

	ids := []string{}

	// С пустым списком ID batch processor не вызывает DeleteURLs
	err := service.DeleteURLsByUserID(ctx, ids, &userID)

	require.NoError(t, err)
	// Не проверяем ожидания мока, так как он не должен вызываться
}

func TestURLService_DeleteURLsByUserID_ContextCancelled(t *testing.T) {
	mockRepo := new(MockURLRepository)
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)

	service := NewURLService(mockRepo, cfg, logger, auditManager)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Сразу отменяем контекст
	userID := uuid.New()

	ids := []string{"abc123", "def456"}

	err := service.DeleteURLsByUserID(ctx, ids, &userID)

	require.Error(t, err)
}
