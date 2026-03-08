package domain_test

import (
	"testing"
	"time"

	"github.com/alexdolgov/auth-service/internal/identity/domain"
	"github.com/google/uuid"
)

func TestNewResetToken_Creates(t *testing.T) {
	pair, err := domain.NewResetToken(uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pair.RawToken == "" {
		t.Error("expected non-empty raw token")
	}
	if pair.Token.TokenHash == "" {
		t.Error("expected non-empty token hash")
	}
	if pair.Token.IsExpired() {
		t.Error("new token should not be expired")
	}
	if pair.Token.IsUsed() {
		t.Error("new token should not be used")
	}
}

func TestResetToken_Validate_Expired(t *testing.T) {
	rt := &domain.ResetToken{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		TokenHash: "test",
		ExpiresAt: time.Now().UTC().Add(-1 * time.Hour),
		CreatedAt: time.Now().UTC().Add(-2 * time.Hour),
	}
	if err := rt.Validate(); err != domain.ErrResetTokenExpired {
		t.Errorf("expected ErrResetTokenExpired, got %v", err)
	}
}

func TestResetToken_Validate_Used(t *testing.T) {
	now := time.Now().UTC()
	rt := &domain.ResetToken{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		TokenHash: "test",
		ExpiresAt: now.Add(1 * time.Hour),
		UsedAt:    &now,
		CreatedAt: now,
	}
	if err := rt.Validate(); err != domain.ErrResetTokenAlreadyUsed {
		t.Errorf("expected ErrResetTokenAlreadyUsed, got %v", err)
	}
}

func TestResetToken_MarkUsed(t *testing.T) {
	pair, _ := domain.NewResetToken(uuid.New())
	pair.Token.MarkUsed()
	if !pair.Token.IsUsed() {
		t.Error("expected token to be marked as used")
	}
}

func TestCanRequestReset_NoPrevious(t *testing.T) {
	if !domain.CanRequestReset(nil) {
		t.Error("expected to allow reset with no previous request")
	}
}

func TestCanRequestReset_CooldownActive(t *testing.T) {
	recent := time.Now().UTC().Add(-30 * time.Second)
	if domain.CanRequestReset(&recent) {
		t.Error("expected cooldown to block reset")
	}
}

func TestCanRequestReset_CooldownExpired(t *testing.T) {
	old := time.Now().UTC().Add(-5 * time.Minute)
	if !domain.CanRequestReset(&old) {
		t.Error("expected reset to be allowed after cooldown")
	}
}

func TestHashToken_Deterministic(t *testing.T) {
	h1 := domain.HashToken("test-token")
	h2 := domain.HashToken("test-token")
	if h1 != h2 {
		t.Error("expected same hash for same input")
	}
}

func TestHashToken_Different(t *testing.T) {
	h1 := domain.HashToken("token-a")
	h2 := domain.HashToken("token-b")
	if h1 == h2 {
		t.Error("expected different hashes for different inputs")
	}
}
