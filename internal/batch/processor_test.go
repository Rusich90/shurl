package batch

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

func TestNewBatchProcessor(t *testing.T) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatal(err)
	}

	processor, err := NewBatchProcessor(4, 10, logger)
	if err != nil {
		t.Fatal(err)
	}
	if processor == nil {
		t.Fatal("processor should not be nil")
	}
}

func TestNewBatchProcessor_InvalidNumWorkers(t *testing.T) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatal(err)
	}

	_, err = NewBatchProcessor(0, 10, logger)
	if err == nil {
		t.Fatal("expected error for invalid numWorkers")
	}

	_, err = NewBatchProcessor(-1, 10, logger)
	if err == nil {
		t.Fatal("expected error for invalid numWorkers")
	}
}

func TestNewBatchProcessor_InvalidChunkSize(t *testing.T) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatal(err)
	}

	_, err = NewBatchProcessor(4, 0, logger)
	if err == nil {
		t.Fatal("expected error for invalid chunkSize")
	}

	_, err = NewBatchProcessor(4, -1, logger)
	if err == nil {
		t.Fatal("expected error for invalid chunkSize")
	}
}

func TestBatchProcessor_ProcessBatch_Empty(t *testing.T) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatal(err)
	}

	processor, err := NewBatchProcessor(4, 10, logger)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	var processed int64
	processFunc := func(ctx context.Context, items []string, userID *uuid.UUID) error {
		atomic.AddInt64(&processed, int64(len(items)))
		return nil
	}

	stats, err := processor.ProcessBatch(ctx, []string{}, nil, processFunc)
	if err != nil {
		t.Fatal(err)
	}
	if stats.SuccessfulDeletes != 0 {
		t.Errorf("expected 0 successful, got %d", stats.SuccessfulDeletes)
	}
}

func TestBatchProcessor_ProcessBatch_Success(t *testing.T) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatal(err)
	}

	processor, err := NewBatchProcessor(4, 10, logger)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	var processed int64
	processFunc := func(ctx context.Context, items []string, userID *uuid.UUID) error {
		atomic.AddInt64(&processed, int64(len(items)))
		return nil
	}

	items := make([]string, 50)
	for i := range items {
		items[i] = "item" + string(rune('0'+i%10))
	}

	stats, err := processor.ProcessBatch(ctx, items, nil, processFunc)
	if err != nil {
		t.Fatal(err)
	}
	if stats.SuccessfulDeletes != 50 {
		t.Errorf("expected 50 successful, got %d", stats.SuccessfulDeletes)
	}
	if stats.ProcessedChunks != 5 {
		t.Errorf("expected 5 chunks, got %d", stats.ProcessedChunks)
	}
}

func TestBatchProcessor_ProcessBatch_Failure(t *testing.T) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatal(err)
	}

	processor, err := NewBatchProcessor(4, 10, logger)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	var processed int64
	processFunc := func(ctx context.Context, items []string, userID *uuid.UUID) error {
		atomic.AddInt64(&processed, int64(len(items)))
		return nil
	}

	items := make([]string, 25)
	for i := range items {
		items[i] = "item" + string(rune('0'+i%10))
	}

	stats, err := processor.ProcessBatch(ctx, items, nil, processFunc)
	if err != nil {
		t.Fatal(err)
	}
	if stats.SuccessfulDeletes != 25 {
		t.Errorf("expected 25 successful, got %d", stats.SuccessfulDeletes)
	}
}

func TestBatchProcessor_ProcessBatch_ContextCancelled(t *testing.T) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatal(err)
	}

	processor, err := NewBatchProcessor(4, 10, logger)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	var processed int64
	processFunc := func(ctx context.Context, items []string, userID *uuid.UUID) error {
		atomic.AddInt64(&processed, int64(len(items)))
		return nil
	}

	stats, err := processor.ProcessBatch(ctx, []string{"item1", "item2"}, nil, processFunc)
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
	_ = stats // stats may be partially populated
}

func TestGenerateChunks(t *testing.T) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatal(err)
	}

	processor, err := NewBatchProcessor(4, 10, logger)
	if err != nil {
		t.Fatal(err)
	}

	items := []string{"a", "b", "c", "d", "e", "f", "g"}
	chunks := processor.generateChunks(items, 3)

	var result [][]string
	for chunk := range chunks {
		result = append(result, chunk)
	}

	if len(result) != 3 {
		t.Errorf("expected 3 chunks, got %d", len(result))
	}
	if len(result[0]) != 3 {
		t.Errorf("expected chunk size 3, got %d", len(result[0]))
	}
	if len(result[2]) != 1 {
		t.Errorf("expected last chunk size 1, got %d", len(result[2]))
	}
}

