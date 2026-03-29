// Package audit предоставляет фабричные функции для создания менеджера аудита.
//
// Пакет содержит функции для инициализации менеджера аудита с наблюдателями
// на основе конфигурации приложения.
package audit

import (
	"fmt"

	"github.com/Rusich90/shurl.git/internal/config"
	"go.uber.org/zap"
)

// NewManagerWithConfig создает менеджер аудита с наблюдателями на основе конфигурации.
//
// Создает менеджер с указанным логгером и регистрирует наблюдатели:
// - FileObserver, если указан AuditFile в конфигурации
// - HTTPObserver, если указан AuditURL в конфигурации
//
// Возвращает менеджер аудита и ошибку при неудачной инициализации наблюдателей.
func NewManagerWithConfig(cfg *config.Config, logger *zap.Logger) (*Manager, error) {
	auditManager := NewManager(logger)

	if cfg.AuditFile != "" {
		fileObserver, err := NewFileObserver(cfg.AuditFile)
		if err != nil {
			return nil, fmt.Errorf("audit.NewFileObserver: %w", err)
		}
		auditManager.RegisterObserver(fileObserver)
	}

	if cfg.AuditURL != "" {
		httpObserver := NewHTTPObserver(cfg.AuditURL)
		auditManager.RegisterObserver(httpObserver)
	}

	return auditManager, nil
}
