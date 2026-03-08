package command

import (
	"context"

	"github.com/alexdolgov/auth-service/internal/identity/domain"
)

type LogoutUser struct {
	SessionID domain.SessionID
}

type LogoutHandler struct {
	sessions domain.SessionWriteRepository
}

func NewLogoutHandler(sessions domain.SessionWriteRepository) *LogoutHandler {
	return &LogoutHandler{sessions: sessions}
}

func (h *LogoutHandler) Handle(ctx context.Context, cmd LogoutUser) error {
	return h.sessions.RevokeByID(ctx, cmd.SessionID)
}
