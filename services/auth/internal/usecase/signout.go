package usecase

import (
	"context"
	"time"

	"backify/services/auth/internal/domain"
	"backify/services/auth/internal/port"
)

// SignOutInput tham số đăng xuất.
type SignOutInput struct {
	ProjectID    string
	AccessToken  string
	RefreshToken string
}

// SignOut usecase thu hồi refresh token (bắt buộc) và blacklist access token
// (best-effort — lỗi verify access token không chặn signout).
type SignOut struct {
	refreshTokens port.RefreshTokenRepository
	verifier      port.TokenVerifier
	blacklist     port.TokenBlacklist
}

// NewSignOut inject dependencies qua constructor.
func NewSignOut(
	refreshTokens port.RefreshTokenRepository,
	verifier port.TokenVerifier,
	blacklist port.TokenBlacklist,
) *SignOut {
	return &SignOut{
		refreshTokens: refreshTokens,
		verifier:      verifier,
		blacklist:     blacklist,
	}
}

// Execute thu hồi refresh token rồi (nếu verify được) blacklist access token.
func (uc *SignOut) Execute(ctx context.Context, in SignOutInput) error {
	now := time.Now().UTC()

	// 1. Revoke refresh token — hành động bắt buộc, lỗi propagate.
	hash := domain.HashToken(in.RefreshToken)
	token, err := uc.refreshTokens.GetByTokenHash(ctx, in.ProjectID, hash)
	if err != nil {
		return err
	}
	token.Revoke(now)
	if err := uc.refreshTokens.Update(ctx, in.ProjectID, token); err != nil {
		return err
	}

	// 2. Verify access token — lỗi (hết hạn/sai định dạng) bỏ qua, trả nil.
	claims, err := uc.verifier.Verify(in.AccessToken)
	if err != nil {
		return nil
	}

	// 3. Kiểm tra project + blacklist JTI với TTL còn lại tới exp.
	if !claims.BelongsToProject(in.ProjectID) {
		return nil
	}
	ttl := time.Until(time.Unix(claims.Exp, 0))
	_ = uc.blacklist.Add(ctx, claims.JTI, ttl)

	return nil
}
