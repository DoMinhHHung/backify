package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
)

const (
	AccessTokenTTL      = time.Hour
	RefreshTokenBytes   = 32
	RefreshDuration7d   = 7 * 24 * time.Hour
	RefreshDuration30d  = 30 * 24 * time.Hour
)

type RefreshDuration time.Duration

func (d RefreshDuration) Valid() bool {
	switch time.Duration(d) {
	case RefreshDuration7d, RefreshDuration30d:
		return true
	default:
		return false
	}
}

func ParseRefreshDuration(s string) (RefreshDuration, error) {
	switch s {
	case "7d", "7days", "7":
		return RefreshDuration(RefreshDuration7d), nil
	case "30d", "30days", "30":
		return RefreshDuration(RefreshDuration30d), nil
	default:
		return 0, ErrInvalidInput.WithDetails(map[string]interface{}{
			"field":  "refresh_duration",
			"reason": "must be 7d or 30d",
		})
	}
}

type AccessClaims struct {
	Sub   string
	PID   string
	Email string
	JTI   string
	Iat   int64
	Exp   int64
}

func NewAccessClaims(userID, projectID, email string) (*AccessClaims, error) {
	if userID == "" || projectID == "" {
		return nil, ErrInvalidInput.WithDetails(map[string]interface{}{
			"field":  "claims",
			"reason": "sub and pid are required",
		})
	}
	jti, err := generateID()
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	return &AccessClaims{
		Sub:   userID,
		PID:   projectID,
		Email: email,
		JTI:   jti,
		Iat:   now.Unix(),
		Exp:   now.Add(AccessTokenTTL).Unix(),
	}, nil
}

func (c *AccessClaims) Expired(now time.Time) bool {
	return now.Unix() >= c.Exp
}

func (c *AccessClaims) BelongsToProject(projectID string) bool {
	return c.PID == projectID
}

func HashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

type RefreshToken struct {
	ID        string
	ProjectID string
	UserID    string
	FamilyID  string
	TokenHash string
	ExpiresAt time.Time
	UsedAt    *time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

func NewRefreshToken(projectID, userID string, duration RefreshDuration) (token *RefreshToken, raw string, err error) {
	if projectID == "" || userID == "" {
		return nil, "", ErrInvalidInput.WithDetails(map[string]interface{}{
			"field":  "refresh_token",
			"reason": "project_id and user_id are required",
		})
	}
	if !duration.Valid() {
		return nil, "", ErrInvalidInput.WithDetails(map[string]interface{}{
			"field":  "refresh_duration",
			"reason": "must be 7d or 30d",
		})
	}

	id, err := generateID()
	if err != nil {
		return nil, "", err
	}
	familyID, err := generateID()
	if err != nil {
		return nil, "", err
	}
	rawBytes, err := generateTokenBytes(RefreshTokenBytes)
	if err != nil {
		return nil, "", err
	}
	raw = hex.EncodeToString(rawBytes)
	now := time.Now().UTC()

	return &RefreshToken{
		ID:        id,
		ProjectID: projectID,
		UserID:    userID,
		FamilyID:  familyID,
		TokenHash: HashToken(raw),
		ExpiresAt: now.Add(time.Duration(duration)),
		CreatedAt: now,
	}, raw, nil
}

func (t *RefreshToken) Rotate(duration RefreshDuration) (next *RefreshToken, raw string, err error) {
	if !duration.Valid() {
		duration = RefreshDuration(RefreshDuration7d)
	}
	id, err := generateID()
	if err != nil {
		return nil, "", err
	}
	rawBytes, err := generateTokenBytes(RefreshTokenBytes)
	if err != nil {
		return nil, "", err
	}
	raw = hex.EncodeToString(rawBytes)
	now := time.Now().UTC()
	return &RefreshToken{
		ID:        id,
		ProjectID: t.ProjectID,
		UserID:    t.UserID,
		FamilyID:  t.FamilyID,
		TokenHash: HashToken(raw),
		ExpiresAt: now.Add(time.Duration(duration)),
		CreatedAt: now,
	}, raw, nil
}

func (t *RefreshToken) MarkUsed(at time.Time) {
	t.UsedAt = &at
}

func (t *RefreshToken) Revoke(at time.Time) {
	t.RevokedAt = &at
}

func (t *RefreshToken) IsExpired(now time.Time) bool {
	return !now.Before(t.ExpiresAt)
}

func (t *RefreshToken) IsRevoked() bool {
	return t.RevokedAt != nil
}

func (t *RefreshToken) IsUsed() bool {
	return t.UsedAt != nil
}

func (t *RefreshToken) CanRedeem(now time.Time) error {
	if t.IsRevoked() {
		return ErrTokenRevoked
	}
	if t.IsExpired(now) {
		return ErrTokenExpired
	}
	if t.IsUsed() {
		return ErrTokenReuseDetected
	}
	return nil
}
