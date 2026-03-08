package domain

import (
	"time"

	"github.com/google/uuid"
)

type UserID = uuid.UUID

type User struct {
	ID           UserID
	Email        Email
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewUser(email Email, passwordHash string) *User {
	now := time.Now().UTC()
	return &User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: passwordHash,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

func (u *User) ChangePassword(newHash string) {
	u.PasswordHash = newHash
	u.UpdatedAt = time.Now().UTC()
}
