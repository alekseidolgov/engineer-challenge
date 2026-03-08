package hasher_test

import (
	"testing"

	"github.com/alexdolgov/auth-service/internal/identity/infrastructure/hasher"
)

func TestArgon2Hasher_HashAndVerify(t *testing.T) {
	h := hasher.NewArgon2Hasher()
	password := "securepassword123"

	hash, err := h.Hash(password)
	if err != nil {
		t.Fatalf("Hash() error: %v", err)
	}

	ok, err := h.Verify(password, hash)
	if err != nil {
		t.Fatalf("Verify() error: %v", err)
	}
	if !ok {
		t.Error("Verify() should return true for correct password")
	}
}

func TestArgon2Hasher_WrongPassword(t *testing.T) {
	h := hasher.NewArgon2Hasher()
	hash, _ := h.Hash("correctpassword")

	ok, err := h.Verify("wrongpassword", hash)
	if err != nil {
		t.Fatalf("Verify() error: %v", err)
	}
	if ok {
		t.Error("Verify() should return false for wrong password")
	}
}

func TestArgon2Hasher_UniqueHashes(t *testing.T) {
	h := hasher.NewArgon2Hasher()
	h1, _ := h.Hash("samepassword")
	h2, _ := h.Hash("samepassword")
	if h1 == h2 {
		t.Error("hashes should differ due to random salt")
	}
}
