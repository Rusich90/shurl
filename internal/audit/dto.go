package audit

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// AuditAction представляет тип аудит-события.
//
// Используется для классификации событий в системе аудита.
type AuditAction string

const (
	// ActionShorten — событие создания короткой URL.
	ActionShorten AuditAction = "shorten"
	// ActionFollow — событие перехода по короткой URL.
	ActionFollow AuditAction = "follow"
)

// AuditEvent представляет собой запись аудит-события.
//
// Содержит информацию о времени события, его типе, пользователе (если доступен)
// и связанном URL.
type AuditEvent struct {
	Timestamp int64       `json:"ts"`      // Unix timestamp of the event
	Action    AuditAction `json:"action"`  // Action type
	UserID    *uuid.UUID  `json:"user_id"` // User identifier (if available)
	URL       string      `json:"url"`     // Original URL
}

// MarshalJSON реализует интерфейс json.Marshaler для AuditEvent.
//
// Обеспечивает корректную сериализацию события в JSON с учетом
// необязательного поля UserID.
func (ae AuditEvent) MarshalJSON() ([]byte, error) {
	type Alias AuditEvent
	return json.Marshal(&struct {
		Timestamp int64       `json:"ts"`
		Action    AuditAction `json:"action"`
		UserID    *uuid.UUID  `json:"user_id,omitempty"`
		URL       string      `json:"url"`
	}{
		Timestamp: ae.Timestamp,
		Action:    ae.Action,
		UserID:    ae.UserID,
		URL:       ae.URL,
	})
}

// NewAuditEvent создает новое аудит-событие с указанными параметрами.
//
// timestamp автоматически устанавливается в текущее время.
func NewAuditEvent(action AuditAction, userID *uuid.UUID, url string) AuditEvent {
	return AuditEvent{
		Timestamp: time.Now().Unix(),
		Action:    action,
		UserID:    userID,
		URL:       url,
	}
}
