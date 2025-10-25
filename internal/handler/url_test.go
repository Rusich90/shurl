package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreateShortURL(t *testing.T) {
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
			req := httptest.NewRequest(tt.method, "/", strings.NewReader(tt.body))
			w := httptest.NewRecorder()

			CreateShortURL(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.checkResponse && w.Code == http.StatusCreated {
				expectedContentType := "text/plain"
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
	body := "https://example.com"
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	CreateShortURL(w, req)

	shortURL := w.Body.String()
	id := shortURL[strings.LastIndex(shortURL, "/")+1:]

	tests := []struct {
		name             string
		method           string
		path             string
		expectedStatus   int
		expectedLocation string
	}{
		{
			name:             "successful retrieval",
			method:           http.MethodGet,
			path:             "/" + id,
			expectedStatus:   http.StatusTemporaryRedirect,
			expectedLocation: "https://example.com",
		},
		{
			name:             "wrong method",
			method:           http.MethodPost,
			path:             "/" + id,
			expectedStatus:   http.StatusMethodNotAllowed,
			expectedLocation: "",
		},
		{
			name:             "empty id",
			method:           http.MethodGet,
			path:             "/",
			expectedStatus:   http.StatusBadRequest,
			expectedLocation: "",
		},
		{
			name:             "non-existent id",
			method:           http.MethodGet,
			path:             "/nonexistent",
			expectedStatus:   http.StatusNotFound,
			expectedLocation: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			GetOriginalURL(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
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
