package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestTrustedSubnetMiddleware_EmptyTrustedSubnet(t *testing.T) {
	gin.SetMode(gin.TestMode)

	middleware := TrustedSubnetMiddleware("")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/api/internal/stats", nil)
	c.Request.Header.Set("X-Real-IP", "192.168.1.1")

	middleware(c)

	assert.Equal(t, http.StatusForbidden, w.Code, "Should return 403 when trusted subnet is empty")
	assert.True(t, c.IsAborted(), "Request should be aborted")
}

func TestTrustedSubnetMiddleware_MissingXRealIPHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	middleware := TrustedSubnetMiddleware("192.168.1.0/24")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/api/internal/stats", nil)
	// Не устанавливаем заголовок X-Real-IP

	middleware(c)

	assert.Equal(t, http.StatusForbidden, w.Code, "Should return 403 when X-Real-IP header is missing")
	assert.True(t, c.IsAborted(), "Request should be aborted")
}

func TestTrustedSubnetMiddleware_IPNotInTrustedSubnet(t *testing.T) {
	gin.SetMode(gin.TestMode)

	middleware := TrustedSubnetMiddleware("192.168.1.0/24")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/api/internal/stats", nil)
	c.Request.Header.Set("X-Real-IP", "10.0.0.1")

	middleware(c)

	assert.Equal(t, http.StatusForbidden, w.Code, "Should return 403 when IP is not in trusted subnet")
	assert.True(t, c.IsAborted(), "Request should be aborted")
}

func TestTrustedSubnetMiddleware_IPInTrustedSubnet(t *testing.T) {
	gin.SetMode(gin.TestMode)

	middleware := TrustedSubnetMiddleware("192.168.1.0/24")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/api/internal/stats", nil)
	c.Request.Header.Set("X-Real-IP", "192.168.1.100")

	middleware(c)

	assert.Equal(t, http.StatusOK, w.Code, "Should return 200 when IP is in trusted subnet")
	assert.False(t, c.IsAborted(), "Request should not be aborted")
}

func TestTrustedSubnetMiddleware_InvalidCIDR(t *testing.T) {
	gin.SetMode(gin.TestMode)

	middleware := TrustedSubnetMiddleware("invalid-cidr")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/api/internal/stats", nil)
	c.Request.Header.Set("X-Real-IP", "192.168.1.1")

	middleware(c)

	assert.Equal(t, http.StatusForbidden, w.Code, "Should return 403 when CIDR is invalid")
	assert.True(t, c.IsAborted(), "Request should be aborted")
}

func TestTrustedSubnetMiddleware_InvalidIP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	middleware := TrustedSubnetMiddleware("192.168.1.0/24")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/api/internal/stats", nil)
	c.Request.Header.Set("X-Real-IP", "invalid-ip")

	middleware(c)

	assert.Equal(t, http.StatusForbidden, w.Code, "Should return 403 when IP is invalid")
	assert.True(t, c.IsAborted(), "Request should be aborted")
}

func TestTrustedSubnetMiddleware_IPAtSubnetBoundary(t *testing.T) {
	gin.SetMode(gin.TestMode)

	middleware := TrustedSubnetMiddleware("192.168.1.0/24")

	testCases := []struct {
		name     string
		ip       string
		expected int
	}{
		{"First IP in subnet", "192.168.1.0", http.StatusOK},
		{"Last IP in subnet", "192.168.1.255", http.StatusOK},
		{"IP just before subnet", "192.168.0.255", http.StatusForbidden},
		{"IP just after subnet", "192.168.2.0", http.StatusForbidden},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest(http.MethodGet, "/api/internal/stats", nil)
			c.Request.Header.Set("X-Real-IP", tc.ip)

			middleware(c)

			assert.Equal(t, tc.expected, w.Code, "Should return expected status for IP at boundary")
		})
	}
}

func TestTrustedSubnetMiddleware_DifferentCIDRFormats(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testCases := []struct {
		name          string
		cidr          string
		ip            string
		expected      int
		description   string
	}{
		{
			name:        "IPv4 /24 subnet",
			cidr:        "192.168.1.0/24",
			ip:          "192.168.1.50",
			expected:    http.StatusOK,
			description: "Should allow IP in /24 subnet",
		},
		{
			name:        "IPv4 /16 subnet",
			cidr:        "10.0.0.0/16",
			ip:          "10.0.255.255",
			expected:    http.StatusOK,
			description: "Should allow IP in /16 subnet",
		},
		{
			name:        "IPv4 /8 subnet",
			cidr:        "172.16.0.0/12",
			ip:          "172.31.255.255",
			expected:    http.StatusOK,
			description: "Should allow IP in /12 subnet",
		},
		{
			name:        "IPv4 /32 single IP",
			cidr:        "192.168.1.100/32",
			ip:          "192.168.1.100",
			expected:    http.StatusOK,
			description: "Should allow exact IP match",
		},
		{
			name:        "IPv4 /32 different IP",
			cidr:        "192.168.1.100/32",
			ip:          "192.168.1.101",
			expected:    http.StatusForbidden,
			description: "Should deny different IP for /32",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			middleware := TrustedSubnetMiddleware(tc.cidr)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest(http.MethodGet, "/api/internal/stats", nil)
			c.Request.Header.Set("X-Real-IP", tc.ip)

			middleware(c)

			assert.Equal(t, tc.expected, w.Code, tc.description)
		})
	}
}