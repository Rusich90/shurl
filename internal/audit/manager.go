// Package audit предоставляет инструменты для аудита событий в системе.
//
// Аудит позволяет отслеживать ключевые события, такие как создание коротких URL
// и переходы по ним, с возможностью отправки уведомлений в различные системы
// (файловые логи, HTTP-сервисы).
package audit

import (
	"context"

	"go.uber.org/zap"
)

// Close закрывает все наблюдатели, которые поддерживают закрытие.
func (am *Manager) Close() {
	for _, observer := range am.observers {
		if closer, ok := observer.(Closer); ok {
			closer.Close()
		}
	}
}

// Manager управляет коллекцией наблюдателей аудит-событий.
//
// Позволяет регистрировать новые наблюдатели и уведомлять их обо всех событиях.
type Manager struct {
	observers []Observer
	logger    *zap.Logger
}

// NewManager создает новый менеджер аудита с указанным логгером.
//
// Менеджер используется для централизованной отправки аудит-событий
// всем зарегистрированным наблюдателям.
func NewManager(logger *zap.Logger) *Manager {
	return &Manager{
		observers: make([]Observer, 0),
		logger:    logger,
	}
}

// RegisterObserver регистрирует новый наблюдатель для получения аудит-событий.
//
// Регистрированные наблюдатели будут получать все события, отправленные через NotifyAll.
func (am *Manager) RegisterObserver(observer Observer) {
	am.observers = append(am.observers, observer)
}

// NotifyAll отправляет аудит-событие всем зарегистрированным наблюдателям.
//
// Метод асинхронный — ошибки отправки логируются, но не прерывают выполнение
// для других наблюдателей.
func (am *Manager) NotifyAll(ctx context.Context, event AuditEvent) {
	for _, observer := range am.observers {
		if err := observer.Notify(ctx, event); err != nil {
			am.logger.Error("Audit observer error", zap.Error(err))
		}
	}
}
