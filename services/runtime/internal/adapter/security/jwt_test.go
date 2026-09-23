package security

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"

	"github.com/DoMinhHHung/backify/services/runtime/internal/domain"
)

func TestJWTServiceIssueAndParseBothTokenTypes(t *testing.T) {
	service := NewJWTService("test-secret", 15*time.Minute, 24*time.Hour)

	pair, err := service.Issue("user-1", "project-1", "admin")

	require.NoError(t, err)
	require.NotEmpty(t, pair.AccessToken)
	require.NotEmpty(t, pair.RefreshToken)
	require.NotEqual(t, pair.AccessToken, pair.RefreshToken)
	require.Equal(t, int64(900), pair.ExpiresIn)

	access, err := service.ParseAccess(pair.AccessToken)
	require.NoError(t, err)
	require.Equal(t, &domain.AuthClaims{
		UserID: "user-1", ProjectID: "project-1", Role: "admin", TokenType: "access",
	}, access)

	refresh, err := service.ParseRefresh(pair.RefreshToken)
	require.NoError(t, err)
	require.Equal(t, "refresh", refresh.TokenType)
}

func TestJWTServiceRejectsWrongTypeSecretAndMalformedToken(t *testing.T) {
	service := NewJWTService("test-secret", time.Minute, time.Hour)
	pair, err := service.Issue("user-1", "project-1", "user")
	require.NoError(t, err)

	_, err = service.ParseAccess(pair.RefreshToken)
	requireUnauthorized(t, err)

	_, err = NewJWTService("different-secret", time.Minute, time.Hour).ParseAccess(pair.AccessToken)
	requireUnauthorized(t, err)

	_, err = service.ParseAccess("not-a-token")
	requireUnauthorized(t, err)
}

func TestJWTServiceRejectsExpiredToken(t *testing.T) {
	service := NewJWTService("test-secret", -time.Second, time.Hour)
	pair, err := service.Issue("user-1", "project-1", "user")
	require.NoError(t, err)

	_, err = service.ParseAccess(pair.AccessToken)

	requireUnauthorized(t, err)
}

func TestJWTServiceRejectsUnexpectedSigningMethod(t *testing.T) {
	service := NewJWTService("test-secret", time.Minute, time.Hour)
	token := jwt.NewWithClaims(jwt.SigningMethodHS384, claims{
		UserID: "user-1", ProjectID: "project-1", Role: "user", TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute))},
	})
	signed, err := token.SignedString([]byte("test-secret"))
	require.NoError(t, err)

	_, err = service.ParseAccess(signed)

	requireUnauthorized(t, err)
}

func requireUnauthorized(t *testing.T, err error) {
	t.Helper()
	var domainErr *domain.DomainError
	require.ErrorAs(t, err, &domainErr)
	require.Equal(t, "unauthorized", domainErr.Code)
}
