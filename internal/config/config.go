// Package config предоставляет инструменты для управления конфигурацией приложения.
//
// Конфигурация может быть задана через флаги командной строки или переменные окружения.
// Переменные окружения имеют приоритет над флагами.
package config

import (
	"flag"
	"os"

	"github.com/joho/godotenv"
)

// Config содержит параметры конфигурации приложения.
//
// Поля могут быть заданы через флаги командной строки или переменные окружения.
// Приоритет: переменные окружения > флаги командной строки.
type Config struct {
	// ServerAddress — адрес HTTP-сервера в формате "host:port".
	ServerAddress string
	// BaseURL — базовый URL для генерации коротких ссылок.
	BaseURL string
	// FileStoragePath — путь к файлу для хранения URL (используется при отключенной БД).
	FileStoragePath string
	// DatabaseDSN — DSN для подключения к PostgreSQL (если пуст, используется файловое хранилище).
	DatabaseDSN string
	// MigrationsPath — путь к миграциям базы данных.
	MigrationsPath string
	// AuthSecret — секретный ключ для подписи JWT-токенов.
	AuthSecret string
	// AuditFile — путь к файлу для аудит-логов (если задан, используется FileObserver).
	AuditFile string
	// AuditURL — URL для отправки аудит-событий (если задан, используется HTTPObserver).
	AuditURL string
	// EnableHTTPS — флаг для включения HTTPS.
	EnableHTTPS bool
}

// InitConfig инициализирует конфигурацию из флагов командной строки и переменных окружения.
//
// Поддерживаемые флаги:
//
//	-a, --address     Адрес HTTP-сервера (по умолчанию: localhost:8080)
//	-b, --base-url    Базовый URL для коротких ссылок (по умолчанию: http://localhost:8080)
//	-f, --file        Путь к файлу хранилища (по умолчанию: file_storage.jsonl)
//	-d, --database    DSN базы данных (если не задан, используется файловое хранилище)
//	-m, --migrations  Путь к миграциям (по умолчанию: file://migrations)
//	-s, --secret      Секретный ключ для JWT (по умолчанию: default_secret_key)
//	--audit-file      Путь к файлу аудит-логов
//	--audit-url       URL для аудит-логов
//	--enable-https     Включение HTTPS (по умолчанию: false)
//
// Поддерживаемые переменные окружения:
//
//	SERVER_ADDRESS, BASE_URL, FILE_STORAGE_PATH, DATABASE_DSN, MIGRATIONS_PATH,
//	AUTH_SECRET, AUDIT_FILE, AUDIT_URL, ENABLE_HTTPS
//
// Пример использования:
//
//	cfg := config.InitConfig()
//	if cfg.DatabaseDSN != "" {
//	    // Использовать PostgreSQL
//	} else {
//	    // Использовать файловое хранилище
//	}
func InitConfig() *Config {
	_ = godotenv.Load()

	config := &Config{}

	flag.StringVar(&config.ServerAddress, "a", "localhost:8080", "HTTP server address")
	flag.StringVar(&config.BaseURL, "b", "http://localhost:8080", "Base URL for shortened URLs")
	flag.StringVar(&config.FileStoragePath, "f", "file_storage.jsonl", "Path to file storage")
	flag.StringVar(&config.DatabaseDSN, "d", "", "Database DSN (if not set, file storage will be used)")
	flag.StringVar(&config.MigrationsPath, "m", "file://migrations", "Path to migrations")
	flag.StringVar(&config.AuthSecret, "s", "default_secret_key", "Secret key for JWT signing")
	flag.StringVar(&config.AuditFile, "audit-file", "", "Path to audit log file")
	flag.StringVar(&config.AuditURL, "audit-url", "", "URL for audit log service")
	flag.BoolVar(&config.EnableHTTPS, "enable-https", false, "Enable HTTPS")
	flag.Parse()

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

	return config
}
