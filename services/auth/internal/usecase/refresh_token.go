package usecase

import (
	"context"
	"errors"
	"time"

	"backify/services/auth/internal/domain"
	"backify/services/auth/internal/port"
)

// RefreshTokenInput chỉ cần project + refresh token thô.
// Duration của token mới kế thừa đúng duration gốc (ExpiresAt - CreatedAt).
type RefreshTokenInput struct {
	ProjectID    string
	RefreshToken string
}

// RefreshTokenOutput cặp token mới sau rotate.
type RefreshTokenOutput struct {
	Tokens AuthTokens
}

// RefreshToken usecase rotate refresh token + phát hành access token mới.
type RefreshToken struct {
	users         port.UserRepository
	refreshTokens port.RefreshTokenRepository
	issuer        port.TokenIssuer
}

// NewRefreshToken inject dependencies qua constructor.
func NewRefreshToken(
	users port.UserRepository,
	refreshTokens port.RefreshTokenRepository,
	issuer port.TokenIssuer,
) *RefreshToken {
	return &RefreshToken{
		users:         users,
		refreshTokens: refreshTokens,
		issuer:        issuer,
	}
}

// Execute redeem refresh token: rotate, mark used, issue access mới.
// Reuse detection → RevokeAllByUser rồi trả ErrTokenReuseDetected.
func (uc *RefreshToken) Execute(ctx context.Context, in RefreshTokenInput) (*RefreshTokenOutput, error) {
	now := time.Now().UTC()

	hash := domain.HashToken(in.RefreshToken)
	token, err := uc.refreshTokens.GetByTokenHash(ctx, in.ProjectID, hash)
	if err != nil {
		return nil, err
	}

	if err := token.CanRedeem(now); err != nil {
		if errors.Is(err, domain.ErrTokenReuseDetected) || isDomainCode(err, domain.CodeTokenReuseDetected) {
			_ = uc.refreshTokens.RevokeAllByUser(ctx, in.ProjectID, token.UserID)
		}
		return nil, err
	}

	user, err := uc.users.GetByID(ctx, in.ProjectID, token.UserID)
	if err != nil {
		return nil, err
	}

	// Kế thừa duration gốc từ token cũ.
	originalDuration := domain.RefreshDuration(token.ExpiresAt.Sub(token.CreatedAt))
	// Làm tròn về 7d/30d hợp lệ nếu drift nhỏ; nếu không hợp lệ thì dùng 7d
	// gần nhất — nhưng domain.Rotate yêu cầu Valid(), nên chuẩn hóa.
	originalDuration = normalizeRefreshDuration(originalDuration)

	next, raw, err := token.Rotate(originalDuration)
	if err != nil {
		return nil, err
	}

	token.MarkUsed(now)
	if err := uc.refreshTokens.Update(ctx, in.ProjectID, token); err != nil {
		return nil, err
	}
	if err := uc.refreshTokens.Create(ctx, in.ProjectID, next); err != nil {
		return nil, err
	}

	claims, err := domain.NewAccessClaims(user.ID, in.ProjectID, user.Email)
	if err != nil {
		return nil, err
	}
	accessToken, err := uc.issuer.Issue(claims)
	if err != nil {
		return nil, err
	}

	return &RefreshTokenOutput{
		Tokens: AuthTokens{
			AccessToken:  accessToken,
			RefreshToken: raw,
			ExpiresIn:    int64(domain.AccessTokenTTL / time.Second),
		},
	}, nil
}

// normalizeRefreshDuration map duration gần đúng về 7d hoặc 30d hợp lệ.
func normalizeRefreshDuration(d domain.RefreshDuration) domain.RefreshDuration {
	if d.Valid() {
		return d
	}
	sec := time.Duration(d)
	// Ngưỡng giữa 7d và 30d: nếu >= 15 ngày coi là 30d, ngược lại 7d.
	if sec >= 15*24*time.Hour {
		return domain.RefreshDuration(domain.RefreshDuration30d)
	}
	return domain.RefreshDuration(domain.RefreshDuration7d)
}
