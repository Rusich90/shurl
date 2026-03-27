// Package middleware предоставляет функции-обработчики Gin middleware.
//
// Содержит middleware для логирования, сжатия и аутентификации запросов.
package middleware

import (
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
)

// TrustedSubnetMiddleware проверяет, что IP-адрес клиента из заголовка X-Real-IP
// входит в доверенную подсеть, указанную в формате CIDR.
//
// Если trusted_subnet пустой, доступ запрещён для любого запроса.
// Если IP-адрес не входит в доверенную подсеть, возвращается статус 403 Forbidden.
func TrustedSubnetMiddleware(trustedSubnet string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Если доверенная подсеть не задана, запрещаем доступ
		if trustedSubnet == "" {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		// Получаем IP-адрес из заголовка X-Real-IP
		clientIP := c.GetHeader("X-Real-IP")
		if clientIP == "" {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		// Парсим доверенную подсеть
		_, ipNet, err := net.ParseCIDR(trustedSubnet)
		if err != nil {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		// Парсим IP-адрес клиента
		clientAddr := net.ParseIP(clientIP)
		if clientAddr == nil {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		// Проверяем, что IP-адрес клиента входит в доверенную подсеть
		if !ipNet.Contains(clientAddr) {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		c.Next()
	}
}