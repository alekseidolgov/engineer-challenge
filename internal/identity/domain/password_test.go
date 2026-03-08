package domain_test

import (
	"errors"
	"testing"

	"github.com/alexdolgov/auth-service/internal/identity/domain"
)

func TestNewPassword_Valid(t *testing.T) {
	p, err := domain.NewPassword("12345678")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Raw() != "12345678" {
		t.Error("password raw value mismatch")
	}
}

func TestNewPassword_TooShort(t *testing.T) {
	_, err := domain.NewPassword("1234567")
	if !errors.Is(err, domain.ErrPasswordTooShort) {
		t.Errorf("expected ErrPasswordTooShort, got %v", err)
	}
}

func TestNewPasswordPair_Matching(t *testing.T) {
	_, err := domain.NewPasswordPair("abcdefgh", "abcdefgh")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewPasswordPair_Mismatch(t *testing.T) {
	_, err := domain.NewPasswordPair("abcdefgh", "hgfedcba")
	if !errors.Is(err, domain.ErrPasswordMismatch) {
		t.Errorf("expected ErrPasswordMismatch, got %v", err)
	}
}

func TestNewPasswordPair_TooShort(t *testing.T) {
	_, err := domain.NewPasswordPair("short", "short")
	if !errors.Is(err, domain.ErrPasswordTooShort) {
		t.Errorf("expected ErrPasswordTooShort, got %v", err)
	}
}
