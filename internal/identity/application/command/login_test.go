package command_test

import (
	"context"
	"errors"
	"testing"

	"github.com/alexdolgov/auth-service/internal/identity/application/command"
	"github.com/alexdolgov/auth-service/internal/identity/domain"
)

func TestLogin_Success(t *testing.T) {
	userRepo := newMemUserRepo()
	sessionRepo := newMemSessionRepo()
	regHandler := command.NewRegisterHandler(userRepo, userRepo, &fakeHasher{}, &noopPublisher{})
	loginHandler := command.NewLoginHandler(userRepo, sessionRepo, &fakeHasher{}, &fakeTokenIssuer{})

	_, _ = regHandler.Handle(context.Background(), command.RegisterUser{
		Email: "user@example.com", Password: "password123", PasswordConfirm: "password123",
	})

	result, err := loginHandler.Handle(context.Background(), command.LoginUser{
		Email: "user@example.com", Password: "password123",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.AccessToken == "" {
		t.Error("expected non-empty access token")
	}
	if result.RefreshToken == "" {
		t.Error("expected non-empty refresh token")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	userRepo := newMemUserRepo()
	sessionRepo := newMemSessionRepo()
	regHandler := command.NewRegisterHandler(userRepo, userRepo, &fakeHasher{}, &noopPublisher{})
	loginHandler := command.NewLoginHandler(userRepo, sessionRepo, &fakeHasher{}, &fakeTokenIssuer{})

	_, _ = regHandler.Handle(context.Background(), command.RegisterUser{
		Email: "user@example.com", Password: "password123", PasswordConfirm: "password123",
	})

	_, err := loginHandler.Handle(context.Background(), command.LoginUser{
		Email: "user@example.com", Password: "wrongpassword",
	})
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLogin_NonexistentUser(t *testing.T) {
	userRepo := newMemUserRepo()
	sessionRepo := newMemSessionRepo()
	loginHandler := command.NewLoginHandler(userRepo, sessionRepo, &fakeHasher{}, &fakeTokenIssuer{})

	_, err := loginHandler.Handle(context.Background(), command.LoginUser{
		Email: "noone@example.com", Password: "password123",
	})
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}
