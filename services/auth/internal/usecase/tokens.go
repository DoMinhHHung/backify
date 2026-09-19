package usecase

import (
	"context"
	"time"

	"backify/services/auth/internal/domain"
	"backify/services/auth/internal/port"
)

// AuthTokens là cặp access + refresh token trả về client sau signup/signin/refresh.
type AuthTokens struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64 // giây còn lại của access token (domain.AccessTokenTTL)
}

// issueTokens tạo access JWT + refresh token mới, lưu refresh vào repo.
// Dùng chung bởi SignUp, SignIn, RefreshToken.
func issueTokens(
	ctx context.Context,
	projectID, userID, email string,
	duration domain.RefreshDuration,
	issuer port.TokenIssuer,
	refreshTokens port.RefreshTokenRepository,
) (AuthTokens, error) {
	claims, err := domain.NewAccessClaims(userID, projectID, email)
	if err != nil {
		return AuthTokens{}, err
	}
	accessToken, err := issuer.Issue(claims)
	if err != nil {
		return AuthTokens{}, err
	}

	refresh, raw, err := domain.NewRefreshToken(projectID, userID, duration)
	if err != nil {
		return AuthTokens{}, err
	}
	if err := refreshTokens.Create(ctx, projectID, refresh); err != nil {
		return AuthTokens{}, err
	}

	return AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: raw,
		ExpiresIn:    int64(domain.AccessTokenTTL / time.Second),
	}, nil
}
