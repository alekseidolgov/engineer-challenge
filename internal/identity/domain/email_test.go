package domain_test

import (
	"testing"

	"github.com/alexdolgov/auth-service/internal/identity/domain"
)

func TestNewEmail_Valid(t *testing.T) {
	cases := []string{
		"user@example.com",
		"USER@Example.COM",
		"  alice@test.org  ",
		"a.b+c@domain.co",
	}
	for _, tc := range cases {
		e, err := domain.NewEmail(tc)
		if err != nil {
			t.Errorf("NewEmail(%q) returned error: %v", tc, err)
		}
		if e.String() == "" {
			t.Errorf("NewEmail(%q) returned empty string", tc)
		}
	}
}

func TestNewEmail_Invalid(t *testing.T) {
	cases := []string{
		"",
		"   ",
		"noatsign",
		"@no-local.com",
		"user@",
		"user@nodot",
		"user@@double.com",
	}
	for _, tc := range cases {
		_, err := domain.NewEmail(tc)
		if err == nil {
			t.Errorf("NewEmail(%q) expected error, got nil", tc)
		}
	}
}

func TestEmail_Equals(t *testing.T) {
	e1, _ := domain.NewEmail("test@example.com")
	e2, _ := domain.NewEmail("TEST@Example.COM")
	if !e1.Equals(e2) {
		t.Error("expected emails to be equal after normalization")
	}
}
