package domain

import "time"

type DomainEvent interface {
	OccurredAt() time.Time
	EventName() string
}

type UserRegistered struct {
	UserID    UserID
	Email     string
	Timestamp time.Time
}

func (e UserRegistered) OccurredAt() time.Time { return e.Timestamp }
func (e UserRegistered) EventName() string     { return "user.registered" }

type PasswordResetRequested struct {
	UserID    UserID
	Email     string
	Timestamp time.Time
}

func (e PasswordResetRequested) OccurredAt() time.Time { return e.Timestamp }
func (e PasswordResetRequested) EventName() string     { return "password_reset.requested" }

type PasswordResetCompleted struct {
	UserID    UserID
	Timestamp time.Time
}

func (e PasswordResetCompleted) OccurredAt() time.Time { return e.Timestamp }
func (e PasswordResetCompleted) EventName() string     { return "password_reset.completed" }
