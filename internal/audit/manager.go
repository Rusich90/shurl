package audit

import (
	"context"

	"go.uber.org/zap"
)

type Manager struct {
	observers []Observer
	logger    *zap.Logger
}

func NewManager(logger *zap.Logger) *Manager {
	return &Manager{
		observers: make([]Observer, 0),
		logger:    logger,
	}
}

func (am *Manager) RegisterObserver(observer Observer) {
	am.observers = append(am.observers, observer)
}

func (am *Manager) NotifyAll(ctx context.Context, event AuditEvent) {
	for _, observer := range am.observers {
		if err := observer.Notify(ctx, event); err != nil {
			am.logger.Error("Audit observer error", zap.Error(err))
		}
	}
}
