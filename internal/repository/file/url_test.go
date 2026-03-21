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

func TestFileStorage_GetURLs(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_storage_*.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	fileStorage, err := NewFileStorage(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}

	// Пустой файл должен вернуть пустой срез
	urls, err := fileStorage.GetURLs()
	if err != nil {
		t.Fatal(err)
	}
	if len(urls) != 0 {
		t.Errorf("expected empty slice, got %d URLs", len(urls))
	}
}

func TestFileStorage_GetURLs_WithEntries(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_storage_*.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	fileStorage, err := NewFileStorage(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}

	userID := uuid.New()
	rows := []domainurl.URL{
		{
			ShortURL:    "test1",
			OriginalURL: "https://example.com/1",
			UserID:      &userID,
		},
		{
			ShortURL:    "test2",
			OriginalURL: "https://example.com/2",
			UserID:      &userID,
		},
		{
			ShortURL:    "test3",
			OriginalURL: "https://example.com/3",
			UserID:      nil,
		},
	}

	// Сохраняем записи
	for _, row := range rows {
		err = fileStorage.SaveRow(row)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Читаем записи
	urls, err := fileStorage.GetURLs()
	if err != nil {
		t.Fatal(err)
	}

	if len(urls) != len(rows) {
		t.Errorf("expected %d URLs, got %d", len(rows), len(urls))
	}

	// Проверяем содержимое
	for i, url := range urls {
		if url.ShortURL != rows[i].ShortURL {
			t.Errorf("expected ShortURL %s, got %s", rows[i].ShortURL, url.ShortURL)
		}
		if url.OriginalURL != rows[i].OriginalURL {
			t.Errorf("expected OriginalURL %s, got %s", rows[i].OriginalURL, url.OriginalURL)
		}
		if (url.UserID == nil) != (rows[i].UserID == nil) {
			t.Errorf("expected UserID %v, got %v", rows[i].UserID, url.UserID)
		}
	}
}

func TestFileStorage_GetURLs_InvalidJSON(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_storage_*.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	// Записываем файл с валидной и невалидной строками
	content := `{"short_url":"valid1","original_url":"https://example.com/1"}
invalid json line
{"short_url":"valid2","original_url":"https://example.com/2"}
`
	err = os.WriteFile(tmpFile.Name(), []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}

	fileStorage, err := NewFileStorage(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}

	// Невалидные строки должны быть пропущены
	urls, err := fileStorage.GetURLs()
	if err != nil {
		t.Fatal(err)
	}

	if len(urls) != 2 {
		t.Errorf("expected 2 URLs (invalid lines skipped), got %d", len(urls))
	}
}

func TestFileURLRepository_DeleteURLs(t *testing.T) {
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
	rows := []domainurl.URL{
		{
			ShortURL:    "test1",
			OriginalURL: "https://example.com/1",
			UserID:      &userID,
		},
		{
			ShortURL:    "test2",
			OriginalURL: "https://example.com/2",
			UserID:      &userID,
		},
		{
			ShortURL:    "test3",
			OriginalURL: "https://example.com/3",
			UserID:      nil,
		},
	}

	// Сохраняем записи
	for _, row := range rows {
		err = repo.SaveIfNotExists(ctx, row)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Удаляем одну ссылку
	err = repo.DeleteURLs(ctx, []string{"test1"}, &userID)
	if err != nil {
		t.Fatal(err)
	}

	// Проверяем, что ссылка помечена как удаленная
	url, ok := repo.Get(ctx, "test1")
	if !ok {
		t.Fatal("URL not found")
	}
	if !url.IsDeleted {
		t.Errorf("expected IsDeleted to be true, got false")
	}

	// Проверяем, что GetAllByUserID не возвращает удаленные ссылки и URL с nil UserID
	result, err := repo.GetAllByUserID(ctx, &userID)
	if err != nil {
		t.Fatal(err)
	}
	// test1 удален, test3 имеет nil UserID, поэтому ожидаем только 1 URL (test2)
	if len(result) != 1 {
		t.Errorf("expected 1 URL (deleted one and nil UserID excluded), got %d", len(result))
	}
	for _, u := range result {
		if u.ShortURL == "test1" {
			t.Errorf("deleted URL test1 should not be in results")
		}
		if u.ShortURL != "test2" {
			t.Errorf("expected test2, got %s", u.ShortURL)
		}
	}
}

func TestFileURLRepository_DeleteURLs_Multiple(t *testing.T) {
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
	rows := []domainurl.URL{
		{
			ShortURL:    "test1",
			OriginalURL: "https://example.com/1",
			UserID:      &userID,
		},
		{
			ShortURL:    "test2",
			OriginalURL: "https://example.com/2",
			UserID:      &userID,
		},
		{
			ShortURL:    "test3",
			OriginalURL: "https://example.com/3",
			UserID:      &userID,
		},
	}

	// Сохраняем записи
	for _, row := range rows {
		err = repo.SaveIfNotExists(ctx, row)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Удаляем несколько ссылок
	err = repo.DeleteURLs(ctx, []string{"test1", "test2"}, &userID)
	if err != nil {
		t.Fatal(err)
	}

	// Проверяем, что обе ссылки помечены как удаленные
	for _, id := range []string{"test1", "test2"} {
		url, ok := repo.Get(ctx, id)
		if !ok {
			t.Fatalf("URL %s not found", id)
		}
		if !url.IsDeleted {
			t.Errorf("expected IsDeleted to be true for %s, got false", id)
		}
	}

	// Проверяем, что неудаленная ссылка осталась
	url, ok := repo.Get(ctx, "test3")
	if !ok {
		t.Fatal("URL test3 not found")
	}
	if url.IsDeleted {
		t.Errorf("expected IsDeleted to be false for test3, got true")
	}
}

func TestFileURLRepository_DeleteURLs_NotFound(t *testing.T) {
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
		ShortURL:    "test1",
		OriginalURL: "https://example.com/1",
		UserID:      &userID,
	}

	err = repo.SaveIfNotExists(ctx, row)
	if err != nil {
		t.Fatal(err)
	}

	// Удаляем несуществующую ссылку (не должно быть ошибки)
	err = repo.DeleteURLs(ctx, []string{"nonexistent"}, &userID)
	if err != nil {
		t.Fatal(err)
	}
}

func TestFileURLRepository_GetAllByUserID(t *testing.T) {
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

	// Создаем двух пользователей
	userID1 := uuid.New()
	userID2 := uuid.New()

	rows := []domainurl.URL{
		{
			ShortURL:    "user1_url1",
			OriginalURL: "https://example.com/user1/1",
			UserID:      &userID1,
		},
		{
			ShortURL:    "user1_url2",
			OriginalURL: "https://example.com/user1/2",
			UserID:      &userID1,
		},
		{
			ShortURL:    "user2_url1",
			OriginalURL: "https://example.com/user2/1",
			UserID:      &userID2,
		},
		{
			ShortURL:    "no_user_url",
			OriginalURL: "https://example.com/no_user",
			UserID:      nil,
		},
	}

	// Сохраняем записи
	for _, row := range rows {
		err = repo.SaveIfNotExists(ctx, row)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Получаем URL для первого пользователя
	result, err := repo.GetAllByUserID(ctx, &userID1)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 URLs for user1, got %d", len(result))
	}
	for _, url := range result {
		if url.UserID == nil || *url.UserID != userID1 {
			t.Errorf("expected URL with UserID %v, got %v", userID1, url.UserID)
		}
	}

	// Получаем URL для второго пользователя
	result, err = repo.GetAllByUserID(ctx, &userID2)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 URL for user2, got %d", len(result))
	}
	if result[0].UserID == nil || *result[0].UserID != userID2 {
		t.Errorf("expected URL with UserID %v, got %v", userID2, result[0].UserID)
	}

	// Получаем URL для пользователя без URL
	nonExistentUserID := uuid.New()
	result, err = repo.GetAllByUserID(ctx, &nonExistentUserID)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 0 {
		t.Errorf("expected 0 URLs for non-existent user, got %d", len(result))
	}
}

func TestFileURLRepository_GetAllByUserID_Deleted(t *testing.T) {
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
	rows := []domainurl.URL{
		{
			ShortURL:    "test1",
			OriginalURL: "https://example.com/1",
			UserID:      &userID,
		},
		{
			ShortURL:    "test2",
			OriginalURL: "https://example.com/2",
			UserID:      &userID,
		},
	}

	// Сохраняем записи
	for _, row := range rows {
		err = repo.SaveIfNotExists(ctx, row)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Удаляем одну ссылку
	err = repo.DeleteURLs(ctx, []string{"test1"}, &userID)
	if err != nil {
		t.Fatal(err)
	}

	// GetAllByUserID не должна возвращать удаленные ссылки
	result, err := repo.GetAllByUserID(ctx, &userID)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 URL (deleted one excluded), got %d", len(result))
	}
	if result[0].ShortURL != "test2" {
		t.Errorf("expected test2, got %s", result[0].ShortURL)
	}
}

func TestFileURLRepository_SaveBatch(t *testing.T) {
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
	rows := []domainurl.URL{
		{
			ShortURL:    "batch1",
			OriginalURL: "https://example.com/batch/1",
			UserID:      &userID,
		},
		{
			ShortURL:    "batch2",
			OriginalURL: "https://example.com/batch/2",
			UserID:      &userID,
		},
		{
			ShortURL:    "batch3",
			OriginalURL: "https://example.com/batch/3",
			UserID:      nil,
		},
	}

	// Сохраняем пакет
	err = repo.SaveBatch(ctx, rows)
	if err != nil {
		t.Fatal(err)
	}

	// Проверяем, что все URL сохранены
	for _, expected := range rows {
		url, ok := repo.Get(ctx, expected.ShortURL)
		if !ok {
			t.Fatalf("URL %s not found after SaveBatch", expected.ShortURL)
		}
		if url.OriginalURL != expected.OriginalURL {
			t.Errorf("expected OriginalURL %s for %s, got %s", expected.OriginalURL, expected.ShortURL, url.OriginalURL)
		}
		if (url.UserID == nil) != (expected.UserID == nil) {
			t.Errorf("expected UserID %v for %s, got %v", expected.UserID, expected.ShortURL, url.UserID)
		}
	}

	// Проверяем, что GetAllByUserID возвращает все URL
	result, err := repo.GetAllByUserID(ctx, &userID)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 URLs for user, got %d", len(result))
	}
}

