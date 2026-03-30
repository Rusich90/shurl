package handler

import (
	"context"
	"errors"
	"testing"

	"github.com/Rusich90/shurl.git/internal/audit"
	"github.com/Rusich90/shurl.git/internal/config"
	domainurl "github.com/Rusich90/shurl.git/internal/domain/url"
	"github.com/Rusich90/shurl.git/internal/service"
	"github.com/Rusich90/shurl.git/internal/transport/grpc/authcontext"
	proto "github.com/Rusich90/shurl.git/pkg"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
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

func TestNewHandler(t *testing.T) {
	mockRepo := new(MockURLRepository)
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)
	urlService := service.NewURLService(mockRepo, cfg, logger, auditManager)

	handler := NewHandler(urlService, cfg, logger)

	assert.NotNil(t, handler)
	assert.Equal(t, urlService, handler.urlService)
	assert.Equal(t, cfg, handler.cfg)
	assert.Equal(t, logger, handler.logger)
}

func TestShortenURL_Success(t *testing.T) {
	mockRepo := new(MockURLRepository)
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)
	urlService := service.NewURLService(mockRepo, cfg, logger, auditManager)
	handler := NewHandler(urlService, cfg, logger)

	ctx := context.Background()
	userID := uuid.New()
	ctx = authcontext.SetUserID(ctx, &userID)

	originalURL := "https://example.com"
	mockRepo.On("SaveIfNotExists", ctx, mock.Anything).Return(nil)

	req := proto.URLShortenRequest_builder{
		Url: originalURL,
	}.Build()

	resp, err := handler.ShortenURL(ctx, req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.GetResult())
	assert.Contains(t, resp.GetResult(), "http://localhost:8080/")
	mockRepo.AssertExpectations(t)
}

func TestShortenURL_EmptyURL(t *testing.T) {
	mockRepo := new(MockURLRepository)
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)
	urlService := service.NewURLService(mockRepo, cfg, logger, auditManager)
	handler := NewHandler(urlService, cfg, logger)

	ctx := context.Background()
	userID := uuid.New()
	ctx = authcontext.SetUserID(ctx, &userID)

	req := proto.URLShortenRequest_builder{
		Url: "",
	}.Build()

	resp, err := handler.ShortenURL(ctx, req)

	require.Error(t, err)
	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Equal(t, "URL is required", st.Message())
}

func TestShortenURL_InvalidURLFormat(t *testing.T) {
	mockRepo := new(MockURLRepository)
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)
	urlService := service.NewURLService(mockRepo, cfg, logger, auditManager)
	handler := NewHandler(urlService, cfg, logger)

	ctx := context.Background()
	userID := uuid.New()
	ctx = authcontext.SetUserID(ctx, &userID)

	tests := []struct {
		name string
		url  string
	}{
		{
			name: "no scheme",
			url:  "example.com",
		},
		{
			name: "ftp scheme",
			url:  "ftp://example.com",
		},
		{
			name: "file scheme",
			url:  "file:///path/to/file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := proto.URLShortenRequest_builder{
				Url: tt.url,
			}.Build()

			resp, err := handler.ShortenURL(ctx, req)

			require.Error(t, err)
			assert.Nil(t, resp)
			st, ok := status.FromError(err)
			require.True(t, ok)
			assert.Equal(t, codes.InvalidArgument, st.Code())
			assert.Equal(t, "Invalid URL format", st.Message())
		})
	}
}

func TestShortenURL_MissingUserID(t *testing.T) {
	mockRepo := new(MockURLRepository)
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)
	urlService := service.NewURLService(mockRepo, cfg, logger, auditManager)
	handler := NewHandler(urlService, cfg, logger)

	ctx := context.Background()

	req := proto.URLShortenRequest_builder{
		Url: "https://example.com",
	}.Build()

	resp, err := handler.ShortenURL(ctx, req)

	require.Error(t, err)
	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
	assert.Equal(t, "Failed to get user ID", st.Message())
}

func TestExpandURL_Success(t *testing.T) {
	mockRepo := new(MockURLRepository)
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)
	urlService := service.NewURLService(mockRepo, cfg, logger, auditManager)
	handler := NewHandler(urlService, cfg, logger)

	ctx := context.Background()
	shortID := "abc123"
	expectedURL := domainurl.URL{
		ShortURL:    shortID,
		OriginalURL: "https://example.com",
		IsDeleted:   false,
	}

	mockRepo.On("Get", ctx, shortID).Return(expectedURL, true)

	req := proto.URLExpandRequest_builder{
		Id: shortID,
	}.Build()

	resp, err := handler.ExpandURL(ctx, req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, expectedURL.OriginalURL, resp.GetResult())
	mockRepo.AssertExpectations(t)
}

func TestExpandURL_EmptyID(t *testing.T) {
	mockRepo := new(MockURLRepository)
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)
	urlService := service.NewURLService(mockRepo, cfg, logger, auditManager)
	handler := NewHandler(urlService, cfg, logger)

	ctx := context.Background()

	req := proto.URLExpandRequest_builder{
		Id: "",
	}.Build()

	resp, err := handler.ExpandURL(ctx, req)

	require.Error(t, err)
	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Equal(t, "ID is required", st.Message())
}

func TestExpandURL_NotFound(t *testing.T) {
	mockRepo := new(MockURLRepository)
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)
	urlService := service.NewURLService(mockRepo, cfg, logger, auditManager)
	handler := NewHandler(urlService, cfg, logger)

	ctx := context.Background()
	shortID := "nonexistent"

	mockRepo.On("Get", ctx, shortID).Return(domainurl.URL{}, false)

	req := proto.URLExpandRequest_builder{
		Id: shortID,
	}.Build()

	resp, err := handler.ExpandURL(ctx, req)

	require.Error(t, err)
	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.NotFound, st.Code())
	assert.Equal(t, "URL not found", st.Message())
	mockRepo.AssertExpectations(t)
}

