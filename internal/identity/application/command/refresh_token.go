package command

import (
	"context"

	"github.com/alexdolgov/auth-service/internal/identity/domain"
)

type RefreshToken struct {
	Token string
}

type RefreshResult struct {
	AccessToken  string
	RefreshToken string
	SessionID    domain.SessionID
}

type RefreshTokenHandler struct {
	sessions    domain.SessionWriteRepository
	sessReader  domain.SessionReadRepository
	userReader  domain.UserReadRepository
	tokenIssuer domain.AccessTokenIssuer
}

func NewRefreshTokenHandler(
	sessions domain.SessionWriteRepository,
	sessReader domain.SessionReadRepository,
	userReader domain.UserReadRepository,
	tokenIssuer domain.AccessTokenIssuer,
) *RefreshTokenHandler {
	return &RefreshTokenHandler{
		sessions:    sessions,
		sessReader:  sessReader,
		userReader:  userReader,
		tokenIssuer: tokenIssuer,
	}
}

func (h *RefreshTokenHandler) Handle(ctx context.Context, cmd RefreshToken) (*RefreshResult, error) {
	if len(cmd.Token) == 0 || len(cmd.Token) > domain.MaxRawTokenLength {
		return nil, domain.ErrSessionNotFound
	}

	session, err := h.sessReader.FindByRefreshToken(ctx, cmd.Token)
	if err != nil {
		return nil, domain.ErrSessionNotFound
	}

	if !session.IsValid() {
		// Token reuse на отозванной/expired сессии — возможна компрометация.
		// Отзываем все сессии пользователя как мера предосторожности.
		_ = h.sessions.RevokeAllByUserID(ctx, session.UserID)
		return nil, domain.ErrSessionExpired
	}

	user, err := h.userReader.FindByID(ctx, session.UserID)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}

	accessToken, err := h.tokenIssuer.Issue(user.ID, user.Email.String())
	if err != nil {
		return nil, err
	}

	newRefresh, err := session.Rotate()
	if err != nil {
		return nil, err
	}
	if err := h.sessions.Update(ctx, session); err != nil {
		return nil, err
	}

	return &RefreshResult{
		AccessToken:  accessToken,
		RefreshToken: newRefresh,
		SessionID:    session.ID,
	}, nil
}