func TestFileURLRepository_SaveBatch_Conflict(t *testing.T) {
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

	// Сохраняем первую запись
	userID := uuid.New()
	row := domainurl.URL{
		ShortURL:    "test1",
		OriginalURL: "https://example.com/1",
		UserID:      &userID,
	}
	err = repo.SaveIfNotExists(ctx, row)
	if err != nil {
		t.Fatal(err)
	}

	// Пытаемся сохранить пакет с конфликтом
	rows := []domainurl.URL{
		{
			ShortURL:    "test1",  // Конфликт!
			OriginalURL: "https://example.com/2",
			UserID:      &userID,
		},
		{
			ShortURL:    "new1",
			OriginalURL: "https://example.com/3",
			UserID:      &userID,
		},
	}

	err = repo.SaveBatch(ctx, rows)
	if err == nil {
		t.Fatal("expected error on conflict")
	}

	// Проверяем, что новая запись не была добавлена
	_, ok := repo.Get(ctx, "new1")
	if ok {
		t.Fatal("new URL should not be added on conflict")
	}
}

func TestFileURLRepository_SaveBatch_Empty(t *testing.T) {
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

	// Пустой пакет должен пройти успешно
	err = repo.SaveBatch(ctx, []domainurl.URL{})
	if err != nil {
		t.Fatal(err)
	}
}

func TestFileURLRepository_SaveBatch_ContextCancelled(t *testing.T) {
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

	// Создаем отменяемый контекст
	ctx, cancel := context.WithCancel(context.Background())

	userID := uuid.New()
	rows := []domainurl.URL{
		{
			ShortURL:    "batch1",
			OriginalURL: "https://example.com/batch/1",
			UserID:      &userID,
		},
		{
			ShortURL:    "batch2",
			OriginalURL: "https://example.com/batch/2",
			UserID:      &userID,
		},
	}

	// Отменяем контекст до вызова SaveBatch
	cancel()

	err = repo.SaveBatch(ctx, rows)
	if err != context.Canceled {
		t.Errorf("expected context.Canceled error, got %v", err)
	}

	// Проверяем, что ни одна запись не была добавлена
	for _, row := range rows {
		_, ok := repo.Get(ctx, row.ShortURL)
		if ok {
			t.Fatalf("URL %s should not be added when context is cancelled", row.ShortURL)
		}
	}
}

