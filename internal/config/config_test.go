package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitConfig(t *testing.T) {
	oldEnv := map[string]string{}
	keys := []string{"SERVER_ADDRESS", "BASE_URL", "FILE_STORAGE_PATH", "DATABASE_DSN", "MIGRATIONS_PATH", "AUTH_SECRET", "AUDIT_FILE", "AUDIT_URL"}
	for _, k := range keys {
		if v, ok := os.LookupEnv(k); ok {
			oldEnv[k] = v
		}
	}

	for _, k := range keys {
		os.Unsetenv(k)
	}

	defer func() {
		for k, v := range oldEnv {
			os.Setenv(k, v)
		}
	}()

	cfg := InitConfig()

	assert.Equal(t, "localhost:8080", cfg.ServerAddress)
	assert.Equal(t, "http://localhost:8080", cfg.BaseURL)
	assert.Equal(t, "file_storage.jsonl", cfg.FileStoragePath)
	assert.Equal(t, "", cfg.DatabaseDSN)
	assert.Equal(t, "file://migrations", cfg.MigrationsPath)
	assert.Equal(t, "default_secret_key", cfg.AuthSecret)
	assert.Equal(t, "", cfg.AuditFile)
	assert.Equal(t, "", cfg.AuditURL)
}
