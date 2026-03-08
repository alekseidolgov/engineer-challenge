package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/alexdolgov/auth-service/internal/identity/domain"
	"github.com/google/uuid"
)

type SessionRepository struct {
	db *sql.DB
}

func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Create(ctx context.Context, s *domain.Session) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO sessions (id, user_id, refresh_token, expires_at, created_at) VALUES ($1, $2, $3, $4, $5)`,
		s.ID, s.UserID, s.RefreshToken, s.ExpiresAt, s.CreatedAt,
	)
	return err
}

func (r *SessionRepository) Update(ctx context.Context, s *domain.Session) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE sessions SET refresh_token = $1 WHERE id = $2`,
		s.RefreshToken, s.ID,
	)
	return err
}

func (r *SessionRepository) RevokeByID(ctx context.Context, id domain.SessionID) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE sessions SET revoked_at = NOW() WHERE id = $1 AND revoked_at IS NULL`, id,
	)
	return err
}

func (r *SessionRepository) RevokeAllByUserID(ctx context.Context, userID domain.UserID) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE sessions SET revoked_at = NOW() WHERE user_id = $1 AND revoked_at IS NULL`, userID,
	)
	return err
}

func (r *SessionRepository) FindByID(ctx context.Context, id domain.SessionID) (*domain.Session, error) {
	return r.scanSession(r.db.QueryRowContext(ctx,
		`SELECT id, user_id, refresh_token, expires_at, created_at, revoked_at FROM sessions WHERE id = $1`, id,
	))
}

func (r *SessionRepository) FindByRefreshToken(ctx context.Context, token string) (*domain.Session, error) {
	return r.scanSession(r.db.QueryRowContext(ctx,
		`SELECT id, user_id, refresh_token, expires_at, created_at, revoked_at FROM sessions WHERE refresh_token = $1`, token,
	))
}

func (r *SessionRepository) scanSession(row *sql.Row) (*domain.Session, error) {
	var s domain.Session
	var id, userID uuid.UUID
	err := row.Scan(&id, &userID, &s.RefreshToken, &s.ExpiresAt, &s.CreatedAt, &s.RevokedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrSessionNotFound
	}
	if err != nil {
		return nil, err
	}
	s.ID = id
	s.UserID = userID
	return &s, nil
}