func TestExpandURL_Deleted(t *testing.T) {
	mockRepo := new(MockURLRepository)
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)
	urlService := service.NewURLService(mockRepo, cfg, logger, auditManager)
	handler := NewHandler(urlService, cfg, logger)

	ctx := context.Background()
	shortID := "abc123"
	deletedURL := domainurl.URL{
		ShortURL:    shortID,
		OriginalURL: "https://example.com",
		IsDeleted:   true,
	}

	mockRepo.On("Get", ctx, shortID).Return(deletedURL, true)

	req := proto.URLExpandRequest_builder{
		Id: shortID,
	}.Build()

	resp, err := handler.ExpandURL(ctx, req)

	require.Error(t, err)
	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.NotFound, st.Code())
	assert.Equal(t, "URL is deleted", st.Message())
	mockRepo.AssertExpectations(t)
}

func TestListUserURLs_Success(t *testing.T) {
	mockRepo := new(MockURLRepository)
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)
	urlService := service.NewURLService(mockRepo, cfg, logger, auditManager)
	handler := NewHandler(urlService, cfg, logger)

	ctx := context.Background()
	userID := uuid.New()
	ctx = authcontext.SetUserID(ctx, &userID)

	expectedURLs := []domainurl.URL{
		{ShortURL: "abc123", OriginalURL: "https://example1.com"},
		{ShortURL: "def456", OriginalURL: "https://example2.com"},
	}

	mockRepo.On("GetAllByUserID", ctx, &userID).Return(expectedURLs, nil)

	req := &emptypb.Empty{}

	resp, err := handler.ListUserURLs(ctx, req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.GetUrl(), 2)
	assert.Equal(t, "http://localhost:8080/abc123", resp.GetUrl()[0].GetShortUrl())
	assert.Equal(t, "https://example1.com", resp.GetUrl()[0].GetOriginalUrl())
	assert.Equal(t, "http://localhost:8080/def456", resp.GetUrl()[1].GetShortUrl())
	assert.Equal(t, "https://example2.com", resp.GetUrl()[1].GetOriginalUrl())
	mockRepo.AssertExpectations(t)
}

func TestListUserURLs_EmptyList(t *testing.T) {
	mockRepo := new(MockURLRepository)
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)
	urlService := service.NewURLService(mockRepo, cfg, logger, auditManager)
	handler := NewHandler(urlService, cfg, logger)

	ctx := context.Background()
	userID := uuid.New()
	ctx = authcontext.SetUserID(ctx, &userID)

	mockRepo.On("GetAllByUserID", ctx, &userID).Return([]domainurl.URL{}, nil)

	req := &emptypb.Empty{}

	resp, err := handler.ListUserURLs(ctx, req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.GetUrl(), 0)
	mockRepo.AssertExpectations(t)
}

func TestListUserURLs_MissingUserID(t *testing.T) {
	mockRepo := new(MockURLRepository)
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)
	urlService := service.NewURLService(mockRepo, cfg, logger, auditManager)
	handler := NewHandler(urlService, cfg, logger)

	ctx := context.Background()

	req := &emptypb.Empty{}

	resp, err := handler.ListUserURLs(ctx, req)

	require.Error(t, err)
	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
	assert.Equal(t, "Failed to get user ID", st.Message())
}

func TestListUserURLs_ServiceError(t *testing.T) {
	mockRepo := new(MockURLRepository)
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)
	urlService := service.NewURLService(mockRepo, cfg, logger, auditManager)
	handler := NewHandler(urlService, cfg, logger)

	ctx := context.Background()
	userID := uuid.New()
	ctx = authcontext.SetUserID(ctx, &userID)

	mockRepo.On("GetAllByUserID", ctx, &userID).Return(nil, errors.New("service error"))

	req := &emptypb.Empty{}

	resp, err := handler.ListUserURLs(ctx, req)

	require.Error(t, err)
	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
	assert.Equal(t, "Failed to get user URLs", st.Message())
	mockRepo.AssertExpectations(t)
}

func TestShortenURL_HTTPScheme(t *testing.T) {
	mockRepo := new(MockURLRepository)
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)
	urlService := service.NewURLService(mockRepo, cfg, logger, auditManager)
	handler := NewHandler(urlService, cfg, logger)

	ctx := context.Background()
	userID := uuid.New()
	ctx = authcontext.SetUserID(ctx, &userID)

	originalURL := "http://example.com"
	mockRepo.On("SaveIfNotExists", ctx, mock.Anything).Return(nil)

	req := proto.URLShortenRequest_builder{
		Url: originalURL,
	}.Build()

	resp, err := handler.ShortenURL(ctx, req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.GetResult())
	assert.Contains(t, resp.GetResult(), "http://localhost:8080/")
	mockRepo.AssertExpectations(t)
}

func TestShortenURL_HTTPSScheme(t *testing.T) {
	mockRepo := new(MockURLRepository)
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)
	urlService := service.NewURLService(mockRepo, cfg, logger, auditManager)
	handler := NewHandler(urlService, cfg, logger)

	ctx := context.Background()
	userID := uuid.New()
	ctx = authcontext.SetUserID(ctx, &userID)

	originalURL := "https://example.com"
	mockRepo.On("SaveIfNotExists", ctx, mock.Anything).Return(nil)

	req := proto.URLShortenRequest_builder{
		Url: originalURL,
	}.Build()

	resp, err := handler.ShortenURL(ctx, req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.GetResult())
	assert.Contains(t, resp.GetResult(), "http://localhost:8080/")
	mockRepo.AssertExpectations(t)
}
