package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"backify/services/auth/internal/domain"
)

// RefreshTokenRepo implement port.RefreshTokenRepository.
type RefreshTokenRepo struct {
	conns *ConnManager
}

func NewRefreshTokenRepo(conns *ConnManager) *RefreshTokenRepo {
	return &RefreshTokenRepo{conns: conns}
}

func (r *RefreshTokenRepo) Create(ctx context.Context, projectID string, token *domain.RefreshToken) error {
	pool, err := r.conns.Pool(ctx, projectID)
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO refresh_tokens (id, project_id, user_id, family_id, token_hash, expires_at, used_at, revoked_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, token.ID, token.ProjectID, token.UserID, token.FamilyID, token.TokenHash, token.ExpiresAt, token.UsedAt, token.RevokedAt, token.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert refresh token: %w", err)
	}
	return nil
}

func (r *RefreshTokenRepo) GetByTokenHash(ctx context.Context, projectID, tokenHash string) (*domain.RefreshToken, error) {
	pool, err := r.conns.Pool(ctx, projectID)
	if err != nil {
		return nil, err
	}

	var t domain.RefreshToken
	err = pool.QueryRow(ctx, `
		SELECT id, project_id, user_id, family_id, token_hash, expires_at, used_at, revoked_at, created_at
		FROM refresh_tokens WHERE token_hash = $1
	`, tokenHash).Scan(&t.ID, &t.ProjectID, &t.UserID, &t.FamilyID, &t.TokenHash, &t.ExpiresAt, &t.UsedAt, &t.RevokedAt, &t.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTokenInvalid
		}
		return nil, fmt.Errorf("scan refresh token: %w", err)
	}
	return &t, nil
}

func (r *RefreshTokenRepo) Update(ctx context.Context, projectID string, token *domain.RefreshToken) error {
	pool, err := r.conns.Pool(ctx, projectID)
	if err != nil {
		return err
	}

	tag, err := pool.Exec(ctx, `
		UPDATE refresh_tokens SET used_at = $1, revoked_at = $2 WHERE id = $3
	`, token.UsedAt, token.RevokedAt, token.ID)
	if err != nil {
		return fmt.Errorf("update refresh token: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrTokenInvalid
	}
	return nil
}

// RevokeAllByUser thu hồi mọi refresh token còn sống (chưa revoked) của
// user bằng một câu UPDATE duy nhất — không SELECT rồi loop từng token, vì
// đây chính là hành động khẩn cấp khi phát hiện reuse, cần nhanh và không
// có khoảng hở giữa đọc và ghi.
func (r *RefreshTokenRepo) RevokeAllByUser(ctx context.Context, projectID, userID string) error {
	pool, err := r.conns.Pool(ctx, projectID)
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx, `
		UPDATE refresh_tokens SET revoked_at = now() WHERE user_id = $1 AND revoked_at IS NULL
	`, userID)
	if err != nil {
		return fmt.Errorf("revoke refresh tokens for user: %w", err)
	}
	return nil
}
