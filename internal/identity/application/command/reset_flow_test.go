package command_test

import (
	"context"
	"errors"
	"testing"

	"github.com/alexdolgov/auth-service/internal/identity/application/command"
	"github.com/alexdolgov/auth-service/internal/identity/domain"
)

func setupResetTestEnv(t *testing.T) (
	*memUserRepo, *memResetTokenRepo,
	*command.RequestResetHandler, *command.ConfirmResetHandler,
) {
	t.Helper()
	userRepo := newMemUserRepo()
	resetRepo := newMemResetTokenRepo()

	regHandler := command.NewRegisterHandler(userRepo, userRepo, &fakeHasher{}, &noopPublisher{})
	_, err := regHandler.Handle(context.Background(), command.RegisterUser{
		Email: "user@example.com", Password: "oldpass12", PasswordConfirm: "oldpass12",
	})
	if err != nil {
		t.Fatalf("setup: register failed: %v", err)
	}

	requestHandler := command.NewRequestResetHandler(userRepo, resetRepo, resetRepo, &noopPublisher{})
	confirmHandler := command.NewConfirmResetHandler(userRepo, userRepo, resetRepo, resetRepo, &fakeHasher{}, &noopPublisher{})

	return userRepo, resetRepo, requestHandler, confirmHandler
}

func TestResetFlow_FullCycle(t *testing.T) {
	_, _, requestHandler, confirmHandler := setupResetTestEnv(t)

	rawToken, err := requestHandler.Handle(context.Background(), command.RequestPasswordReset{
		Email: "user@example.com",
	})
	if err != nil {
		t.Fatalf("request reset error: %v", err)
	}
	if rawToken == "" {
		t.Fatal("expected non-empty raw token")
	}

	err = confirmHandler.Handle(context.Background(), command.ConfirmPasswordReset{
		Token:           rawToken,
		NewPassword:     "newpass12",
		ConfirmPassword: "newpass12",
	})
	if err != nil {
		t.Fatalf("confirm reset error: %v", err)
	}
}

func TestResetFlow_UnknownEmail(t *testing.T) {
	userRepo := newMemUserRepo()
	resetRepo := newMemResetTokenRepo()
	requestHandler := command.NewRequestResetHandler(userRepo, resetRepo, resetRepo, &noopPublisher{})

	_, err := requestHandler.Handle(context.Background(), command.RequestPasswordReset{
		Email: "unknown@example.com",
	})
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestResetFlow_PasswordMismatch(t *testing.T) {
	_, _, requestHandler, confirmHandler := setupResetTestEnv(t)

	rawToken, _ := requestHandler.Handle(context.Background(), command.RequestPasswordReset{
		Email: "user@example.com",
	})

	err := confirmHandler.Handle(context.Background(), command.ConfirmPasswordReset{
		Token:           rawToken,
		NewPassword:     "newpass12",
		ConfirmPassword: "different",
	})
	if !errors.Is(err, domain.ErrPasswordMismatch) {
		t.Errorf("expected ErrPasswordMismatch, got %v", err)
	}
}

func TestResetFlow_PasswordTooShort(t *testing.T) {
	_, _, requestHandler, confirmHandler := setupResetTestEnv(t)

	rawToken, _ := requestHandler.Handle(context.Background(), command.RequestPasswordReset{
		Email: "user@example.com",
	})

	err := confirmHandler.Handle(context.Background(), command.ConfirmPasswordReset{
		Token:           rawToken,
		NewPassword:     "short",
		ConfirmPassword: "short",
	})
	if !errors.Is(err, domain.ErrPasswordTooShort) {
		t.Errorf("expected ErrPasswordTooShort, got %v", err)
	}
}

func TestResetFlow_TokenReuse(t *testing.T) {
	_, _, requestHandler, confirmHandler := setupResetTestEnv(t)

	rawToken, _ := requestHandler.Handle(context.Background(), command.RequestPasswordReset{
		Email: "user@example.com",
	})

	_ = confirmHandler.Handle(context.Background(), command.ConfirmPasswordReset{
		Token: rawToken, NewPassword: "newpass12", ConfirmPassword: "newpass12",
	})

	err := confirmHandler.Handle(context.Background(), command.ConfirmPasswordReset{
		Token: rawToken, NewPassword: "another1", ConfirmPassword: "another1",
	})
	if !errors.Is(err, domain.ErrResetTokenAlreadyUsed) {
		t.Errorf("expected ErrResetTokenAlreadyUsed, got %v", err)
	}
}

func TestResetFlow_InvalidToken(t *testing.T) {
	_, _, _, confirmHandler := setupResetTestEnv(t)

	err := confirmHandler.Handle(context.Background(), command.ConfirmPasswordReset{
		Token: "nonexistent-token", NewPassword: "newpass12", ConfirmPassword: "newpass12",
	})
	if !errors.Is(err, domain.ErrResetTokenNotFound) {
		t.Errorf("expected ErrResetTokenNotFound, got %v", err)
	}
}
