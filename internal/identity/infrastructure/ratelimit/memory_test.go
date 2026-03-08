package ratelimit_test

import (
	"testing"
	"time"

	"github.com/alexdolgov/auth-service/internal/identity/domain"
	"github.com/alexdolgov/auth-service/internal/identity/infrastructure/ratelimit"
)

func TestInMemoryLimiter_ImplementsInterface(t *testing.T) {
	var _ domain.RateLimiter = ratelimit.NewInMemoryLimiter(5, time.Minute)
}

func TestInMemoryLimiter_AllowsWithinLimit(t *testing.T) {
	lim := ratelimit.NewInMemoryLimiter(3, time.Minute)
	for i := 0; i < 3; i++ {
		if !lim.Allow("key") {
			t.Errorf("attempt %d should be allowed", i+1)
		}
	}
}

func TestInMemoryLimiter_BlocksOverLimit(t *testing.T) {
	lim := ratelimit.NewInMemoryLimiter(2, time.Minute)
	lim.Allow("key")
	lim.Allow("key")
	if lim.Allow("key") {
		t.Error("3rd attempt should be blocked")
	}
}

func TestInMemoryLimiter_Reset(t *testing.T) {
	lim := ratelimit.NewInMemoryLimiter(1, time.Minute)
	lim.Allow("key")
	lim.Reset("key")
	if !lim.Allow("key") {
		t.Error("should allow after reset")
	}
}

func TestInMemoryLimiter_IndependentKeys(t *testing.T) {
	lim := ratelimit.NewInMemoryLimiter(1, time.Minute)
	lim.Allow("a")
	if !lim.Allow("b") {
		t.Error("different keys should be independent")
	}
}
