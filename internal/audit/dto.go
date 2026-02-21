package audit

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type AuditAction string

const (
	ActionShorten AuditAction = "shorten"
	ActionFollow  AuditAction = "follow"
)

type AuditEvent struct {
	Timestamp int64       `json:"ts"`      // Unix timestamp of the event
	Action    AuditAction `json:"action"`  // Action type
	UserID    *uuid.UUID  `json:"user_id"` // User identifier (if available)
	URL       string      `json:"url"`     // Original URL
}

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

func NewAuditEvent(action AuditAction, userID *uuid.UUID, url string) AuditEvent {
	return AuditEvent{
		Timestamp: time.Now().Unix(),
		Action:    action,
		UserID:    userID,
		URL:       url,
	}
}