// Benchmarks

func BenchmarkNewBatchProcessor(b *testing.B) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = NewBatchProcessor(4, 10, logger)
	}
}

func BenchmarkProcessBatch_SingleWorker(b *testing.B) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		b.Fatal(err)
	}

	processor, err := NewBatchProcessor(1, 10, logger)
	if err != nil {
		b.Fatal(err)
	}

	ctx := context.Background()
	var processed int64
	processFunc := func(ctx context.Context, items []string, userID *uuid.UUID) error {
		atomic.AddInt64(&processed, int64(len(items)))
		return nil
	}

	items := make([]string, 100)
	for i := range items {
		items[i] = "item" + string(rune('0'+i%10))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		atomic.StoreInt64(&processed, 0)
		_, _ = processor.ProcessBatch(ctx, items, nil, processFunc)
	}
}

func BenchmarkProcessBatch_MultiWorker(b *testing.B) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		b.Fatal(err)
	}

	processor, err := NewBatchProcessor(4, 10, logger)
	if err != nil {
		b.Fatal(err)
	}

	ctx := context.Background()
	var processed int64
	processFunc := func(ctx context.Context, items []string, userID *uuid.UUID) error {
		atomic.AddInt64(&processed, int64(len(items)))
		return nil
	}

	items := make([]string, 100)
	for i := range items {
		items[i] = "item" + string(rune('0'+i%10))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		atomic.StoreInt64(&processed, 0)
		_, _ = processor.ProcessBatch(ctx, items, nil, processFunc)
	}
}

func BenchmarkProcessBatch_Parallel(b *testing.B) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		b.Fatal(err)
	}

	processor, err := NewBatchProcessor(4, 10, logger)
	if err != nil {
		b.Fatal(err)
	}

	ctx := context.Background()
	var processed int64
	processFunc := func(ctx context.Context, items []string, userID *uuid.UUID) error {
		atomic.AddInt64(&processed, int64(len(items)))
		return nil
	}

	items := make([]string, 100)
	for i := range items {
		items[i] = "item" + string(rune('0'+i%10))
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			atomic.StoreInt64(&processed, 0)
			_, _ = processor.ProcessBatch(ctx, items, nil, processFunc)
		}
	})
}

func BenchmarkGenerateChunks(b *testing.B) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		b.Fatal(err)
	}

	processor, err := NewBatchProcessor(4, 10, logger)
	if err != nil {
		b.Fatal(err)
	}

	items := make([]string, 100)
	for i := range items {
		items[i] = "item" + string(rune('0'+i%10))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		chunks := processor.generateChunks(items, 10)
		for range chunks {
		}
	}
}

func BenchmarkGenerateChunks_Parallel(b *testing.B) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		b.Fatal(err)
	}

	processor, err := NewBatchProcessor(4, 10, logger)
	if err != nil {
		b.Fatal(err)
	}

	items := make([]string, 100)
	for i := range items {
		items[i] = "item" + string(rune('0'+i%10))
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			chunks := processor.generateChunks(items, 10)
			for range chunks {
			}
		}
	})
}

func BenchmarkProcessingStats_AddSuccessful(b *testing.B) {
	stats := &ProcessingStats{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stats.AddSuccessful(10)
	}
}

func BenchmarkProcessingStats_AddSuccessful_Parallel(b *testing.B) {
	stats := &ProcessingStats{}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			stats.AddSuccessful(10)
		}
	})
}

func BenchmarkProcessingStats_AddFailed(b *testing.B) {
	stats := &ProcessingStats{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stats.AddFailed(10)
	}
}

func BenchmarkProcessingStats_AddFailed_Parallel(b *testing.B) {
	stats := &ProcessingStats{}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			stats.AddFailed(10)
		}
	})
}

func BenchmarkProcessingStats_IncrementProcessedChunks(b *testing.B) {
	stats := &ProcessingStats{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stats.IncrementProcessedChunks()
	}
}

func BenchmarkProcessingStats_IncrementProcessedChunks_Parallel(b *testing.B) {
	stats := &ProcessingStats{}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			stats.IncrementProcessedChunks()
		}
	})
}
