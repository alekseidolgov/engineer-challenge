package command

import (
	"context"

	"github.com/alexdolgov/auth-service/internal/identity/domain"
)

type LoginUser struct {
	Email    string
	Password string
}

type LoginResult struct {
	AccessToken  string
	RefreshToken string
	SessionID    domain.SessionID
}

type LoginHandler struct {
	readers  domain.UserReadRepository
	sessions domain.SessionWriteRepository
	hasher   domain.PasswordHasher
	tokens   domain.AccessTokenIssuer
}

func NewLoginHandler(
	readers domain.UserReadRepository,
	sessions domain.SessionWriteRepository,
	hasher domain.PasswordHasher,
	tokens domain.AccessTokenIssuer,
) *LoginHandler {
	return &LoginHandler{readers: readers, sessions: sessions, hasher: hasher, tokens: tokens}
}

func (h *LoginHandler) Handle(ctx context.Context, cmd LoginUser) (*LoginResult, error) {
	email, err := domain.NewEmail(cmd.Email)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	user, err := h.readers.FindByEmail(ctx, email)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	ok, err := h.hasher.Verify(cmd.Password, user.PasswordHash)
	if err != nil || !ok {
		return nil, domain.ErrInvalidCredentials
	}

	session, err := domain.NewSession(user.ID)
	if err != nil {
		return nil, err
	}
	if err := h.sessions.Create(ctx, session); err != nil {
		return nil, err
	}

	accessToken, err := h.tokens.Issue(user.ID, user.Email.String())
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		AccessToken:  accessToken,
		RefreshToken: session.RefreshToken,
		SessionID:    session.ID,
	}, nil
}
