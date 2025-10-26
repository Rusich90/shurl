package handler

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Rusich90/shurl.git/internal/config"
	"github.com/Rusich90/shurl.git/internal/repository"
	"github.com/Rusich90/shurl.git/internal/service"
)

var (
	store      = repository.NewURLStore()
	urlService = service.NewURLService(store)
)

func CreateShortURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	url := strings.TrimSpace(string(body))
	if url == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}

	id := urlService.CreateShortURL(url)
	shortURL := fmt.Sprintf("%s/%s", config.GetConfig().BaseURL, id)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}

func GetOriginalURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Path[1:]
	if id == "" {
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}

	url, exists := urlService.GetOriginalURL(id)
	if !exists {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
