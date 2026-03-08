package outbox

import (
	"context"
	"log/slog"

	"github.com/alexdolgov/auth-service/internal/identity/domain"
)

type MockOutbox struct {
	logger *slog.Logger
}

func NewMockOutbox(logger *slog.Logger) *MockOutbox {
	return &MockOutbox{logger: logger}
}

func (m *MockOutbox) Publish(ctx context.Context, event domain.DomainEvent) error {
	m.logger.Info("domain event published",
		"event", event.EventName(),
		"occurred_at", event.OccurredAt(),
	)
	return nil
}
