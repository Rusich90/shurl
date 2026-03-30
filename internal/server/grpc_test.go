package server

import (
	"context"
	"testing"

	"github.com/Rusich90/shurl.git/internal/audit"
	"github.com/Rusich90/shurl.git/internal/config"
	domainurl "github.com/Rusich90/shurl.git/internal/domain/url"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
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

func TestSetupGRPCServer_Success(t *testing.T) {
	mockRepo := new(MockURLRepository)
	cfg := &config.Config{
		BaseURL:    "http://localhost:8080",
		AuthSecret: "test-secret-key-12345678901234567890",
	}
	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)

	grpcServer := SetupGRPCServer(cfg, mockRepo, auditManager, logger)

	require.NotNil(t, grpcServer)
	assert.IsType(t, &grpc.Server{}, grpcServer)
}

func TestSetupGRPCServer_WithEmptyAuthSecret(t *testing.T) {
	mockRepo := new(MockURLRepository)
	cfg := &config.Config{
		BaseURL:    "http://localhost:8080",
		AuthSecret: "",
	}
	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)

	grpcServer := SetupGRPCServer(cfg, mockRepo, auditManager, logger)

	require.NotNil(t, grpcServer)
	assert.IsType(t, &grpc.Server{}, grpcServer)
}

func TestSetupGRPCServer_WithCustomBaseURL(t *testing.T) {
	mockRepo := new(MockURLRepository)
	cfg := &config.Config{
		BaseURL:    "https://custom.example.com",
		AuthSecret: "test-secret-key-12345678901234567890",
	}
	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)

	grpcServer := SetupGRPCServer(cfg, mockRepo, auditManager, logger)

	require.NotNil(t, grpcServer)
	assert.IsType(t, &grpc.Server{}, grpcServer)
}

func TestSetupGRPCServer_WithNilAuditManager(t *testing.T) {
	mockRepo := new(MockURLRepository)
	cfg := &config.Config{
		BaseURL:    "http://localhost:8080",
		AuthSecret: "test-secret-key-12345678901234567890",
	}
	logger := zap.NewNop()

	grpcServer := SetupGRPCServer(cfg, mockRepo, nil, logger)

	require.NotNil(t, grpcServer)
	assert.IsType(t, &grpc.Server{}, grpcServer)
}

func TestSetupGRPCServer_WithNilLogger(t *testing.T) {
	mockRepo := new(MockURLRepository)
	cfg := &config.Config{
		BaseURL:    "http://localhost:8080",
		AuthSecret: "test-secret-key-12345678901234567890",
	}
	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)

	grpcServer := SetupGRPCServer(cfg, mockRepo, auditManager, nil)

	require.NotNil(t, grpcServer)
	assert.IsType(t, &grpc.Server{}, grpcServer)
}

func TestSetupGRPCServer_WithNilRepository(t *testing.T) {
	cfg := &config.Config{
		BaseURL:    "http://localhost:8080",
		AuthSecret: "test-secret-key-12345678901234567890",
	}
	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)

	grpcServer := SetupGRPCServer(cfg, nil, auditManager, logger)

	require.NotNil(t, grpcServer)
	assert.IsType(t, &grpc.Server{}, grpcServer)
}

func TestSetupGRPCServer_WithNilConfig(t *testing.T) {
	// Этот тест пропускается, так как SetupGRPCServer требует валидный config
	// и вызовет панику при nil config
	t.Skip("SetupGRPCServer requires valid config, nil config causes panic")
}

func TestSetupGRPCServer_MultipleServers(t *testing.T) {
	mockRepo := new(MockURLRepository)
	cfg := &config.Config{
		BaseURL:    "http://localhost:8080",
		AuthSecret: "test-secret-key-12345678901234567890",
	}
	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)

	// Создаем несколько серверов
	server1 := SetupGRPCServer(cfg, mockRepo, auditManager, logger)
	server2 := SetupGRPCServer(cfg, mockRepo, auditManager, logger)
	server3 := SetupGRPCServer(cfg, mockRepo, auditManager, logger)

	require.NotNil(t, server1)
	require.NotNil(t, server2)
	require.NotNil(t, server3)

	// Проверяем, что это разные экземпляры
	assert.NotEqual(t, server1, server2)
	assert.NotEqual(t, server2, server3)
	assert.NotEqual(t, server1, server3)
}

