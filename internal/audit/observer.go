// Package audit предоставляет инструменты для аудита событий в системе.
//
// Аудит позволяет отслеживать ключевые события, такие как создание коротких URL
// и переходы по ним, с возможностью отправки уведомлений в различные системы
// (файловые логи, HTTP-сервисы).
package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// Observer — интерфейс для наблюдателей аудит-событий.
//
// Реализует паттерн "Наблюдатель" для асинхронной обработки событий.
type Observer interface {
	// Notify отправляет аудит-событие наблюдателю.
	//
	// Возвращает ошибку при неудачной отправке события.
	Notify(ctx context.Context, event AuditEvent) error
}

// Closer — расширение Observer с возможностью закрытия.
type Closer interface {
	Observer
	// Close останавливает работу наблюдателя и дожидается завершения всех операций.
	Close()
}

// FileObserver — наблюдатель, записывающий аудит-события в файл.
//
// Использует буферизированный канал для асинхронной записи событий,
// что позволяет избежать блокировки основного потока выполнения.
type FileObserver struct {
	filePath string
	events   chan AuditEvent
	stop     chan struct{}
	done     chan struct{}
}

// NewFileObserver создает новый FileObserver, который будет записывать события в указанный файл.
//
// Файл создается при необходимости с правами 0644. Если файл уже существует,
// новые события будут добавлены в конец файла.
//
// Пример использования:
//
//	observer, err := audit.NewFileObserver("/var/log/audit.log")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer observer.Close()
func NewFileObserver(filePath string) (*FileObserver, error) {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to access file: %w", err)
	}
	file.Close()

	fo := &FileObserver{
		filePath: filePath,
		events:   make(chan AuditEvent),
		stop:     make(chan struct{}),
		done:     make(chan struct{}),
	}

	go fo.writer()

	return fo, nil
}

// Notify добавляет аудит-событие в буфер для записи в файл.
//
// Метод неблокирующий — событие помещается в канал и будет записано
// в отдельной горутине. Если контекст отменен, возвращается ошибка ctx.Err().
func (fo *FileObserver) Notify(ctx context.Context, event AuditEvent) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case fo.events <- event:
		return nil
	}
}

// Close останавливает работу FileObserver и дожидается завершения записи
// всех оставшихся событий.
func (fo *FileObserver) Close() {
	close(fo.stop)
	<-fo.done
}

// writer — внутренняя горутина, записывающая события в файл.
//
// Запускается автоматически при создании FileObserver и работает до вызова Close().
func (fo *FileObserver) writer() {
	defer func() {
		close(fo.done)
	}()

	file, err := os.OpenFile(fo.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer file.Close()

	for {
		select {
		case event, ok := <-fo.events:
			if !ok {
				return
			}

			data, err := json.Marshal(event)
			if err != nil {
				continue
			}

			_, err = file.Write(append(data, '\n'))
			if err != nil {
				continue
			}

		case <-fo.stop:
			for {
				select {
				case event, ok := <-fo.events:
					if !ok {
						return
					}

					data, _ := json.Marshal(event)
					file.Write(append(data, '\n'))
				default:
					return
				}
			}
		}
	}
}

// HTTPObserver — наблюдатель, отправляющий аудит-события по HTTP.
//
// Отправляет события методом POST на указанный URL в формате JSON.
type HTTPObserver struct {
	url    string
	client *http.Client
}

// NewHTTPObserver создает новый HTTPObserver для отправки событий по HTTP.
//
// Использует стандартный http.Client без дополнительной настройки.
func NewHTTPObserver(url string) *HTTPObserver {
	return &HTTPObserver{
		url:    url,
		client: &http.Client{},
	}
}

// Notify отправляет аудит-событие по HTTP.
//
// Метод блокирующий и возвращает ошибку при:
//   - Неудачной сериализации события в JSON
//   - Ошибке создания HTTP-запроса
//   - Ошибке отправки запроса
//   - Нестатусе ответа 2xx
func (ho *HTTPObserver) Notify(ctx context.Context, event AuditEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal audit event: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ho.url, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := ho.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send audit event: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected response status: %d", resp.StatusCode)
	}

	return nil
}
