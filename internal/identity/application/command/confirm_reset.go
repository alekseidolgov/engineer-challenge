package command

import (
	"context"
	"time"

	"github.com/alexdolgov/auth-service/internal/identity/domain"
)

type ConfirmPasswordReset struct {
	Token           string
	NewPassword     string
	ConfirmPassword string
}

type ConfirmResetHandler struct {
	users       domain.UserWriteRepository
	userReader  domain.UserReadRepository
	resetTokens domain.ResetTokenWriteRepository
	resetReader domain.ResetTokenReadRepository
	hasher      domain.PasswordHasher
	events      domain.EventPublisher
}

func NewConfirmResetHandler(
	users domain.UserWriteRepository,
	userReader domain.UserReadRepository,
	resetTokens domain.ResetTokenWriteRepository,
	resetReader domain.ResetTokenReadRepository,
	hasher domain.PasswordHasher,
	events domain.EventPublisher,
) *ConfirmResetHandler {
	return &ConfirmResetHandler{
		users:       users,
		userReader:  userReader,
		resetTokens: resetTokens,
		resetReader: resetReader,
		hasher:      hasher,
		events:      events,
	}
}

func (h *ConfirmResetHandler) Handle(ctx context.Context, cmd ConfirmPasswordReset) error {
	if len(cmd.Token) == 0 || len(cmd.Token) > domain.MaxRawTokenLength {
		return domain.ErrResetTokenNotFound
	}

	pwd, err := domain.NewPasswordPair(cmd.NewPassword, cmd.ConfirmPassword)
	if err != nil {
		return err
	}

	tokenHash := domain.HashToken(cmd.Token)
	rt, err := h.resetReader.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		return domain.ErrResetTokenNotFound
	}

	if err := rt.Validate(); err != nil {
		return err
	}

	user, err := h.userReader.FindByID(ctx, rt.UserID)
	if err != nil {
		return domain.ErrUserNotFound
	}

	hash, err := h.hasher.Hash(pwd.Raw())
	if err != nil {
		return err
	}

	user.ChangePassword(hash)
	if err := h.users.Update(ctx, user); err != nil {
		return err
	}

	rt.MarkUsed()
	if err := h.resetTokens.MarkUsed(ctx, rt.ID); err != nil {
		return err
	}

	_ = h.events.Publish(ctx, domain.PasswordResetCompleted{
		UserID:    user.ID,
		Timestamp: time.Now().UTC(),
	})

	return nil
}
