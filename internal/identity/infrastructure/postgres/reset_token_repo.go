package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/alexdolgov/auth-service/internal/identity/domain"
	"github.com/google/uuid"
)

type ResetTokenRepository struct {
	db *sql.DB
}

func NewResetTokenRepository(db *sql.DB) *ResetTokenRepository {
	return &ResetTokenRepository{db: db}
}

func (r *ResetTokenRepository) Create(ctx context.Context, t *domain.ResetToken) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO password_reset_tokens (id, user_id, token_hash, expires_at, created_at) VALUES ($1, $2, $3, $4, $5)`,
		t.ID, t.UserID, t.TokenHash, t.ExpiresAt, t.CreatedAt,
	)
	return err
}

func (r *ResetTokenRepository) MarkUsed(ctx context.Context, id domain.ResetTokenID) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE password_reset_tokens SET used_at = NOW() WHERE id = $1`, id,
	)
	return err
}

func (r *ResetTokenRepository) FindByTokenHash(ctx context.Context, hash string) (*domain.ResetToken, error) {
	return r.scanToken(r.db.QueryRowContext(ctx,
		`SELECT id, user_id, token_hash, expires_at, used_at, created_at FROM password_reset_tokens WHERE token_hash = $1`, hash,
	))
}

func (r *ResetTokenRepository) FindLatestByUserID(ctx context.Context, userID domain.UserID) (*domain.ResetToken, error) {
	return r.scanToken(r.db.QueryRowContext(ctx,
		`SELECT id, user_id, token_hash, expires_at, used_at, created_at FROM password_reset_tokens WHERE user_id = $1 ORDER BY created_at DESC LIMIT 1`, userID,
	))
}

func (r *ResetTokenRepository) scanToken(row *sql.Row) (*domain.ResetToken, error) {
	var t domain.ResetToken
	var id, userID uuid.UUID
	err := row.Scan(&id, &userID, &t.TokenHash, &t.ExpiresAt, &t.UsedAt, &t.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrResetTokenNotFound
	}
	if err != nil {
		return nil, err
	}
	t.ID = id
	t.UserID = userID
	return &t, nil
}
