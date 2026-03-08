package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
)

const (
	ResetTokenTTL       = 30 * time.Minute
	ResetTokenLen       = 32
	ResetCooldownPeriod = 2 * time.Minute
	MaxRawTokenLength   = 128 // hex-encoded 64 bytes; защита от DoS до SHA-256
)

type ResetTokenID = uuid.UUID

type ResetToken struct {
	ID        ResetTokenID
	UserID    UserID
	TokenHash string
	ExpiresAt time.Time
	UsedAt    *time.Time
	CreatedAt time.Time
}

type ResetTokenPair struct {
	Token    *ResetToken
	RawToken string
}

func NewResetToken(userID UserID) (*ResetTokenPair, error) {
	raw, err := generateSecureToken(ResetTokenLen)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	token := &ResetToken{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: HashToken(raw),
		ExpiresAt: now.Add(ResetTokenTTL),
		CreatedAt: now,
	}
	return &ResetTokenPair{Token: token, RawToken: raw}, nil
}

func HashToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

func (rt *ResetToken) IsExpired() bool {
	return time.Now().UTC().After(rt.ExpiresAt)
}

func (rt *ResetToken) IsUsed() bool {
	return rt.UsedAt != nil
}

func (rt *ResetToken) Validate() error {
	if rt.IsUsed() {
		return ErrResetTokenAlreadyUsed
	}
	if rt.IsExpired() {
		return ErrResetTokenExpired
	}
	return nil
}

func (rt *ResetToken) MarkUsed() {
	now := time.Now().UTC()
	rt.UsedAt = &now
}

func CanRequestReset(lastCreatedAt *time.Time) bool {
	if lastCreatedAt == nil {
		return true
	}
	return time.Since(*lastCreatedAt) >= ResetCooldownPeriod
}
