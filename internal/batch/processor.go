package batch

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type ProcessFunc func(ctx context.Context, items []string, userID *uuid.UUID) error

type BatchProcessor struct {
	numWorkers int
	chunkSize  int
	logger     *zap.Logger
}

func NewBatchProcessor(numWorkers, chunkSize int, logger *zap.Logger) *BatchProcessor {
	return &BatchProcessor{
		numWorkers: numWorkers,
		chunkSize:  chunkSize,
		logger:     logger,
	}
}

func (bp *BatchProcessor) ProcessBatch(ctx context.Context, items []string, userID *uuid.UUID, processFunc ProcessFunc) (*ProcessingStats, error) {
	if len(items) == 0 {
		return &ProcessingStats{}, nil
	}

	stats := &ProcessingStats{}

	chunksChan := GenerateChunks(items, bp.chunkSize)

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
