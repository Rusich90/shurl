# internal/config

В этом пакете хранятся конфигурации приложения.

## Загрузка конфигурации

Конфигурация может быть задана через флаги командной строки, переменные окружения или файл конфигурации в формате JSON.

### Приоритет значений

1. **Флаги командной строки** - имеют наивысший приоритет
2. **Переменные окружения** - имеют средний приоритет
3. **Файл конфигурации** - имеет низший приоритет
4. **Значения по умолчанию** - используются, если ничего не задано

### Флаги командной строки

| Флаг | Переменная окружения | Описание | Значение по умолчанию |
|------|---------------------|----------|----------------------|
| `-a` | `SERVER_ADDRESS` | Адрес HTTP-сервера | `localhost:8080` |
| `-b` | `BASE_URL` | Базовый URL для коротких ссылок | `http://localhost:8080` |
| `-f` | `FILE_STORAGE_PATH` | Путь к файлу хранилища | `file_storage.jsonl` |
| `-d` | `DATABASE_DSN` | DSN базы данных | (пусто, используется файловое хранилище) |
| `-m` | `MIGRATIONS_PATH` | Путь к миграциям | `file://migrations` |
| `-s` | `AUTH_SECRET` | Секретный ключ для JWT | `default_secret_key` |
| `--audit-file` | `AUDIT_FILE` | Путь к файлу аудит-логов | (пусто) |
| `--audit-url` | `AUDIT_URL` | URL для аудит-логов | (пусто) |
| `--enable-https` | `ENABLE_HTTPS` | Включение HTTPS | `false` |
| `-c` | `CONFIG` | Путь к файлу конфигурации JSON | (пусто) |

### Файл конфигурации

Файл конфигурации должен быть в формате JSON и содержать следующие поля:

```json
{
    "server_address": "localhost:8080",
    "base_url": "http://localhost",
    "file_storage_path": "/path/to/file.db",
    "database_dsn": "",
    "migrations_path": "file://migrations",
    "auth_secret": "secret_key",
    "audit_file": "/path/to/audit.log",
    "audit_url": "http://audit.example.com",
    "enable_https": true
}
```

### Пример использования

```go
package main

import (
    "log"
    "github.com/Rusich90/shurl.git/internal/config"
)

func main() {
    cfg := config.InitConfig()
    
    if cfg.DatabaseDSN != "" {
        // Использовать PostgreSQL
        log.Printf("Using database: %s", cfg.DatabaseDSN)
    } else {
        // Использовать файловое хранилище
        log.Printf("Using file storage: %s", cfg.FileStoragePath)
    }
    
    log.Printf("Server address: %s", cfg.ServerAddress)
    log.Printf("Base URL: %s", cfg.BaseURL)
}
```

### Примеры запуска

```bash
# Запуск с файлом конфигурации
./shortener -c config.json

# Запуск с переменной окружения CONFIG
export CONFIG=config.json
./shortener

# Запуск с флагами (переопределяет значения из файла)
./shortener -a localhost:9090 -c config.json

# Запуск с переменными окружения (переопределяет значения из файла)
export SERVER_ADDRESS=localhost:9090
./shortener -c config.json
