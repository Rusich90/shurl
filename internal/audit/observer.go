package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type Observer interface {
	Notify(ctx context.Context, event AuditEvent) error
}

type FileObserver struct {
	filePath string
	events   chan AuditEvent
	stop     chan struct{}
	done     chan struct{}
}

func NewFileObserver(filePath string) (*FileObserver, error) {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to access file: %w", err)
	}
	file.Close()

	fo := &FileObserver{
		filePath: filePath,
		events:   make(chan AuditEvent, 1000),
		stop:     make(chan struct{}),
		done:     make(chan struct{}),
	}

	go fo.writer()

	return fo, nil
}

func (fo *FileObserver) Notify(ctx context.Context, event AuditEvent) error {
	select {
	case fo.events <- event:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (fo *FileObserver) Close() {
	close(fo.stop)
	<-fo.done
}

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

type HTTPObserver struct {
	url    string
	client *http.Client
}

func NewHTTPObserver(url string) *HTTPObserver {
	return &HTTPObserver{
		url:    url,
		client: &http.Client{},
	}
}

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
