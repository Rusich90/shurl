package audit

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestFileObserver_NewFileObserver(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_audit_*.log")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	observer, err := NewFileObserver(tmpFile.Name())
	assert.NoError(t, err)
	assert.NotNil(t, observer)

	observer.Close()
}

func TestFileObserver_Notify(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_audit_*.log")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	observer, err := NewFileObserver(tmpFile.Name())
	assert.NoError(t, err)
	defer observer.Close()

	ctx := context.Background()
	event := NewAuditEvent(ActionShorten, nil, "https://example.com")

	err = observer.Notify(ctx, event)
	assert.NoError(t, err)

	time.Sleep(100 * time.Millisecond)
}

func TestFileObserver_Notify_WithContextCancellation(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_audit_*.log")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	observer, err := NewFileObserver(tmpFile.Name())
	assert.NoError(t, err)
	defer observer.Close()

	ctx, cancel := context.WithCancel(context.Background())

	cancel()

	event := NewAuditEvent(ActionShorten, nil, "https://example.com")

	err = observer.Notify(ctx, event)
	assert.Error(t, err)
}

func TestFileObserver_Close(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_audit_*.log")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	observer, err := NewFileObserver(tmpFile.Name())
	assert.NoError(t, err)

	observer.Close()
}

func TestFileObserver_MultipleEvents(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_audit_*.log")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	observer, err := NewFileObserver(tmpFile.Name())
	assert.NoError(t, err)
	defer observer.Close()

	ctx := context.Background()
	userID := uuid.New()

	events := []AuditEvent{
		NewAuditEvent(ActionShorten, &userID, "https://example.com/1"),
		NewAuditEvent(ActionShorten, &userID, "https://example.com/2"),
		NewAuditEvent(ActionFollow, &userID, "https://example.com/3"),
	}

	for _, event := range events {
		err = observer.Notify(ctx, event)
		assert.NoError(t, err)
	}

	time.Sleep(100 * time.Millisecond)
}

func TestHTTPObserver_NewHTTPObserver(t *testing.T) {
	observer := NewHTTPObserver("http://localhost:8080/audit")
	assert.NotNil(t, observer)
	assert.Equal(t, "http://localhost:8080/audit", observer.url)
	assert.NotNil(t, observer.client)
}

func TestHTTPObserver_Notify(t *testing.T) {
	observer := NewHTTPObserver("http://localhost:9999/nonexistent") // Несуществующий сервер

	ctx := context.Background()
	event := NewAuditEvent(ActionShorten, nil, "https://example.com")

	err := observer.Notify(ctx, event)
	assert.Error(t, err)
}

func TestHTTPObserver_Notify_WithContextCancellation(t *testing.T) {
	observer := NewHTTPObserver("http://localhost:9999/nonexistent")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	event := NewAuditEvent(ActionShorten, nil, "https://example.com")

	err := observer.Notify(ctx, event)
	assert.Error(t, err)
}

func TestAuditEvent_MarshalJSON(t *testing.T) {
	userID := uuid.New()
	event := AuditEvent{
		Timestamp: 1234567890,
		Action:    ActionShorten,
		UserID:    &userID,
		URL:       "https://example.com",
	}

	data, err := event.MarshalJSON()
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"ts":1234567890`)
	assert.Contains(t, string(data), `"action":"shorten"`)
	assert.Contains(t, string(data), `"url":"https://example.com"`)
	assert.Contains(t, string(data), `"user_id":"`+userID.String()+`"`)
}

func TestAuditEvent_MarshalJSON_NoUserID(t *testing.T) {
	event := AuditEvent{
		Timestamp: 1234567890,
		Action:    ActionShorten,
		UserID:    nil,
		URL:       "https://example.com",
	}

	data, err := event.MarshalJSON()
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"ts":1234567890`)
	assert.Contains(t, string(data), `"action":"shorten"`)
	assert.Contains(t, string(data), `"url":"https://example.com"`)
	assert.NotContains(t, string(data), `"user_id"`)
}

func TestNewAuditEvent(t *testing.T) {
	userID := uuid.New()
	event := NewAuditEvent(ActionShorten, &userID, "https://example.com")

	assert.Equal(t, ActionShorten, event.Action)
	assert.Equal(t, "https://example.com", event.URL)
	assert.Equal(t, userID, *event.UserID)
	assert.NotZero(t, event.Timestamp)
}

func TestNewAuditEvent_NoUserID(t *testing.T) {
	event := NewAuditEvent(ActionFollow, nil, "https://example.com")

	assert.Equal(t, ActionFollow, event.Action)
	assert.Equal(t, "https://example.com", event.URL)
	assert.Nil(t, event.UserID)
	assert.NotZero(t, event.Timestamp)
}
