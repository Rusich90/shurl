// Package config предоставляет инструменты для управления конфигурацией приложения.
//
// Конфигурация может быть задана через флаги командной строки, переменные окружения
// или файл конфигурации в формате JSON.
//
// Приоритет значений (от высшего к низшему):
// 1. Переменные окружения
// 2. Флаги командной строки
// 3. Файл конфигурации (если указан)
// 4. Значения по умолчанию
package config

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// Config содержит параметры конфигурации приложения.
type Config struct {
	// ServerAddress — адрес HTTP-сервера в формате "host:port".
	ServerAddress string `json:"server_address"`
	// BaseURL — базовый URL для генерации коротких ссылок.
	BaseURL string `json:"base_url"`
	// FileStoragePath — путь к файлу для хранения URL (используется при отключенной БД).
	FileStoragePath string `json:"file_storage_path"`
	// DatabaseDSN — DSN для подключения к PostgreSQL (если пуст, используется файловое хранилище).
	DatabaseDSN string `json:"database_dsn"`
	// MigrationsPath — путь к миграциям базы данных.
	MigrationsPath string `json:"migrations_path,omitempty"`
	// AuthSecret — секретный ключ для подписи JWT-токенов.
	AuthSecret string `json:"auth_secret,omitempty"`
	// AuditFile — путь к файлу для аудит-логов (если задан, используется FileObserver).
	AuditFile string `json:"audit_file,omitempty"`
	// AuditURL — URL для отправки аудит-событий (если задан, используется HTTPObserver).
	AuditURL string `json:"audit_url,omitempty"`
	// EnableHTTPS — флаг для включения HTTPS.
	EnableHTTPS bool `json:"enable_https"`
	// TrustedSubnet — доверенная подсеть в формате CIDR для доступа к /api/internal/stats.
	TrustedSubnet string `json:"trusted_subnet,omitempty"`
}

// InitConfig инициализирует конфигурацию из флагов командной строки, переменных окружения
// и файла конфигурации.
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
//	-t, --trusted-subnet Доверенная подсеть в формате CIDR для доступа к /api/internal/stats
//	-c, --config      Путь к файлу конфигурации JSON
//
// Поддерживаемые переменные окружения:
//
//	SERVER_ADDRESS, BASE_URL, FILE_STORAGE_PATH, DATABASE_DSN, MIGRATIONS_PATH,
//	AUTH_SECRET, AUDIT_FILE, AUDIT_URL, ENABLE_HTTPS, TRUSTED_SUBNET, CONFIG
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

	config := &Config{
		ServerAddress:   "localhost:8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: "file_storage.jsonl",
		DatabaseDSN:     "",
		MigrationsPath:  "file://migrations",
		AuthSecret:      "default_secret_key",
		AuditFile:       "",
		AuditURL:        "",
		EnableHTTPS:     false,
		TrustedSubnet:   "",
	}

	// Сначала загружаем конфигурацию из файла (самый низкий приоритет)
	// Используем отдельный FlagSet для парсинга только флага -c
	fileFlagSet := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	var configPath string
	fileFlagSet.StringVar(&configPath, "c", "", "Path to JSON configuration file")
	_ = fileFlagSet.Parse(os.Args[1:])

	// Проверяем переменную окружения CONFIG, если флаг не задан
	if configPath == "" {
		if envConfigPath, exists := os.LookupEnv("CONFIG"); exists {
			configPath = envConfigPath
		}
	}

	// Загружаем конфигурацию из файла (перезаписывает дефолтные значения)
	if configPath != "" {
		if err := loadConfigFromFile(configPath, config); err != nil {
			// Если файл не найден или ошибка чтения, продолжаем с пустой конфигурацией
			_ = err // Логирование можно добавить при необходимости
		}
	}

	// Парсим флаги (средний приоритет) - они перезапишут значения из файла, если заданы
	// Используем отдельный FlagSet, чтобы избежать конфликта с флагами тестирования
	cfgFlagSet := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	cfgFlagSet.SetOutput(nil) // Отключаем вывод сообщений об ошибках
	cfgFlagSet.StringVar(&config.ServerAddress, "a", "localhost:8080", "HTTP server address")
	cfgFlagSet.StringVar(&config.BaseURL, "b", "http://localhost:8080", "Base URL for shortened URLs")
	cfgFlagSet.StringVar(&config.FileStoragePath, "f", "file_storage.jsonl", "Path to file storage")
	cfgFlagSet.StringVar(&config.DatabaseDSN, "d", "", "Database DSN (if not set, file storage will be used)")
	cfgFlagSet.StringVar(&config.MigrationsPath, "m", "file://migrations", "Path to migrations")
	cfgFlagSet.StringVar(&config.AuthSecret, "s", "default_secret_key", "Secret key for JWT signing")
	cfgFlagSet.StringVar(&config.AuditFile, "audit-file", "", "Path to audit log file")
	cfgFlagSet.StringVar(&config.AuditURL, "audit-url", "", "URL for audit log service")
	cfgFlagSet.BoolVar(&config.EnableHTTPS, "enable-https", false, "Enable HTTPS")
	cfgFlagSet.StringVar(&config.TrustedSubnet, "t", "", "Trusted subnet in CIDR format")

	// Пытаемся распарсить флаги, игнорируя ошибки (например, флаги тестирования)
	_ = cfgFlagSet.Parse(os.Args[1:])

	// Очищаем флаги, чтобы избежать конфликтов при повторном вызове
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	// Применяем значения из переменных окружения (самый высокий приоритет)
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
	if envTrustedSubnet, exists := os.LookupEnv("TRUSTED_SUBNET"); exists {
		config.TrustedSubnet = envTrustedSubnet
	}

	return config
}

// loadConfigFromFile загружает конфигурацию из JSON файла.
// Значения из файла применяются только если они не пустые (для строк) или false (для bool).
// Файловая конфигурация имеет самый низкий приоритет.
func loadConfigFromFile(filePath string, config *Config) error {
	// Получаем абсолютный путь к файлу
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return err
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return err
	}

	// Парсим JSON в map
	var dataMap map[string]interface{}
	if err := json.Unmarshal(data, &dataMap); err != nil {
		return err
	}

	// Применяем значения из файла, если они заданы
	if val, ok := dataMap["server_address"].(string); ok && val != "" {
		config.ServerAddress = val
	}
	if val, ok := dataMap["base_url"].(string); ok && val != "" {
		config.BaseURL = val
	}
	if val, ok := dataMap["file_storage_path"].(string); ok && val != "" {
		config.FileStoragePath = val
	}
	if val, ok := dataMap["database_dsn"].(string); ok && val != "" {
		config.DatabaseDSN = val
	}
	if val, ok := dataMap["migrations_path"].(string); ok && val != "" {
		config.MigrationsPath = val
	}
	if val, ok := dataMap["auth_secret"].(string); ok && val != "" {
		config.AuthSecret = val
	}
	if val, ok := dataMap["audit_file"].(string); ok && val != "" {
		config.AuditFile = val
	}
	if val, ok := dataMap["audit_url"].(string); ok && val != "" {
		config.AuditURL = val
	}
	if val, ok := dataMap["enable_https"].(bool); ok {
		config.EnableHTTPS = val
	}
	if val, ok := dataMap["trusted_subnet"].(string); ok && val != "" {
		config.TrustedSubnet = val
	}

	return nil
}
