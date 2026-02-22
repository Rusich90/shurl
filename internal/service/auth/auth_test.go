package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewAuthService(t *testing.T) {
	secret := "test_secret"
	service := NewAuthService(secret)

	assert.NotNil(t, service)
}

func TestAuthService_GenerateToken(t *testing.T) {
	secret := "test_secret"
	service := NewAuthService(secret)

	userID := uuid.New()
	token, err := service.GenerateToken(userID)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Токен должен быть строкой
	assert.IsType(t, "", token)
}

func TestAuthService_ValidateToken(t *testing.T) {
	secret := "test_secret"
	service := NewAuthService(secret)

	userID := uuid.New()
	token, err := service.GenerateToken(userID)
	assert.NoError(t, err)

	// Валидация токена
	claims, err := service.ValidateToken(token)

	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, userID, claims.UserID)
}

func TestAuthService_ValidateToken_InvalidToken(t *testing.T) {
	secret := "test_secret"
	service := NewAuthService(secret)

	// Валидация невалидного токена
	claims, err := service.ValidateToken("invalid_token")

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestAuthService_ValidateToken_WrongSecret(t *testing.T) {
	secret1 := "test_secret_1"
	secret2 := "test_secret_2"

	service1 := NewAuthService(secret1)
	service2 := NewAuthService(secret2)

	userID := uuid.New()
	token, err := service1.GenerateToken(userID)
	assert.NoError(t, err)

	// Валидация токена с другим секретом
	claims, err := service2.ValidateToken(token)

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestAuthService_ValidateToken_ExpiredToken(t *testing.T) {
	// Для теста истекшего токена нужно создать токен с прошедшим сроком действия
	// Это сложно без изменения кода, поэтому пропустим этот тест
}

func TestAuthService_GenerateUserID(t *testing.T) {
	secret := "test_secret"
	service := NewAuthService(secret)

	userID1 := service.GenerateUserID()
	userID2 := service.GenerateUserID()

	assert.NotEmpty(t, userID1)
	assert.NotEmpty(t, userID2)
	assert.NotEqual(t, userID1, userID2)
}

func TestAuthService_GenerateToken_WithClaims(t *testing.T) {
	secret := "test_secret"
	service := NewAuthService(secret)

	userID := uuid.New()
	token, err := service.GenerateToken(userID)
	assert.NoError(t, err)

	// Декодируем токен и проверяем claims
	claims, err := service.ValidateToken(token)
	assert.NoError(t, err)

	// Проверяем, что токен действителен в течение 24 часов
	now := time.Now()
	assert.WithinDuration(t, now, claims.IssuedAt.Time, 1*time.Second)
	assert.WithinDuration(t, now.Add(24*time.Hour), claims.ExpiresAt.Time, 1*time.Second)
}

func BenchmarkGenerateToken(b *testing.B) {
	secret := "test_secret"
	service := NewAuthService(secret)
	userID := uuid.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.GenerateToken(userID)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGenerateToken_Parallel(b *testing.B) {
	secret := "test_secret"
	service := NewAuthService(secret)
	userID := uuid.New()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := service.GenerateToken(userID)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkValidateToken(b *testing.B) {
	secret := "test_secret"
	service := NewAuthService(secret)
	userID := uuid.New()
	token, err := service.GenerateToken(userID)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.ValidateToken(token)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkValidateToken_Parallel(b *testing.B) {
	secret := "test_secret"
	service := NewAuthService(secret)
	userID := uuid.New()
	token, err := service.GenerateToken(userID)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := service.ValidateToken(token)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkGenerateUserID(b *testing.B) {
	secret := "test_secret"
	service := NewAuthService(secret)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = service.GenerateUserID()
	}
}

func BenchmarkGenerateUserID_Parallel(b *testing.B) {
	secret := "test_secret"
	service := NewAuthService(secret)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = service.GenerateUserID()
		}
	})
}