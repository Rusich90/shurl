package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Rusich90/shurl.git/internal/config"
	"github.com/Rusich90/shurl.git/internal/repository"
	"github.com/Rusich90/shurl.git/internal/service"
	"github.com/gin-gonic/gin"
)

func TestCreateShortURL(t *testing.T) {
	// Set up dependencies for tests
	cfg := &config.Config{
		ServerAddress: "localhost:8080",
		BaseURL:       "http://localhost:8080",
	}

	store := repository.NewURLStore()
	urlService := service.NewURLService(store, cfg)
	handler := NewHandler(urlService, cfg)

	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		method         string
		body           string
		expectedStatus int
		checkResponse  bool
	}{
		{
			name:           "successful creation",
			method:         http.MethodPost,
			body:           "https://example.com",
			expectedStatus: http.StatusCreated,
			checkResponse:  true,
		},
		{
			name:           "wrong method",
			method:         http.MethodGet,
			body:           "",
			expectedStatus: http.StatusMethodNotAllowed,
			checkResponse:  false,
		},
		{
			name:           "empty url",
			method:         http.MethodPost,
			body:           "",
			expectedStatus: http.StatusBadRequest,
			checkResponse:  false,
		},
		{
			name:           "invalid url format",
			method:         http.MethodPost,
			body:           "invalid-url",
			expectedStatus: http.StatusBadRequest,
			checkResponse:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			var err error
			c.Request, err = http.NewRequest(tt.method, "/", strings.NewReader(tt.body))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			handler.CreateShortURL(c)

			if w.Code != tt.expectedStatus {
				t.Fatalf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.checkResponse && w.Code == http.StatusCreated {
				expectedContentType := "text/plain; charset=utf-8"
				if contentType := w.Header().Get("Content-Type"); contentType != expectedContentType {
					t.Errorf("Expected Content-Type %s, got %s", expectedContentType, contentType)
				}

				responseBody := w.Body.String()
				if responseBody == "" {
					t.Error("Expected non-empty response body")
				}

				if !strings.Contains(responseBody, "http://localhost:8080/") {
					t.Errorf("Expected response to contain short URL, got %s", responseBody)
				}
			}
		})
	}
}

func TestGetOriginalURL(t *testing.T) {
	// Set up dependencies for tests
	cfg := &config.Config{
		ServerAddress: "localhost:8080",
		BaseURL:       "http://localhost:8080",
	}

	store := repository.NewURLStore()
	urlService := service.NewURLService(store, cfg)
	handler := NewHandler(urlService, cfg)

	// Create a short URL first
	w1 := httptest.NewRecorder()
	c1, _ := gin.CreateTestContext(w1)
	var err error
	c1.Request, err = http.NewRequest(http.MethodPost, "/", strings.NewReader("https://example.com"))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	handler.CreateShortURL(c1)

	shortURL := w1.Body.String()
	id := shortURL[strings.LastIndex(shortURL, "/")+1:]

	gin.SetMode(gin.TestMode)

	tests := []struct {
		name             string
		method           string
		path             string
		param            string
		expectedStatus   int
		expectedLocation string
	}{
		{
			name:             "successful retrieval",
			method:           http.MethodGet,
			path:             "/",
			param:            id,
			expectedStatus:   http.StatusTemporaryRedirect,
			expectedLocation: "https://example.com",
		},
		{
			name:             "wrong method",
			method:           http.MethodPost,
			path:             "/",
			param:            id,
			expectedStatus:   http.StatusMethodNotAllowed,
			expectedLocation: "",
		},
		{
			name:             "empty id",
			method:           http.MethodGet,
			path:             "/",
			param:            "",
			expectedStatus:   http.StatusBadRequest,
			expectedLocation: "",
		},
		{
			name:             "non-existent id",
			method:           http.MethodGet,
			path:             "/",
			param:            "nonexistent",
			expectedStatus:   http.StatusNotFound,
			expectedLocation: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			var err error
			c.Request, err = http.NewRequest(tt.method, tt.path, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			c.AddParam("id", tt.param)

			handler.GetOriginalURL(c)

			if w.Code != tt.expectedStatus {
				t.Fatalf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedLocation != "" {
				location := w.Header().Get("Location")
				if location != tt.expectedLocation {
					t.Errorf("Expected Location header to be %s, got %s", tt.expectedLocation, location)
				}
			}
		})
	}
}
