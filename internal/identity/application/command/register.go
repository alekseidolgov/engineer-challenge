package command

import (
	"context"
	"time"

	"github.com/alexdolgov/auth-service/internal/identity/domain"
)

type RegisterUser struct {
	Email           string
	Password        string
	PasswordConfirm string
}

type RegisterHandler struct {
	users    domain.UserWriteRepository
	readers  domain.UserReadRepository
	hasher   domain.PasswordHasher
	events   domain.EventPublisher
}

func NewRegisterHandler(
	users domain.UserWriteRepository,
	readers domain.UserReadRepository,
	hasher domain.PasswordHasher,
	events domain.EventPublisher,
) *RegisterHandler {
	return &RegisterHandler{users: users, readers: readers, hasher: hasher, events: events}
}

func (h *RegisterHandler) Handle(ctx context.Context, cmd RegisterUser) (*domain.User, error) {
	email, err := domain.NewEmail(cmd.Email)
	if err != nil {
		return nil, err
	}

	exists, err := h.users.ExistsByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domain.ErrEmailAlreadyTaken
	}

	pwd, err := domain.NewPasswordPair(cmd.Password, cmd.PasswordConfirm)
	if err != nil {
		return nil, err
	}

	hash, err := h.hasher.Hash(pwd.Raw())
	if err != nil {
		return nil, err
	}

	user := domain.NewUser(email, hash)
	if err := h.users.Create(ctx, user); err != nil {
		return nil, err
	}

	_ = h.events.Publish(ctx, domain.UserRegistered{
		UserID:    user.ID,
		Email:     email.String(),
		Timestamp: time.Now().UTC(),
	})

	return user, nil
}
