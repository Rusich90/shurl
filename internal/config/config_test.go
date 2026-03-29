package config

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitConfig(t *testing.T) {
	oldEnv := map[string]string{}
	keys := []string{"SERVER_ADDRESS", "BASE_URL", "FILE_STORAGE_PATH", "DATABASE_DSN", "MIGRATIONS_PATH", "AUTH_SECRET", "AUDIT_FILE", "AUDIT_URL", "ENABLE_HTTPS", "CONFIG"}
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
	assert.Equal(t, false, cfg.EnableHTTPS)
}

func TestLoadConfigFromFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	configContent := `{
		"server_address": "localhost:9090",
		"base_url": "http://example.com",
		"file_storage_path": "/path/to/file.db",
		"database_dsn": "postgres://user:pass@localhost:5432/db",
		"migrations_path": "file://custom/migrations",
		"auth_secret": "custom_secret",
		"audit_file": "/path/to/audit.log",
		"audit_url": "http://audit.example.com",
		"enable_https": true
	}`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	assert.NoError(t, err)

	cfg := &Config{}
	err = loadConfigFromFile(configPath, cfg)

	assert.NoError(t, err)
	assert.Equal(t, "localhost:9090", cfg.ServerAddress)
	assert.Equal(t, "http://example.com", cfg.BaseURL)
	assert.Equal(t, "/path/to/file.db", cfg.FileStoragePath)
	assert.Equal(t, "postgres://user:pass@localhost:5432/db", cfg.DatabaseDSN)
	assert.Equal(t, "file://custom/migrations", cfg.MigrationsPath)
	assert.Equal(t, "custom_secret", cfg.AuthSecret)
	assert.Equal(t, "/path/to/audit.log", cfg.AuditFile)
	assert.Equal(t, "http://audit.example.com", cfg.AuditURL)
	assert.Equal(t, true, cfg.EnableHTTPS)
}

func TestLoadConfigFromFileNotFound(t *testing.T) {
	cfg := &Config{}
	err := loadConfigFromFile("/nonexistent/path/config.json", cfg)

	assert.Error(t, err)
}

func TestLoadConfigFromFileInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	configContent := `invalid json content`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	assert.NoError(t, err)

	cfg := &Config{}
	err = loadConfigFromFile(configPath, cfg)

	assert.Error(t, err)
}

func TestLoadConfigFromFilePartial(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	// Файл с частичными настройками
	configContent := `{
		"server_address": "localhost:9090",
		"enable_https": true
	}`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	assert.NoError(t, err)

	cfg := &Config{}
	cfg.ServerAddress = "default:8080"
	cfg.EnableHTTPS = false

	err = loadConfigFromFile(configPath, cfg)

	assert.NoError(t, err)
	// Значение из файла должно быть применено
	assert.Equal(t, "localhost:9090", cfg.ServerAddress)
	// Значение из файла должно быть применено
	assert.Equal(t, true, cfg.EnableHTTPS)
}

func TestLoadConfigFromFileEmptyValues(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	// Файл с пустыми значениями (должны быть проигнорированы)
	configContent := `{
		"server_address": "",
		"base_url": "",
		"file_storage_path": "",
		"database_dsn": "",
		"migrations_path": "",
		"auth_secret": "",
		"audit_file": "",
		"audit_url": "",
		"enable_https": false
	}`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	assert.NoError(t, err)

	cfg := &Config{}
	cfg.ServerAddress = "default:8080"
	cfg.BaseURL = "http://default.com"
	cfg.FileStoragePath = "default.jsonl"
	cfg.DatabaseDSN = "default://db"
	cfg.MigrationsPath = "default://migrations"
	cfg.AuthSecret = "default_secret"
	cfg.AuditFile = "default_audit.log"
	cfg.AuditURL = "http://default.com/audit"
	cfg.EnableHTTPS = false

	err = loadConfigFromFile(configPath, cfg)

	assert.NoError(t, err)
	// Пустые значения должны быть проигнорированы
	assert.Equal(t, "default:8080", cfg.ServerAddress)
	assert.Equal(t, "http://default.com", cfg.BaseURL)
	assert.Equal(t, "default.jsonl", cfg.FileStoragePath)
	assert.Equal(t, "default://db", cfg.DatabaseDSN)
	assert.Equal(t, "default://migrations", cfg.MigrationsPath)
	assert.Equal(t, "default_secret", cfg.AuthSecret)
	assert.Equal(t, "default_audit.log", cfg.AuditFile)
	assert.Equal(t, "http://default.com/audit", cfg.AuditURL)
	// Для bool значение false является нулевым, поэтому оно будет применено
	assert.Equal(t, false, cfg.EnableHTTPS)
}

