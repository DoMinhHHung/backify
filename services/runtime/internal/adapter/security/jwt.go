package security

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/DoMinhHHung/backify/services/runtime/internal/domain"
)

type JWTService struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewJWTService(secret string, accessTTL, refreshTTL time.Duration) *JWTService {
	return &JWTService{
		secret:     []byte(secret),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

type claims struct {
	UserID    string `json:"uid"`
	ProjectID string `json:"pid"`
	Role      string `json:"role"`
	TokenType string `json:"typ"`
	jwt.RegisteredClaims
}

// Issue phát hành cặp token HS256 tại cùng một thời điểm; ExpiresIn là thời hạn
// access token tính bằng giây.
func (s *JWTService) Issue(userID, projectID, role string) (*domain.TokenPair, error) {
	now := time.Now()

	access, err := s.sign(userID, projectID, role, "access", now, s.accessTTL)
	if err != nil {
		return nil, err
	}
	refresh, err := s.sign(userID, projectID, role, "refresh", now, s.refreshTTL)
	if err != nil {
		return nil, err
	}

	return &domain.TokenPair{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    int64(s.accessTTL.Seconds()),
	}, nil
}

// ParseAccess xác thực chữ ký, hạn dùng và loại access; mọi lỗi được chuyển thành ErrUnauthorized.
func (s *JWTService) ParseAccess(token string) (*domain.AuthClaims, error) {
	return s.parse(token, "access")
}

// ParseRefresh xác thực chữ ký, hạn dùng và loại refresh; mọi lỗi được chuyển thành ErrUnauthorized.
func (s *JWTService) ParseRefresh(token string) (*domain.AuthClaims, error) {
	return s.parse(token, "refresh")
}

func (s *JWTService) sign(userID, projectID, role, typ string, now time.Time, ttl time.Duration) (string, error) {
	c := claims{
		UserID:    userID,
		ProjectID: projectID,
		Role:      role,
		TokenType: typ,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   userID,
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	return t.SignedString(s.secret)
}

func (s *JWTService) parse(tokenStr, expectedType string) (*domain.AuthClaims, error) {
	t, err := jwt.ParseWithClaims(tokenStr, &claims{}, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, domain.ErrUnauthorized()
	}
	c, ok := t.Claims.(*claims)
	if !ok || !t.Valid {
		return nil, domain.ErrUnauthorized()
	}
	if c.TokenType != expectedType {
		return nil, domain.ErrUnauthorized()
	}
	return &domain.AuthClaims{
		UserID:    c.UserID,
		ProjectID: c.ProjectID,
		Role:      c.Role,
		TokenType: c.TokenType,
	}, nil
}