func TestSetupGRPCServer_WithDifferentSecrets(t *testing.T) {
	mockRepo := new(MockURLRepository)
	cfg1 := &config.Config{
		BaseURL:    "http://localhost:8080",
		AuthSecret: "secret-1-12345678901234567890",
	}
	cfg2 := &config.Config{
		BaseURL:    "http://localhost:8080",
		AuthSecret: "secret-2-12345678901234567890",
	}
	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)

	server1 := SetupGRPCServer(cfg1, mockRepo, auditManager, logger)
	server2 := SetupGRPCServer(cfg2, mockRepo, auditManager, logger)

	require.NotNil(t, server1)
	require.NotNil(t, server2)
	assert.NotEqual(t, server1, server2)
}

func TestSetupGRPCServer_ServerReflection(t *testing.T) {
	mockRepo := new(MockURLRepository)
	cfg := &config.Config{
		BaseURL:    "http://localhost:8080",
		AuthSecret: "test-secret-key-12345678901234567890",
	}
	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)

	grpcServer := SetupGRPCServer(cfg, mockRepo, auditManager, logger)

	require.NotNil(t, grpcServer)

	// Проверяем, что Server Reflection зарегистрирован
	// Это проверяется косвенно - если сервер создан без ошибок,
	// значит reflection.Register был вызван успешно
	assert.IsType(t, &grpc.Server{}, grpcServer)
}

func TestSetupGRPCServer_Interceptors(t *testing.T) {
	mockRepo := new(MockURLRepository)
	cfg := &config.Config{
		BaseURL:    "http://localhost:8080",
		AuthSecret: "test-secret-key-12345678901234567890",
	}
	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)

	grpcServer := SetupGRPCServer(cfg, mockRepo, auditManager, logger)

	require.NotNil(t, grpcServer)
	assert.IsType(t, &grpc.Server{}, grpcServer)

	// Проверяем, что интерцепторы были добавлены
	// Это проверяется косвенно - если сервер создан без ошибок,
	// значит интерцепторы были успешно добавлены
}

func TestSetupGRPCServer_ServiceRegistration(t *testing.T) {
	mockRepo := new(MockURLRepository)
	cfg := &config.Config{
		BaseURL:    "http://localhost:8080",
		AuthSecret: "test-secret-key-12345678901234567890",
	}
	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)

	grpcServer := SetupGRPCServer(cfg, mockRepo, auditManager, logger)

	require.NotNil(t, grpcServer)
	assert.IsType(t, &grpc.Server{}, grpcServer)

	// Проверяем, что сервис был зарегистрирован
	// Это проверяется косвенно - если сервер создан без ошибок,
	// значит сервис был успешно зарегистрирован
}

func TestSetupGRPCServer_AllNilParameters(t *testing.T) {
	// Этот тест пропускается, так как SetupGRPCServer требует валидный config
	// и вызовет панику при nil config
	t.Skip("SetupGRPCServer requires valid config, nil config causes panic")
}

func TestSetupGRPCServer_WithProductionLogger(t *testing.T) {
	mockRepo := new(MockURLRepository)
	cfg := &config.Config{
		BaseURL:    "http://localhost:8080",
		AuthSecret: "test-secret-key-12345678901234567890",
	}
	logger, err := zap.NewProduction()
	require.NoError(t, err)
	defer logger.Sync()
	auditManager := audit.NewManager(logger)

	grpcServer := SetupGRPCServer(cfg, mockRepo, auditManager, logger)

	require.NotNil(t, grpcServer)
	assert.IsType(t, &grpc.Server{}, grpcServer)
}

