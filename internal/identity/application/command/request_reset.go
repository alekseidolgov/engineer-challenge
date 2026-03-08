package command

import (
	"context"
	"time"

	"github.com/alexdolgov/auth-service/internal/identity/domain"
)

type RequestPasswordReset struct {
	Email string
}

type RequestResetHandler struct {
	readers     domain.UserReadRepository
	resetTokens domain.ResetTokenWriteRepository
	resetReader domain.ResetTokenReadRepository
	events      domain.EventPublisher
}

func NewRequestResetHandler(
	readers domain.UserReadRepository,
	resetTokens domain.ResetTokenWriteRepository,
	resetReader domain.ResetTokenReadRepository,
	events domain.EventPublisher,
) *RequestResetHandler {
	return &RequestResetHandler{
		readers:     readers,
		resetTokens: resetTokens,
		resetReader: resetReader,
		events:      events,
	}
}

func (h *RequestResetHandler) Handle(ctx context.Context, cmd RequestPasswordReset) (string, error) {
	email, err := domain.NewEmail(cmd.Email)
	if err != nil {
		return "", domain.ErrUserNotFound
	}

	user, err := h.readers.FindByEmail(ctx, email)
	if err != nil {
		return "", domain.ErrUserNotFound
	}

	latest, _ := h.resetReader.FindLatestByUserID(ctx, user.ID)
	if latest != nil && !domain.CanRequestReset(&latest.CreatedAt) {
		return "", domain.ErrResetCooldown
	}

	pair, err := domain.NewResetToken(user.ID)
	if err != nil {
		return "", err
	}

	if err := h.resetTokens.Create(ctx, pair.Token); err != nil {
		return "", err
	}

	_ = h.events.Publish(ctx, domain.PasswordResetRequested{
		UserID:    user.ID,
		Email:     email.String(),
		Timestamp: time.Now().UTC(),
	})

	return pair.RawToken, nil
}
