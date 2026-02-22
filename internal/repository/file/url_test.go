package file

import (
	"context"
	"os"
	"testing"

	domainurl "github.com/Rusich90/shurl.git/internal/domain/url"
	"github.com/google/uuid"
)

func TestFileURLRepository_Get(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_storage_*.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	fileStorage, err := NewFileStorage(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}

	repo, err := NewFileURLRepository(*fileStorage)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	// Добавляем URL
	userID := uuid.New()
	row := domainurl.URL{
		ShortURL:    "test123",
		OriginalURL: "https://example.com",
		UserID:      &userID,
	}

	err = repo.SaveIfNotExists(ctx, row)
	if err != nil {
		t.Fatal(err)
	}

	// Проверяем Get
	url, ok := repo.Get(ctx, "test123")
	if !ok {
		t.Fatal("URL not found")
	}
	if url.OriginalURL != "https://example.com" {
		t.Errorf("expected https://example.com, got %s", url.OriginalURL)
	}
}

func TestFileURLRepository_SaveIfNotExists(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_storage_*.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	fileStorage, err := NewFileStorage(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}

	repo, err := NewFileURLRepository(*fileStorage)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	userID := uuid.New()
	row := domainurl.URL{
		ShortURL:    "test123",
		OriginalURL: "https://example.com",
		UserID:      &userID,
	}

	err = repo.SaveIfNotExists(ctx, row)
	if err != nil {
		t.Fatal(err)
	}

	// Повторное сохранение должно вернуть ошибку
	err = repo.SaveIfNotExists(ctx, row)
	if err == nil {
		t.Fatal("expected error on duplicate save")
	}
}

func TestFileURLRepository_GetByOriginalURL(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_storage_*.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	fileStorage, err := NewFileStorage(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}

	repo, err := NewFileURLRepository(*fileStorage)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	userID := uuid.New()
	row := domainurl.URL{
		ShortURL:    "test123",
		OriginalURL: "https://example.com",
		UserID:      &userID,
	}

	err = repo.SaveIfNotExists(ctx, row)
	if err != nil {
		t.Fatal(err)
	}

	shortURL, ok := repo.GetByOriginalURL(ctx, "https://example.com")
	if !ok {
		t.Fatal("URL not found by original URL")
	}
	if shortURL != "test123" {
		t.Errorf("expected test123, got %s", shortURL)
	}
}

func TestFileURLRepository_Close(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_storage_*.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	fileStorage, err := NewFileStorage(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}

	repo, err := NewFileURLRepository(*fileStorage)
	if err != nil {
		t.Fatal(err)
	}

	err = repo.Close()
	if err != nil {
		t.Fatal(err)
	}
}

func TestFileURLRepository_Ping(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_storage_*.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	fileStorage, err := NewFileStorage(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}

	repo, err := NewFileURLRepository(*fileStorage)
	if err != nil {
		t.Fatal(err)
	}

	err = repo.Ping(context.Background())
	if err != nil {
		t.Fatal(err)
	}
}

// Benchmarks

func BenchmarkGet(b *testing.B) {
	tmpFile, err := os.CreateTemp("", "test_storage_*.jsonl")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	fileStorage, err := NewFileStorage(tmpFile.Name())
	if err != nil {
		b.Fatal(err)
	}

	repo, err := NewFileURLRepository(*fileStorage)
	if err != nil {
		b.Fatal(err)
	}

	ctx := context.Background()

	userID := uuid.New()
	row := domainurl.URL{
		ShortURL:    "test123",
		OriginalURL: "https://example.com",
		UserID:      &userID,
	}

	err = repo.SaveIfNotExists(ctx, row)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.Get(ctx, "test123")
	}
}

func BenchmarkGet_Parallel(b *testing.B) {
	tmpFile, err := os.CreateTemp("", "test_storage_*.jsonl")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	fileStorage, err := NewFileStorage(tmpFile.Name())
	if err != nil {
		b.Fatal(err)
	}

	repo, err := NewFileURLRepository(*fileStorage)
	if err != nil {
		b.Fatal(err)
	}

	ctx := context.Background()

	userID := uuid.New()
	row := domainurl.URL{
		ShortURL:    "test123",
		OriginalURL: "https://example.com",
		UserID:      &userID,
	}

	err = repo.SaveIfNotExists(ctx, row)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = repo.Get(ctx, "test123")
		}
	})
}