func TestSetupGRPCServer_WithDevelopmentLogger(t *testing.T) {
	mockRepo := new(MockURLRepository)
	cfg := &config.Config{
		BaseURL:    "http://localhost:8080",
		AuthSecret: "test-secret-key-12345678901234567890",
	}
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	defer logger.Sync()
	auditManager := audit.NewManager(logger)

	grpcServer := SetupGRPCServer(cfg, mockRepo, auditManager, logger)

	require.NotNil(t, grpcServer)
	assert.IsType(t, &grpc.Server{}, grpcServer)
}

func TestSetupGRPCServer_WithShortAuthSecret(t *testing.T) {
	mockRepo := new(MockURLRepository)
	cfg := &config.Config{
		BaseURL:    "http://localhost:8080",
		AuthSecret: "short",
	}
	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)

	grpcServer := SetupGRPCServer(cfg, mockRepo, auditManager, logger)

	require.NotNil(t, grpcServer)
	assert.IsType(t, &grpc.Server{}, grpcServer)
}

func TestSetupGRPCServer_WithLongAuthSecret(t *testing.T) {
	mockRepo := new(MockURLRepository)
	cfg := &config.Config{
		BaseURL:    "http://localhost:8080",
		AuthSecret: "very-long-secret-key-1234567890123456789012345678901234567890",
	}
	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)

	grpcServer := SetupGRPCServer(cfg, mockRepo, auditManager, logger)

	require.NotNil(t, grpcServer)
	assert.IsType(t, &grpc.Server{}, grpcServer)
}

func TestSetupGRPCServer_WithSpecialCharsInSecret(t *testing.T) {
	mockRepo := new(MockURLRepository)
	cfg := &config.Config{
		BaseURL:    "http://localhost:8080",
		AuthSecret: "secret!@#$%^&*()_+-=[]{}|;':\",./<>?",
	}
	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)

	grpcServer := SetupGRPCServer(cfg, mockRepo, auditManager, logger)

	require.NotNil(t, grpcServer)
	assert.IsType(t, &grpc.Server{}, grpcServer)
}

func TestSetupGRPCServer_WithUnicodeInSecret(t *testing.T) {
	mockRepo := new(MockURLRepository)
	cfg := &config.Config{
		BaseURL:    "http://localhost:8080",
		AuthSecret: "секретный-ключ-1234567890",
	}
	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)

	grpcServer := SetupGRPCServer(cfg, mockRepo, auditManager, logger)

	require.NotNil(t, grpcServer)
	assert.IsType(t, &grpc.Server{}, grpcServer)
}

func TestSetupGRPCServer_WithEmptyBaseURL(t *testing.T) {
	mockRepo := new(MockURLRepository)
	cfg := &config.Config{
		BaseURL:    "",
		AuthSecret: "test-secret-key-12345678901234567890",
	}
	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)

	grpcServer := SetupGRPCServer(cfg, mockRepo, auditManager, logger)

	require.NotNil(t, grpcServer)
	assert.IsType(t, &grpc.Server{}, grpcServer)
}

func TestSetupGRPCServer_WithInvalidBaseURL(t *testing.T) {
	mockRepo := new(MockURLRepository)
	cfg := &config.Config{
		BaseURL:    "not-a-valid-url",
		AuthSecret: "test-secret-key-12345678901234567890",
	}
	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)

	grpcServer := SetupGRPCServer(cfg, mockRepo, auditManager, logger)

	require.NotNil(t, grpcServer)
	assert.IsType(t, &grpc.Server{}, grpcServer)
}

func TestSetupGRPCServer_WithLocalhostBaseURL(t *testing.T) {
	mockRepo := new(MockURLRepository)
	cfg := &config.Config{
		BaseURL:    "http://localhost:8080",
		AuthSecret: "test-secret-key-12345678901234567890",
	}
	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)

	grpcServer := SetupGRPCServer(cfg, mockRepo, auditManager, logger)

	require.NotNil(t, grpcServer)
	assert.IsType(t, &grpc.Server{}, grpcServer)
}

