package domain_test

import (
	"testing"
	"time"

	"github.com/alexdolgov/auth-service/internal/identity/domain"
	"github.com/google/uuid"
)

func TestNewSession_Valid(t *testing.T) {
	s, err := domain.NewSession(uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.RefreshToken == "" {
		t.Error("expected non-empty refresh token")
	}
	if s.IsExpired() {
		t.Error("new session should not be expired")
	}
	if s.IsRevoked() {
		t.Error("new session should not be revoked")
	}
	if !s.IsValid() {
		t.Error("new session should be valid")
	}
}

func TestSession_Expired(t *testing.T) {
	s := &domain.Session{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		ExpiresAt: time.Now().UTC().Add(-1 * time.Hour),
		CreatedAt: time.Now().UTC().Add(-2 * time.Hour),
	}
	if !s.IsExpired() {
		t.Error("session should be expired")
	}
	if s.IsValid() {
		t.Error("expired session should not be valid")
	}
}

func TestSession_Revoke(t *testing.T) {
	s, _ := domain.NewSession(uuid.New())
	s.Revoke()
	if !s.IsRevoked() {
		t.Error("session should be revoked")
	}
	if s.IsValid() {
		t.Error("revoked session should not be valid")
	}
}
