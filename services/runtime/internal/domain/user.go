package domain

import "time"

type User struct {
	ID           string
	Email        string
	PasswordHash string
	FullName     string
	Phone        string
	Role         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Extra        map[string]any
}

type TokenPair struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int64  `json:"expiresIn"`
}

type AuthClaims struct {
	UserID    string
	ProjectID string
	Role      string
	TokenType string
}