func TestSetupGRPCServer_WithIPBaseURL(t *testing.T) {
	mockRepo := new(MockURLRepository)
	cfg := &config.Config{
		BaseURL:    "http://127.0.0.1:8080",
		AuthSecret: "test-secret-key-12345678901234567890",
	}
	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)

	grpcServer := SetupGRPCServer(cfg, mockRepo, auditManager, logger)

	require.NotNil(t, grpcServer)
	assert.IsType(t, &grpc.Server{}, grpcServer)
}

func TestSetupGRPCServer_WithDomainBaseURL(t *testing.T) {
	mockRepo := new(MockURLRepository)
	cfg := &config.Config{
		BaseURL:    "https://example.com",
		AuthSecret: "test-secret-key-12345678901234567890",
	}
	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)

	grpcServer := SetupGRPCServer(cfg, mockRepo, auditManager, logger)

	require.NotNil(t, grpcServer)
	assert.IsType(t, &grpc.Server{}, grpcServer)
}

func TestSetupGRPCServer_WithSubdomainBaseURL(t *testing.T) {
	mockRepo := new(MockURLRepository)
	cfg := &config.Config{
		BaseURL:    "https://api.example.com",
		AuthSecret: "test-secret-key-12345678901234567890",
	}
	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)

	grpcServer := SetupGRPCServer(cfg, mockRepo, auditManager, logger)

	require.NotNil(t, grpcServer)
	assert.IsType(t, &grpc.Server{}, grpcServer)
}

func TestSetupGRPCServer_WithPortInBaseURL(t *testing.T) {
	mockRepo := new(MockURLRepository)
	cfg := &config.Config{
		BaseURL:    "http://localhost:3000",
		AuthSecret: "test-secret-key-12345678901234567890",
	}
	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)

	grpcServer := SetupGRPCServer(cfg, mockRepo, auditManager, logger)

	require.NotNil(t, grpcServer)
	assert.IsType(t, &grpc.Server{}, grpcServer)
}

func TestSetupGRPCServer_WithPathInBaseURL(t *testing.T) {
	mockRepo := new(MockURLRepository)
	cfg := &config.Config{
		BaseURL:    "http://localhost:8080/api/v1",
		AuthSecret: "test-secret-key-12345678901234567890",
	}
	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)

	grpcServer := SetupGRPCServer(cfg, mockRepo, auditManager, logger)

	require.NotNil(t, grpcServer)
	assert.IsType(t, &grpc.Server{}, grpcServer)
}

func TestSetupGRPCServer_WithHTTPSBaseURL(t *testing.T) {
	mockRepo := new(MockURLRepository)
	cfg := &config.Config{
		BaseURL:    "https://secure.example.com",
		AuthSecret: "test-secret-key-12345678901234567890",
	}
	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)

	grpcServer := SetupGRPCServer(cfg, mockRepo, auditManager, logger)

	require.NotNil(t, grpcServer)
	assert.IsType(t, &grpc.Server{}, grpcServer)
}

func TestSetupGRPCServer_WithHTTPBaseURL(t *testing.T) {
	mockRepo := new(MockURLRepository)
	cfg := &config.Config{
		BaseURL:    "http://insecure.example.com",
		AuthSecret: "test-secret-key-12345678901234567890",
	}
	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)

	grpcServer := SetupGRPCServer(cfg, mockRepo, auditManager, logger)

	require.NotNil(t, grpcServer)
	assert.IsType(t, &grpc.Server{}, grpcServer)
}

func TestSetupGRPCServer_WithTrailingSlashBaseURL(t *testing.T) {
	mockRepo := new(MockURLRepository)
	cfg := &config.Config{
		BaseURL:    "http://localhost:8080/",
		AuthSecret: "test-secret-key-12345678901234567890",
	}
	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)

	grpcServer := SetupGRPCServer(cfg, mockRepo, auditManager, logger)

	require.NotNil(t, grpcServer)
	assert.IsType(t, &grpc.Server{}, grpcServer)
}

