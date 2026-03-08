package token_test

import (
	"testing"
	"time"

	"github.com/alexdolgov/auth-service/internal/identity/infrastructure/token"
	"github.com/google/uuid"
)

func TestJWTIssuer_IssueAndValidate(t *testing.T) {
	key, err := token.GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey() error: %v", err)
	}
	issuer := token.NewJWTIssuer(key, 15*time.Minute)
	userID := uuid.New()
	email := "test@example.com"

	tok, err := issuer.Issue(userID, email)
	if err != nil {
		t.Fatalf("Issue() error: %v", err)
	}
	if tok == "" {
		t.Fatal("expected non-empty token")
	}

	claims, err := issuer.Validate(tok)
	if err != nil {
		t.Fatalf("Validate() error: %v", err)
	}
	if claims.UserID != userID {
		t.Errorf("expected user ID %s, got %s", userID, claims.UserID)
	}
	if claims.Email != email {
		t.Errorf("expected email %s, got %s", email, claims.Email)
	}
}

func TestJWTIssuer_InvalidToken(t *testing.T) {
	key, _ := token.GenerateKey()
	issuer := token.NewJWTIssuer(key, 15*time.Minute)
	_, err := issuer.Validate("invalid-token")
	if err == nil {
		t.Error("expected error for invalid token")
	}
}

func TestJWTIssuer_WrongKey(t *testing.T) {
	key1, _ := token.GenerateKey()
	key2, _ := token.GenerateKey()
	issuer1 := token.NewJWTIssuer(key1, 15*time.Minute)
	issuer2 := token.NewJWTIssuer(key2, 15*time.Minute)

	tok, _ := issuer1.Issue(uuid.New(), "test@example.com")
	_, err := issuer2.Validate(tok)
	if err == nil {
		t.Error("expected error for wrong key")
	}
}

func TestJWTIssuer_PublicKey(t *testing.T) {
	key, _ := token.GenerateKey()
	issuer := token.NewJWTIssuer(key, 15*time.Minute)
	pub := issuer.PublicKey()
	if pub == nil {
		t.Fatal("expected non-nil public key")
	}
	if pub.Curve == nil {
		t.Error("expected valid curve on public key")
	}
}
