// Package batch предоставляет инструменты для пакетной обработки данных.
//
// Использует паттерн "Worker Pool" для параллельной обработки больших
// объемов данных с контролем ошибок и статистикой.
package batch

import (
	"sync/atomic"
)

// ProcessingStats содержит статистику по пакетной обработке.
type ProcessingStats struct {
	// SuccessfulDeletes — количество успешно обработанных элементов.
	SuccessfulDeletes int64
	// FailedDeletes — количество неудачных попыток обработки.
	FailedDeletes int64
	// ProcessedChunks — количество обработанных пакетов.
	ProcessedChunks int64
}

// AddSuccessful увеличивает счетчик успешных операций на указанное количество.
func (s *ProcessingStats) AddSuccessful(count int) {
	atomic.AddInt64(&s.SuccessfulDeletes, int64(count))
}

// AddFailed увеличивает счетчик неудачных операций на указанное количество.
func (s *ProcessingStats) AddFailed(count int) {
	atomic.AddInt64(&s.FailedDeletes, int64(count))
}

// IncrementProcessedChunks увеличивает счетчик обработанных пакетов на 1.
func (s *ProcessingStats) IncrementProcessedChunks() {
	atomic.AddInt64(&s.ProcessedChunks, 1)
}
