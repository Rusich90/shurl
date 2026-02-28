// Package batch предоставляет инструменты для пакетной обработки данных.
//
// Использует паттерн "Worker Pool" для параллельной обработки больших
// объемов данных с контролем ошибок и статистикой.
package batch

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// ProcessFunc определяет функцию для обработки пакета элементов.
//
// Принимает контекст, срез элементов и идентификатор пользователя.
type ProcessFunc func(ctx context.Context, items []string, userID *uuid.UUID) error

// BatchProcessor управляет пакетной обработкой данных.
//
// Распределяет задачи между несколькими воркерами и собирает статистику.
type BatchProcessor struct {
	numWorkers int
	chunkSize  int
	logger     *zap.Logger
}

// NewBatchProcessor создает новый BatchProcessor с указанными параметрами.
//
// numWorkers — количество параллельных воркеров.
// chunkSize — размер одного пакета элементов.
func NewBatchProcessor(numWorkers, chunkSize int, logger *zap.Logger) (*BatchProcessor, error) {
	if numWorkers <= 0 {
		return nil, fmt.Errorf("numWorkers must be positive, got %d", numWorkers)
	}

	if chunkSize <= 0 {
		return nil, fmt.Errorf("chunkSize must be positive, got %d", chunkSize)
	}

	return &BatchProcessor{
		numWorkers: numWorkers,
		chunkSize:  chunkSize,
		logger:     logger,
	}, nil
}

// ProcessBatch обрабатывает срез элементов с использованием пакетной обработки.
//
// Возвращает статистику обработки и ошибку при неудаче.
func (bp *BatchProcessor) ProcessBatch(ctx context.Context, items []string, userID *uuid.UUID, processFunc ProcessFunc) (*ProcessingStats, error) {
	if len(items) == 0 {
		return &ProcessingStats{}, nil
	}

	stats := &ProcessingStats{}

	chunksChan := bp.generateChunks(items, bp.chunkSize)

	g, ctx := errgroup.WithContext(ctx)

	for i := 0; i < bp.numWorkers; i++ {
		g.Go(func() error {
			for chunk := range chunksChan {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
				}

				err := processFunc(ctx, chunk, userID)
				if err != nil {
					stats.AddFailed(len(chunk))
					bp.logger.Error("Failed to process batch",
						zap.Error(err),
						zap.Int("chunk_size", len(chunk)),
					)
				} else {
					stats.AddSuccessful(len(chunk))
				}
				stats.IncrementProcessedChunks()
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return stats, fmt.Errorf("worker group error: %w", err)
	}

	return stats, nil
}

// generateChunks разбивает срез элементов на пакеты заданного размера.
//
// Возвращает канал, через который передаются пакеты элементов.
func (bp *BatchProcessor) generateChunks(items []string, chunkSize int) <-chan []string {
	inputCh := make(chan []string)

	go func() {
		defer close(inputCh)

		for i := 0; i < len(items); i += chunkSize {
			end := i + chunkSize
			if end > len(items) {
				end = len(items)
			}
			chunk := items[i:end]
			inputCh <- chunk
		}
	}()

	return inputCh
}
