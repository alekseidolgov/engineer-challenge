package command_test

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/alexdolgov/auth-service/internal/identity/domain"
	"github.com/google/uuid"
)

type memUserRepo struct {
	mu    sync.Mutex
	users map[string]*domain.User
}

func newMemUserRepo() *memUserRepo {
	return &memUserRepo{users: make(map[string]*domain.User)}
}

func (r *memUserRepo) Create(_ context.Context, u *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.users[u.Email.String()] = u
	return nil
}

func (r *memUserRepo) Update(_ context.Context, u *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.users[u.Email.String()] = u
	return nil
}

func (r *memUserRepo) ExistsByEmail(_ context.Context, email domain.Email) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, ok := r.users[email.String()]
	return ok, nil
}

func (r *memUserRepo) FindByEmail(_ context.Context, email domain.Email) (*domain.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	u, ok := r.users[email.String()]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return u, nil
}

func (r *memUserRepo) FindByID(_ context.Context, id domain.UserID) (*domain.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, u := range r.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, domain.ErrUserNotFound
}

type memSessionRepo struct {
	mu       sync.Mutex
	sessions map[uuid.UUID]*domain.Session
}

func newMemSessionRepo() *memSessionRepo {
	return &memSessionRepo{sessions: make(map[uuid.UUID]*domain.Session)}
}

func (r *memSessionRepo) Create(_ context.Context, s *domain.Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sessions[s.ID] = s
	return nil
}

func (r *memSessionRepo) Update(_ context.Context, s *domain.Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sessions[s.ID] = s
	return nil
}

func (r *memSessionRepo) RevokeByID(_ context.Context, id domain.SessionID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.sessions[id]
	if ok {
		s.Revoke()
	}
	return nil
}

func (r *memSessionRepo) RevokeAllByUserID(_ context.Context, userID domain.UserID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, s := range r.sessions {
		if s.UserID == userID {
			s.Revoke()
		}
	}
	return nil
}

func (r *memSessionRepo) FindByID(_ context.Context, id domain.SessionID) (*domain.Session, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.sessions[id]
	if !ok {
		return nil, domain.ErrSessionNotFound
	}
	return s, nil
}

func (r *memSessionRepo) FindByRefreshToken(_ context.Context, token string) (*domain.Session, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, s := range r.sessions {
		if s.RefreshToken == token {
			return s, nil
		}
	}
	return nil, domain.ErrSessionNotFound
}

type memResetTokenRepo struct {
	mu     sync.Mutex
	tokens map[uuid.UUID]*domain.ResetToken
}

func newMemResetTokenRepo() *memResetTokenRepo {
	return &memResetTokenRepo{tokens: make(map[uuid.UUID]*domain.ResetToken)}
}

func (r *memResetTokenRepo) Create(_ context.Context, t *domain.ResetToken) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tokens[t.ID] = t
	return nil
}

func (r *memResetTokenRepo) MarkUsed(_ context.Context, id domain.ResetTokenID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.tokens[id]
	if ok {
		now := time.Now().UTC()
		t.UsedAt = &now
	}
	return nil
}

func (r *memResetTokenRepo) FindByTokenHash(_ context.Context, hash string) (*domain.ResetToken, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, t := range r.tokens {
		if t.TokenHash == hash {
			return t, nil
		}
	}
	return nil, domain.ErrResetTokenNotFound
}

func (r *memResetTokenRepo) FindLatestByUserID(_ context.Context, userID domain.UserID) (*domain.ResetToken, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var latest *domain.ResetToken
	for _, t := range r.tokens {
		if t.UserID == userID {
			if latest == nil || t.CreatedAt.After(latest.CreatedAt) {
				latest = t
			}
		}
	}
	if latest == nil {
		return nil, domain.ErrResetTokenNotFound
	}
	return latest, nil
}

type fakeHasher struct{}

func (f *fakeHasher) Hash(password string) (string, error) {
	return fmt.Sprintf("hashed:%s", password), nil
}

func (f *fakeHasher) Verify(password, hash string) (bool, error) {
	return hash == fmt.Sprintf("hashed:%s", password), nil
}

type fakeTokenIssuer struct{}

func (f *fakeTokenIssuer) Issue(userID domain.UserID, email string) (string, error) {
	return fmt.Sprintf("jwt:%s:%s", userID, email), nil
}

func (f *fakeTokenIssuer) Validate(token string) (*domain.AccessTokenClaims, error) {
	return nil, fmt.Errorf("not implemented in tests")
}

type noopPublisher struct{}

func (n *noopPublisher) Publish(_ context.Context, _ domain.DomainEvent) error { return nil }
