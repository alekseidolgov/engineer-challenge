package command_test

import (
	"context"
	"testing"

	"github.com/alexdolgov/auth-service/internal/identity/application/command"
	"github.com/alexdolgov/auth-service/internal/identity/domain"
)

func setupRefreshTest(t *testing.T) (
	*command.RefreshTokenHandler,
	*memUserRepo,
	*memSessionRepo,
	*command.LoginHandler,
) {
	t.Helper()
	userRepo := newMemUserRepo()
	sessionRepo := newMemSessionRepo()
	hasher := &fakeHasher{}
	issuer := &fakeTokenIssuer{}

	regHandler := command.NewRegisterHandler(userRepo, userRepo, hasher, &noopPublisher{})
	loginHandler := command.NewLoginHandler(userRepo, sessionRepo, hasher, issuer)
	refreshHandler := command.NewRefreshTokenHandler(sessionRepo, sessionRepo, userRepo, issuer)

	_, err := regHandler.Handle(context.Background(), command.RegisterUser{
		Email: "refresh@test.com", Password: "password123", PasswordConfirm: "password123",
	})
	if err != nil {
		t.Fatal(err)
	}

	return refreshHandler, userRepo, sessionRepo, loginHandler
}

func TestRefreshToken_Success(t *testing.T) {
	refreshHandler, _, _, loginHandler := setupRefreshTest(t)

	loginResult, err := loginHandler.Handle(context.Background(), command.LoginUser{
		Email: "refresh@test.com", Password: "password123",
	})
	if err != nil {
		t.Fatal(err)
	}

	result, err := refreshHandler.Handle(context.Background(), command.RefreshToken{
		Token: loginResult.RefreshToken,
	})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if result.AccessToken == "" {
		t.Fatal("expected access token")
	}
	if result.RefreshToken == loginResult.RefreshToken {
		t.Fatal("expected rotated refresh token, got same one")
	}
	if result.SessionID != loginResult.SessionID {
		t.Fatal("expected same session ID")
	}
}

func TestRefreshToken_OldTokenInvalid(t *testing.T) {
	refreshHandler, _, _, loginHandler := setupRefreshTest(t)

	loginResult, _ := loginHandler.Handle(context.Background(), command.LoginUser{
		Email: "refresh@test.com", Password: "password123",
	})
	oldToken := loginResult.RefreshToken

	_, err := refreshHandler.Handle(context.Background(), command.RefreshToken{Token: oldToken})
	if err != nil {
		t.Fatal(err)
	}

	// Старый токен после ротации не должен работать
	_, err = refreshHandler.Handle(context.Background(), command.RefreshToken{Token: oldToken})
	if err == nil {
		t.Fatal("expected error for reused token")
	}
}

func TestRefreshToken_RevokedSession(t *testing.T) {
	refreshHandler, _, sessionRepo, loginHandler := setupRefreshTest(t)

	loginResult, _ := loginHandler.Handle(context.Background(), command.LoginUser{
		Email: "refresh@test.com", Password: "password123",
	})

	_ = sessionRepo.RevokeByID(context.Background(), loginResult.SessionID)

	_, err := refreshHandler.Handle(context.Background(), command.RefreshToken{
		Token: loginResult.RefreshToken,
	})
	if err != domain.ErrSessionExpired {
		t.Fatalf("expected ErrSessionExpired, got %v", err)
	}
}

func TestRefreshToken_InvalidToken(t *testing.T) {
	refreshHandler, _, _, _ := setupRefreshTest(t)

	_, err := refreshHandler.Handle(context.Background(), command.RefreshToken{
		Token: "nonexistent-token",
	})
	if err != domain.ErrSessionNotFound {
		t.Fatalf("expected ErrSessionNotFound, got %v", err)
	}
}
