package command_test

import (
	"context"
	"errors"
	"testing"

	"github.com/alexdolgov/auth-service/internal/identity/application/command"
	"github.com/alexdolgov/auth-service/internal/identity/domain"
)

func TestRegister_Success(t *testing.T) {
	repo := newMemUserRepo()
	h := command.NewRegisterHandler(repo, repo, &fakeHasher{}, &noopPublisher{})

	user, err := h.Handle(context.Background(), command.RegisterUser{
		Email:           "test@example.com",
		Password:        "password123",
		PasswordConfirm: "password123",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Email.String() != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", user.Email.String())
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	repo := newMemUserRepo()
	h := command.NewRegisterHandler(repo, repo, &fakeHasher{}, &noopPublisher{})

	_, _ = h.Handle(context.Background(), command.RegisterUser{
		Email: "dup@example.com", Password: "password123", PasswordConfirm: "password123",
	})

	_, err := h.Handle(context.Background(), command.RegisterUser{
		Email: "dup@example.com", Password: "password123", PasswordConfirm: "password123",
	})
	if !errors.Is(err, domain.ErrEmailAlreadyTaken) {
		t.Errorf("expected ErrEmailAlreadyTaken, got %v", err)
	}
}

func TestRegister_InvalidEmail(t *testing.T) {
	repo := newMemUserRepo()
	h := command.NewRegisterHandler(repo, repo, &fakeHasher{}, &noopPublisher{})

	_, err := h.Handle(context.Background(), command.RegisterUser{
		Email: "not-an-email", Password: "password123", PasswordConfirm: "password123",
	})
	if !errors.Is(err, domain.ErrInvalidEmailFormat) {
		t.Errorf("expected ErrInvalidEmailFormat, got %v", err)
	}
}

func TestRegister_PasswordMismatch(t *testing.T) {
	repo := newMemUserRepo()
	h := command.NewRegisterHandler(repo, repo, &fakeHasher{}, &noopPublisher{})

	_, err := h.Handle(context.Background(), command.RegisterUser{
		Email: "test@example.com", Password: "password123", PasswordConfirm: "different",
	})
	if !errors.Is(err, domain.ErrPasswordMismatch) {
		t.Errorf("expected ErrPasswordMismatch, got %v", err)
	}
}

func TestRegister_PasswordTooShort(t *testing.T) {
	repo := newMemUserRepo()
	h := command.NewRegisterHandler(repo, repo, &fakeHasher{}, &noopPublisher{})

	_, err := h.Handle(context.Background(), command.RegisterUser{
		Email: "test@example.com", Password: "short", PasswordConfirm: "short",
	})
	if !errors.Is(err, domain.ErrPasswordTooShort) {
		t.Errorf("expected ErrPasswordTooShort, got %v", err)
	}
}
