package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"backify/services/auth/internal/domain"
)

// PasswordResetRepo implement port.PasswordResetRepository.
type PasswordResetRepo struct {
	conns *ConnManager
}

func NewPasswordResetRepo(conns *ConnManager) *PasswordResetRepo {
	return &PasswordResetRepo{conns: conns}
}

func (r *PasswordResetRepo) Create(ctx context.Context, projectID string, token *domain.PasswordResetToken) error {
	pool, err := r.conns.Pool(ctx, projectID)
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO password_reset_tokens (id, project_id, user_id, token_hash, expires_at, used_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, token.ID, token.ProjectID, token.UserID, token.TokenHash, token.ExpiresAt, token.UsedAt, token.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert password reset token: %w", err)
	}
	return nil
}

func (r *PasswordResetRepo) GetByTokenHash(ctx context.Context, projectID, tokenHash string) (*domain.PasswordResetToken, error) {
	pool, err := r.conns.Pool(ctx, projectID)
	if err != nil {
		return nil, err
	}

	var t domain.PasswordResetToken
	err = pool.QueryRow(ctx, `
		SELECT id, project_id, user_id, token_hash, expires_at, used_at, created_at
		FROM password_reset_tokens WHERE token_hash = $1
	`, tokenHash).Scan(&t.ID, &t.ProjectID, &t.UserID, &t.TokenHash, &t.ExpiresAt, &t.UsedAt, &t.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTokenInvalid
		}
		return nil, fmt.Errorf("scan password reset token: %w", err)
	}
	return &t, nil
}

func (r *PasswordResetRepo) Update(ctx context.Context, projectID string, token *domain.PasswordResetToken) error {
	pool, err := r.conns.Pool(ctx, projectID)
	if err != nil {
		return err
	}

	tag, err := pool.Exec(ctx, `
		UPDATE password_reset_tokens SET used_at = $1 WHERE id = $2
	`, token.UsedAt, token.ID)
	if err != nil {
		return fmt.Errorf("update password reset token: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrTokenInvalid
	}
	return nil
}
