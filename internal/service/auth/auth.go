// Package auth предоставляет функции для аутентификации и авторизации.
//
// Использует JWT-токены для управления сессиями пользователей.
package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims представляет собой утверждения JWT-токена.
//
// Содержит идентификатор пользователя и стандартные JWT-утверждения.
type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	jwt.RegisteredClaims
}

// AuthService предоставляет операции для управления JWT-токенами.
type AuthService struct {
	secret []byte
}

// NewAuthService создает новый AuthService с указанным секретным ключом.
func NewAuthService(secret string) *AuthService {
	return &AuthService{
		secret: []byte(secret),
	}
}

// GenerateToken генерирует JWT-токен для указанного пользователя.
//
// Токен действителен 24 часа и подписывается секретным ключом сервиса.
func (s *AuthService) GenerateToken(userID uuid.UUID) (string, error) {
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenString, err := token.SignedString(s.secret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// ValidateToken проверяет валидность JWT-токена и возвращает утверждения.
//
// Возвращает ошибку при невалидном токене или неудачной проверке подписи.
func (s *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return s.secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("token is invalid")
	}

	return claims, nil
}

// GenerateUserID генерирует новый UUID для пользователя.
//
// Используется для создания идентификаторов при анонимном доступе.
func (s *AuthService) GenerateUserID() uuid.UUID {
	return uuid.New()
}
