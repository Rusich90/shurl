// Package audit содержит тесты для фабричных функций менеджера аудита.
package audit

import (
	"os"
	"testing"

	"github.com/Rusich90/shurl.git/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewManagerWithConfig_NoObservers(t *testing.T) {
	cfg := &config.Config{
		AuditFile: "",
		AuditURL:  "",
	}

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	auditManager, err := NewManagerWithConfig(cfg, logger)
	require.NoError(t, err)
	require.NotNil(t, auditManager)
	defer auditManager.Close()
}

func TestNewManagerWithConfig_WithFileObserver(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_audit_*.log")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	cfg := &config.Config{
		AuditFile: tmpFile.Name(),
		AuditURL:  "",
	}

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	auditManager, err := NewManagerWithConfig(cfg, logger)
	require.NoError(t, err)
	require.NotNil(t, auditManager)
	defer auditManager.Close()
}

func TestNewManagerWithConfig_WithHTTPObserver(t *testing.T) {
	cfg := &config.Config{
		AuditFile: "",
		AuditURL:  "http://localhost:8080/audit",
	}

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	auditManager, err := NewManagerWithConfig(cfg, logger)
	require.NoError(t, err)
	require.NotNil(t, auditManager)
	defer auditManager.Close()
}

func TestNewManagerWithConfig_WithBothObservers(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_audit_*.log")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	cfg := &config.Config{
		AuditFile: tmpFile.Name(),
		AuditURL:  "http://localhost:8080/audit",
	}

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	auditManager, err := NewManagerWithConfig(cfg, logger)
	require.NoError(t, err)
	require.NotNil(t, auditManager)
	defer auditManager.Close()
}

func TestNewManagerWithConfig_InvalidFileObserver(t *testing.T) {
	cfg := &config.Config{
		AuditFile: "/invalid/path/to/file",
		AuditURL:  "",
	}

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	auditManager, err := NewManagerWithConfig(cfg, logger)
	assert.Error(t, err)
	assert.Nil(t, auditManager)
}