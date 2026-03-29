package audit

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// MockObserver — мок-наблюдатель для тестов.
type MockObserver struct {
	notifyFunc func(ctx context.Context, event AuditEvent) error
	called     int
}

func (m *MockObserver) Notify(ctx context.Context, event AuditEvent) error {
	m.called++
	return m.notifyFunc(ctx, event)
}

func TestManager_NewManager(t *testing.T) {
	logger, err := zap.NewDevelopment()
	assert.NoError(t, err)

	manager := NewManager(logger)
	assert.NotNil(t, manager)
	assert.NotNil(t, manager.observers)
}

func TestManager_RegisterObserver(t *testing.T) {
	logger, err := zap.NewDevelopment()
	assert.NoError(t, err)

	manager := NewManager(logger)
	assert.NotNil(t, manager)

	observer := &MockObserver{}
	manager.RegisterObserver(observer)

	assert.Equal(t, 1, len(manager.observers))
}

func TestManager_NotifyAll(t *testing.T) {
	logger, err := zap.NewDevelopment()
	assert.NoError(t, err)

	manager := NewManager(logger)

	observer1 := &MockObserver{
		notifyFunc: func(ctx context.Context, event AuditEvent) error {
			return nil
		},
	}
	observer2 := &MockObserver{
		notifyFunc: func(ctx context.Context, event AuditEvent) error {
			return nil
		},
	}

	manager.RegisterObserver(observer1)
	manager.RegisterObserver(observer2)

	ctx := context.Background()
	event := NewAuditEvent(ActionShorten, nil, "https://example.com")

	manager.NotifyAll(ctx, event)

	assert.Equal(t, 1, observer1.called)
	assert.Equal(t, 1, observer2.called)
}

func TestManager_NotifyAll_MultipleEvents(t *testing.T) {
	logger, err := zap.NewDevelopment()
	assert.NoError(t, err)

	manager := NewManager(logger)

	observer := &MockObserver{
		notifyFunc: func(ctx context.Context, event AuditEvent) error {
			return nil
		},
	}

	manager.RegisterObserver(observer)

	ctx := context.Background()

	events := []AuditEvent{
		NewAuditEvent(ActionShorten, nil, "https://example.com/1"),
		NewAuditEvent(ActionFollow, nil, "https://example.com/2"),
		NewAuditEvent(ActionShorten, nil, "https://example.com/3"),
	}

	for _, event := range events {
		manager.NotifyAll(ctx, event)
	}

	assert.Equal(t, 3, observer.called)
}

func TestManager_NotifyAll_WithErrors(t *testing.T) {
	logger, err := zap.NewDevelopment()
	assert.NoError(t, err)

	manager := NewManager(logger)

	observer1 := &MockObserver{
		notifyFunc: func(ctx context.Context, event AuditEvent) error {
			return nil
		},
	}
	observer2 := &MockObserver{
		notifyFunc: func(ctx context.Context, event AuditEvent) error {
			return assert.AnError
		},
	}

	manager.RegisterObserver(observer1)
	manager.RegisterObserver(observer2)

	ctx := context.Background()
	event := NewAuditEvent(ActionShorten, nil, "https://example.com")

	manager.NotifyAll(ctx, event)

	assert.Equal(t, 1, observer1.called)
	assert.Equal(t, 1, observer2.called)
}

func TestManager_NotifyAll_ContextCancelled(t *testing.T) {
	logger, err := zap.NewDevelopment()
	assert.NoError(t, err)

	manager := NewManager(logger)

	observer := &MockObserver{
		notifyFunc: func(ctx context.Context, event AuditEvent) error {
			return nil
		},
	}

	manager.RegisterObserver(observer)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	event := NewAuditEvent(ActionShorten, nil, "https://example.com")

	manager.NotifyAll(ctx, event)

	assert.Equal(t, 1, observer.called)
}

func TestManager_Close(t *testing.T) {
	logger, err := zap.NewDevelopment()
	assert.NoError(t, err)

	manager := NewManager(logger)

	observer := &MockObserver{
		notifyFunc: func(ctx context.Context, event AuditEvent) error {
			return nil
		},
	}

	manager.RegisterObserver(observer)

	// Close не очищает список наблюдателей, только вызывает Close() у Closer
	// MockObserver не реализует Closer, поэтому список останется неизменным
	manager.Close()

	assert.Equal(t, 1, len(manager.observers))
}
