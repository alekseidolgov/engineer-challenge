package domain

import "context"

type UserWriteRepository interface {
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	ExistsByEmail(ctx context.Context, email Email) (bool, error)
}

type UserReadRepository interface {
	FindByEmail(ctx context.Context, email Email) (*User, error)
	FindByID(ctx context.Context, id UserID) (*User, error)
}

type SessionWriteRepository interface {
	Create(ctx context.Context, session *Session) error
	Update(ctx context.Context, session *Session) error
	RevokeByID(ctx context.Context, id SessionID) error
	RevokeAllByUserID(ctx context.Context, userID UserID) error
}

type SessionReadRepository interface {
	FindByID(ctx context.Context, id SessionID) (*Session, error)
	FindByRefreshToken(ctx context.Context, token string) (*Session, error)
}

type ResetTokenWriteRepository interface {
	Create(ctx context.Context, token *ResetToken) error
	MarkUsed(ctx context.Context, id ResetTokenID) error
}

type ResetTokenReadRepository interface {
	FindByTokenHash(ctx context.Context, hash string) (*ResetToken, error)
	FindLatestByUserID(ctx context.Context, userID UserID) (*ResetToken, error)
}

type EventPublisher interface {
	Publish(ctx context.Context, event DomainEvent) error
}