func TestConfigPriority(t *testing.T) {
	oldEnv := map[string]string{}
	keys := []string{"SERVER_ADDRESS", "BASE_URL", "FILE_STORAGE_PATH", "DATABASE_DSN", "MIGRATIONS_PATH", "AUTH_SECRET", "AUDIT_FILE", "AUDIT_URL", "ENABLE_HTTPS", "CONFIG"}
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

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	// 1) Файл с 3 полями (самый низкий приоритет)
	configContent := `{
		"server_address": "from-file:8080",
		"base_url": "http://from-file.com",
		"enable_https": true
	}`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	assert.NoError(t, err)

	// Задаем дефолтные значения (самый низкий приоритет)
	config := &Config{}
	config.ServerAddress = "localhost:8080"
	config.BaseURL = "http://localhost:8080"
	config.FileStoragePath = "file_storage.jsonl"
	config.DatabaseDSN = ""
	config.MigrationsPath = "file://migrations"
	config.AuthSecret = "default_secret_key"
	config.AuditFile = ""
	config.AuditURL = ""
	config.EnableHTTPS = false

	// Загружаем конфигурацию из файла (перезаписывает дефолтные значения)
	err = loadConfigFromFile(configPath, config)
	assert.NoError(t, err)

	// Применяем флаги (средний приоритет) - один перезаписывает файл, два новых
	// Используем отдельный FlagSet для парсинга флагов
	cfgFlagSet := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	cfgFlagSet.StringVar(&config.ServerAddress, "a", "", "HTTP server address")
	cfgFlagSet.StringVar(&config.FileStoragePath, "f", "", "Path to file storage")
	cfgFlagSet.StringVar(&config.DatabaseDSN, "d", "", "Database DSN")
	cfgFlagSet.Parse([]string{"-a", "from-flag:9999", "-f", "from-flag.jsonl", "-d", "from-flag://db"})

	// 3) Переменные окружения - 4 переменных: 2 новых, 1 перезаписывает флаг, 1 перезаписывает файл
	// ОСТАВЛЯЕМ BASE_URL БЕЗ ИЗМЕНЕНИЙ - должно быть из файла
	os.Setenv("FILE_STORAGE_PATH", "from-env.jsonl") // перезаписывает флаг
	os.Setenv("AUTH_SECRET", "from-env-secret")      // новая переменная (из файла)
	os.Setenv("AUDIT_FILE", "/from-env/audit.log")   // новая переменная (дефолт)

	// Применяем переменные окружения (самый высокий приоритет)
	if envServAddr, exists := os.LookupEnv("SERVER_ADDRESS"); exists {
		config.ServerAddress = envServAddr
	}
	if envBaseURL, exists := os.LookupEnv("BASE_URL"); exists {
		config.BaseURL = envBaseURL
	}
	if envFileStoragePath, exists := os.LookupEnv("FILE_STORAGE_PATH"); exists {
		config.FileStoragePath = envFileStoragePath
	}
	if envDatabaseDSN, exists := os.LookupEnv("DATABASE_DSN"); exists {
		config.DatabaseDSN = envDatabaseDSN
	}
	if envMigrationsPath, exists := os.LookupEnv("MIGRATIONS_PATH"); exists {
		config.MigrationsPath = envMigrationsPath
	}
	if envAuthSecret, exists := os.LookupEnv("AUTH_SECRET"); exists {
		config.AuthSecret = envAuthSecret
	}
	if envAuditFile, exists := os.LookupEnv("AUDIT_FILE"); exists {
		config.AuditFile = envAuditFile
	}
	if envAuditURL, exists := os.LookupEnv("AUDIT_URL"); exists {
		config.AuditURL = envAuditURL
	}
	if envEnableHTTPS, exists := os.LookupEnv("ENABLE_HTTPS"); exists {
		config.EnableHTTPS = envEnableHTTPS == "true"
	}

	// Проверяем приоритеты
	// SERVER_ADDRESS: файл -> флаг -> переменная (но переменная не задана, так что флаг)
	assert.Equal(t, "from-flag:9999", config.ServerAddress, "флаг должен перезаписать файл")

	// BASE_URL: файл (флаг не задан, переменная не задана)
	assert.Equal(t, "http://from-file.com", config.BaseURL, "должно быть из файла")

	// FILE_STORAGE_PATH: файл -> флаг -> переменная
	assert.Equal(t, "from-env.jsonl", config.FileStoragePath, "переменная должна перезаписать флаг")

	// DATABASE_DSN: флаг (файл не задал)
	assert.Equal(t, "from-flag://db", config.DatabaseDSN, "должно быть из флага")

	// MIGRATIONS_PATH: дефолт (файл не задал, флаг не задал)
	assert.Equal(t, "file://migrations", config.MigrationsPath, "должно быть дефолтное значение")

	// AUTH_SECRET: файл -> переменная (флаг не задал)
	assert.Equal(t, "from-env-secret", config.AuthSecret, "переменная должна перезаписать файл")

	// AUDIT_FILE: переменная (файл не задал, флаг не задал)
	assert.Equal(t, "/from-env/audit.log", config.AuditFile, "должно быть из переменной")

	// ENABLE_HTTPS: файл (флаг не задал, переменная не задана)
	assert.Equal(t, true, config.EnableHTTPS, "должно быть из файла")
}