func BenchmarkSaveIfNotExists(b *testing.B) {
	tmpFile, err := os.CreateTemp("", "test_storage_*.jsonl")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	fileStorage, err := NewFileStorage(tmpFile.Name())
	if err != nil {
		b.Fatal(err)
	}

	repo, err := NewFileURLRepository(*fileStorage)
	if err != nil {
		b.Fatal(err)
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		userID := uuid.New()
		row := domainurl.URL{
			ShortURL:    "test" + string(rune('0'+i%10)) + string(rune('a'+i%26)),
			OriginalURL: "https://example" + string(rune('0'+i%10)) + ".com",
			UserID:      &userID,
		}
		_ = repo.SaveIfNotExists(ctx, row)
	}
}

func BenchmarkSaveIfNotExists_Parallel(b *testing.B) {
	tmpFile, err := os.CreateTemp("", "test_storage_*.jsonl")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	fileStorage, err := NewFileStorage(tmpFile.Name())
	if err != nil {
		b.Fatal(err)
	}

	repo, err := NewFileURLRepository(*fileStorage)
	if err != nil {
		b.Fatal(err)
	}

	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			userID := uuid.New()
			row := domainurl.URL{
				ShortURL:    "bench_parallel_" + uuid.New().String()[:8],
				OriginalURL: "https://bench.example.com",
				UserID:      &userID,
			}
			_ = repo.SaveIfNotExists(ctx, row)
		}
	})
}

func BenchmarkGetByOriginalURL(b *testing.B) {
	tmpFile, err := os.CreateTemp("", "test_storage_*.jsonl")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	fileStorage, err := NewFileStorage(tmpFile.Name())
	if err != nil {
		b.Fatal(err)
	}

	repo, err := NewFileURLRepository(*fileStorage)
	if err != nil {
		b.Fatal(err)
	}

	ctx := context.Background()

	userID := uuid.New()
	row := domainurl.URL{
		ShortURL:    "test123",
		OriginalURL: "https://example.com",
		UserID:      &userID,
	}

	err = repo.SaveIfNotExists(ctx, row)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.GetByOriginalURL(ctx, "https://example.com")
	}
}

func BenchmarkGetByOriginalURL_Parallel(b *testing.B) {
	tmpFile, err := os.CreateTemp("", "test_storage_*.jsonl")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	fileStorage, err := NewFileStorage(tmpFile.Name())
	if err != nil {
		b.Fatal(err)
	}

	repo, err := NewFileURLRepository(*fileStorage)
	if err != nil {
		b.Fatal(err)
	}

	ctx := context.Background()

	userID := uuid.New()
	row := domainurl.URL{
		ShortURL:    "test123",
		OriginalURL: "https://example.com",
		UserID:      &userID,
	}

	err = repo.SaveIfNotExists(ctx, row)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = repo.GetByOriginalURL(ctx, "https://example.com")
		}
	})
}

func BenchmarkClose(b *testing.B) {
	for i := 0; i < b.N; i++ {
		tmpFile, err := os.CreateTemp("", "test_storage_*.jsonl")
		if err != nil {
			b.Fatal(err)
		}
		defer os.Remove(tmpFile.Name())

		fileStorage, err := NewFileStorage(tmpFile.Name())
		if err != nil {
			b.Fatal(err)
		}

		repo, err := NewFileURLRepository(*fileStorage)
		if err != nil {
			b.Fatal(err)
		}

		_ = repo.Close()
	}
}

func BenchmarkNewFileURLRepository(b *testing.B) {
	tmpFile, err := os.CreateTemp("", "test_storage_*.jsonl")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	fileStorage, err := NewFileStorage(tmpFile.Name())
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = NewFileURLRepository(*fileStorage)
	}
}