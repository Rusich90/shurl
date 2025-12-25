package batch

import (
	"sync/atomic"
)

type ProcessingStats struct {
	SuccessfulDeletes int64
	FailedDeletes     int64
	ProcessedChunks   int64
}

func (s *ProcessingStats) AddSuccessful(count int) {
	atomic.AddInt64(&s.SuccessfulDeletes, int64(count))
}

func (s *ProcessingStats) AddFailed(count int) {
	atomic.AddInt64(&s.FailedDeletes, int64(count))
}

func (s *ProcessingStats) IncrementProcessedChunks() {
	atomic.AddInt64(&s.ProcessedChunks, 1)
}
