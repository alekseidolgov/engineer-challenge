package domain

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
)

const (
	AccessTokenTTL  = 15 * time.Minute
	RefreshTokenTTL = 7 * 24 * time.Hour
	RefreshTokenLen = 32
)

type SessionID = uuid.UUID

type Session struct {
	ID           SessionID
	UserID       UserID
	RefreshToken string
	ExpiresAt    time.Time
	CreatedAt    time.Time
	RevokedAt    *time.Time
}

func NewSession(userID UserID) (*Session, error) {
	token, err := generateSecureToken(RefreshTokenLen)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	return &Session{
		ID:           uuid.New(),
		UserID:       userID,
		RefreshToken: token,
		ExpiresAt:    now.Add(RefreshTokenTTL),
		CreatedAt:    now,
	}, nil
}

func (s *Session) IsExpired() bool {
	return time.Now().UTC().After(s.ExpiresAt)
}

func (s *Session) IsRevoked() bool {
	return s.RevokedAt != nil
}

func (s *Session) IsValid() bool {
	return !s.IsExpired() && !s.IsRevoked()
}

func (s *Session) Revoke() {
	now := time.Now().UTC()
	s.RevokedAt = &now
}

func (s *Session) Rotate() (string, error) {
	token, err := generateSecureToken(RefreshTokenLen)
	if err != nil {
		return "", err
	}
	s.RefreshToken = token
	return token, nil
}

func generateSecureToken(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
