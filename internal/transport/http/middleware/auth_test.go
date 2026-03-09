package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Rusich90/shurl.git/internal/service/auth"
	"github.com/Rusich90/shurl.git/internal/transport/http/authcontext"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestAuthMiddleware_NoCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)

	authService := auth.NewAuthService("test_secret")
	logger := zap.NewNop()

	middleware := AuthMiddleware(authService, logger)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/", nil)

	middleware(c)

	assert.Equal(t, http.StatusOK, w.Code, "Middleware should not abort request")

	userID, err := authcontext.GetUserID(c)
	assert.NoError(t, err, "UserID should be set in context")
	assert.NotNil(t, userID, "UserID should not be nil")

	cookie := w.Header().Get("Set-Cookie")
	assert.NotEmpty(t, cookie, "Set-Cookie header should be set")
	assert.Contains(t, cookie, "jwt=", "Cookie should contain jwt token")
}

func TestAuthMiddleware_WithValidCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)

	authService := auth.NewAuthService("test_secret")
	logger := zap.NewNop()

	// Генерируем валидный токен
	userID := authService.GenerateUserID()
	token, err := authService.GenerateToken(userID)
	assert.NoError(t, err)

	middleware := AuthMiddleware(authService, logger)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/", nil)
	c.Request.Header.Set("Cookie", "jwt="+token)

	middleware(c)

	assert.Equal(t, http.StatusOK, w.Code, "Middleware should not abort request")

	// Проверяем, что userID из токена установлен в контекст
	storedUserID, err := authcontext.GetUserID(c)
	assert.NoError(t, err, "UserID should be set in context")
	assert.NotNil(t, storedUserID, "UserID should not be nil")
	assert.Equal(t, userID.String(), storedUserID.String(), "UserID should match token claims")
}

func TestAuthMiddleware_WithInvalidCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)

	authService := auth.NewAuthService("test_secret")
	logger := zap.NewNop()

	middleware := AuthMiddleware(authService, logger)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/", nil)
	c.Request.Header.Set("Cookie", "jwt=invalid_token")

	middleware(c)

	assert.Equal(t, http.StatusOK, w.Code, "Middleware should not abort request")

	// При невалидном токене должен быть установлен nil
	userID, err := authcontext.GetUserID(c)
	assert.NoError(t, err, "UserID should be set in context (even if nil)")
	assert.Nil(t, userID, "UserID should be nil for invalid token")
}

func TestAuthMiddleware_WrongSecret(t *testing.T) {
	gin.SetMode(gin.TestMode)

	authService1 := auth.NewAuthService("test_secret_1")
	authService2 := auth.NewAuthService("test_secret_2")
	logger := zap.NewNop()

	// Генерируем токен с помощью authService1
	userID := authService1.GenerateUserID()
	token, err := authService1.GenerateToken(userID)
	assert.NoError(t, err)

	// Используем authService2 для валидации (другой секрет)
	middleware := AuthMiddleware(authService2, logger)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/", nil)
	c.Request.Header.Set("Cookie", "jwt="+token)

	middleware(c)

	assert.Equal(t, http.StatusOK, w.Code, "Middleware should not abort request")

	// При невалидном токене (другой секрет) должен быть установлен nil
	userIDFromContext, err := authcontext.GetUserID(c)
	assert.NoError(t, err, "UserID should be set in context (even if nil)")
	assert.Nil(t, userIDFromContext, "UserID should be nil for token signed with different secret")
}

func TestAuthMiddleware_CookieWithEmptyValue(t *testing.T) {
	gin.SetMode(gin.TestMode)

	authService := auth.NewAuthService("test_secret")
	logger := zap.NewNop()

	middleware := AuthMiddleware(authService, logger)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/", nil)
	// Пустая кука - это когда кука есть, но значение пустое
	// В этом случае c.Cookie вернет пустую строку и err == nil
	c.Request.Header.Set("Cookie", "jwt=")

	middleware(c)

	assert.Equal(t, http.StatusOK, w.Code, "Middleware should not abort request")

	// При пустом значении куки ValidateToken вернет ошибку, и должен быть установлен nil
	userID, err := authcontext.GetUserID(c)
	assert.NoError(t, err, "UserID should be set in context (even if nil)")
	assert.Nil(t, userID, "UserID should be nil for empty cookie value")
}

func TestAuthMiddleware_MultipleRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)

	authService := auth.NewAuthService("test_secret")
	logger := zap.NewNop()

	middleware := AuthMiddleware(authService, logger)

	// Первый запрос - без куки
	w1 := httptest.NewRecorder()
	c1, _ := gin.CreateTestContext(w1)
	c1.Request, _ = http.NewRequest(http.MethodGet, "/", nil)

	middleware(c1)

	userID1, err := authcontext.GetUserID(c1)
	assert.NoError(t, err)
	assert.NotNil(t, userID1)

	// Второй запрос - с той же куки
	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	c2.Request, _ = http.NewRequest(http.MethodGet, "/", nil)

	// Получаем токен из первого запроса
	cookie := w1.Header().Get("Set-Cookie")
	// Извлекаем значение токена из Set-Cookie
	token := extractTokenFromCookie(cookie)
	c2.Request.Header.Set("Cookie", "jwt="+token)

	middleware(c2)

	userID2, err := authcontext.GetUserID(c2)
	assert.NoError(t, err)
	assert.NotNil(t, userID2)
	assert.Equal(t, userID1.String(), userID2.String(), "UserID should be the same for same token")
}

func extractTokenFromCookie(cookieHeader string) string {
	// Простая функция для извлечения токена из Set-Cookie
	// Формат: "jwt=<token>; Path=/; HttpOnly"
	parts := strings.Split(cookieHeader, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if len(part) > 4 && part[:4] == "jwt=" {
			return part[4:]
		}
	}
	return ""
}