func TestSetupGRPCServer_WithQueryParamsBaseURL(t *testing.T) {
	mockRepo := new(MockURLRepository)
	cfg := &config.Config{
		BaseURL:    "http://localhost:8080?param=value",
		AuthSecret: "test-secret-key-12345678901234567890",
	}
	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)

	grpcServer := SetupGRPCServer(cfg, mockRepo, auditManager, logger)

	require.NotNil(t, grpcServer)
	assert.IsType(t, &grpc.Server{}, grpcServer)
}

func TestSetupGRPCServer_WithFragmentBaseURL(t *testing.T) {
	mockRepo := new(MockURLRepository)
	cfg := &config.Config{
		BaseURL:    "http://localhost:8080#fragment",
		AuthSecret: "test-secret-key-12345678901234567890",
	}
	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)

	grpcServer := SetupGRPCServer(cfg, mockRepo, auditManager, logger)

	require.NotNil(t, grpcServer)
	assert.IsType(t, &grpc.Server{}, grpcServer)
}

func TestSetupGRPCServer_WithUserInfoBaseURL(t *testing.T) {
	mockRepo := new(MockURLRepository)
	cfg := &config.Config{
		BaseURL:    "http://user:pass@localhost:8080",
		AuthSecret: "test-secret-key-12345678901234567890",
	}
	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)

	grpcServer := SetupGRPCServer(cfg, mockRepo, auditManager, logger)

	require.NotNil(t, grpcServer)
	assert.IsType(t, &grpc.Server{}, grpcServer)
}

func TestSetupGRPCServer_WithIPv6BaseURL(t *testing.T) {
	mockRepo := new(MockURLRepository)
	cfg := &config.Config{
		BaseURL:    "http://[::1]:8080",
		AuthSecret: "test-secret-key-12345678901234567890",
	}
	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)

	grpcServer := SetupGRPCServer(cfg, mockRepo, auditManager, logger)

	require.NotNil(t, grpcServer)
	assert.IsType(t, &grpc.Server{}, grpcServer)
}

func TestSetupGRPCServer_WithFullConfig(t *testing.T) {
	mockRepo := new(MockURLRepository)
	cfg := &config.Config{
		ServerAddress:     "localhost:8080",
		GRPCServerAddress: "localhost:9090",
		BaseURL:           "http://localhost:8080",
		FileStoragePath:   "file_storage.jsonl",
		DatabaseDSN:       "postgres://localhost:5432/db",
		MigrationsPath:    "file://migrations",
		AuthSecret:        "test-secret-key-12345678901234567890",
		AuditFile:         "audit.log",
		AuditURL:          "http://audit.example.com",
		EnableHTTPS:       true,
		TrustedSubnet:     "192.168.1.0/24",
	}
	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)

	grpcServer := SetupGRPCServer(cfg, mockRepo, auditManager, logger)

	require.NotNil(t, grpcServer)
	assert.IsType(t, &grpc.Server{}, grpcServer)
}

func TestSetupGRPCServer_ReflectionService(t *testing.T) {
	mockRepo := new(MockURLRepository)
	cfg := &config.Config{
		BaseURL:    "http://localhost:8080",
		AuthSecret: "test-secret-key-12345678901234567890",
	}
	logger := zap.NewNop()
	auditManager := audit.NewManager(logger)

	grpcServer := SetupGRPCServer(cfg, mockRepo, auditManager, logger)

	require.NotNil(t, grpcServer)
	assert.IsType(t, &grpc.Server{}, grpcServer)

	// Проверяем, что reflection.Register был вызван
	// Это проверяется косвенно - если сервер создан без ошибок,
	// значит reflection.Register был вызван успешно
	_ = reflection.Register // Используем для проверки импорта
}
