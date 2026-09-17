package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
)

const (
	// AccessTokenTTL là thời hạn JWT access token — chốt 1 giờ theo BACKIFY.md §7.9.
	AccessTokenTTL = time.Hour
	// RefreshTokenBytes là số byte ngẫu nhiên của refresh token thô trước hex-encode.
	RefreshTokenBytes  = 32
	RefreshDuration7d  = 7 * 24 * time.Hour
	RefreshDuration30d = 30 * 24 * time.Hour
)

// RefreshDuration là thời hạn refresh token user chọn khi signup/signin —
// chỉ nhận 7 ngày hoặc 30 ngày (không phải time.Duration bất kỳ) để khớp
// quyết định đã chốt trong BACKIFY.md, ngăn caller truyền giá trị tùy ý.
type RefreshDuration time.Duration

// Valid báo giá trị có phải 7 ngày hoặc 30 ngày hay không.
func (d RefreshDuration) Valid() bool {
	switch time.Duration(d) {
	case RefreshDuration7d, RefreshDuration30d:
		return true
	default:
		return false
	}
}

// ParseRefreshDuration đọc refreshDuration từ request ("7d", "30d" và vài
// alias) thành RefreshDuration; trả ErrInvalidInput cho giá trị khác để
// handler không phải tự viết lại danh sách alias hợp lệ.
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

// AccessClaims là payload JWT access token: {sub, pid, email, jti, iat, exp}
// theo BACKIFY.md §7.9. Không có trường Algorithm/Signature — ký/verify HS256
// là việc của package jwt (Bước 5), domain chỉ mô tả dữ liệu claims.
type AccessClaims struct {
	Sub   string
	PID   string
	Email string
	JTI   string
	Iat   int64
	Exp   int64
}

// NewAccessClaims tạo claims với JTI ngẫu nhiên và Exp = now + AccessTokenTTL.
// JTI cần thiết để blacklist từng token khi signout (Bước 6), không thể suy
// ra từ sub+iat vì user có thể signin nhiều phiên cùng lúc.
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

// Expired báo token đã hết hạn tại thời điểm now hay chưa.
func (c *AccessClaims) Expired(now time.Time) bool {
	return now.Unix() >= c.Exp
}

// BelongsToProject báo claims có thuộc đúng projectID hay không — dùng ở
// VerifyToken (gRPC) để Runtime không verify nhầm token của project khác
// dù chữ ký JWT hợp lệ (secret dùng chung cho toàn Auth Service).
func (c *AccessClaims) BelongsToProject(projectID string) bool {
	return c.PID == projectID
}

// HashToken băm SHA-256 một token thô (refresh token / reset token) để lưu
// vào DB — không bao giờ lưu token thật, chỉ lưu hash (BACKIFY.md §7.9).
func HashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// RefreshToken lưu hash của refresh token cùng trạng thái sống/dùng/thu hồi.
// FamilyID giữ nguyên qua các lần Rotate để token reuse detection (Bước 6)
// có thể thu hồi toàn bộ chuỗi token khi phát hiện một token đã dùng bị dùng lại.
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

// NewRefreshToken tạo refresh token mới (family mới) và trả về token thô
// (raw) đúng một lần — chỉ TokenHash được lưu vào DB, raw chỉ tồn tại để
// trả cho client trong response signup/signin.
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

// Rotate tạo token kế tiếp cùng FamilyID (để CanRedeem/reuse detection nhận
// diện được chuỗi token) và trả token thô mới; token cũ không tự đổi trạng
// thái ở đây — caller (usecase refresh_token) phải tự MarkUsed token cũ.
// Trả ErrInvalidInput nếu duration không hợp lệ, không tự ý mặc định 7 ngày —
// một duration sai lặng lẽ đổi thành 7d sẽ che mất lỗi gọi sai ở tầng trên.
func (t *RefreshToken) Rotate(duration RefreshDuration) (next *RefreshToken, raw string, err error) {
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

// MarkUsed đánh dấu token đã được dùng để redeem (rotate) tại thời điểm at.
func (t *RefreshToken) MarkUsed(at time.Time) {
	t.UsedAt = &at
}

// Revoke đánh dấu token bị thu hồi tại thời điểm at (logout hoặc reuse detection).
func (t *RefreshToken) Revoke(at time.Time) {
	t.RevokedAt = &at
}

// IsExpired báo token đã quá ExpiresAt tại thời điểm now hay chưa.
func (t *RefreshToken) IsExpired(now time.Time) bool {
	return !now.Before(t.ExpiresAt)
}

// IsRevoked báo token đã bị thu hồi hay chưa.
func (t *RefreshToken) IsRevoked() bool {
	return t.RevokedAt != nil
}

// IsUsed báo token đã được redeem (rotate) trước đó hay chưa.
func (t *RefreshToken) IsUsed() bool {
	return t.UsedAt != nil
}

// CanRedeem kiểm tra token còn dùng được để rotate không, theo đúng thứ tự
// ưu tiên: revoked trước (đã bị thu hồi chủ động, ví dụ do logout hoặc do
// chính reuse detection trước đó) rồi mới tới expired, rồi mới tới reuse —
// một token IsUsed=true bị gọi lại là dấu hiệu reuse (đã rotate rồi mà còn
// bị dùng), usecase refresh_token phải revoke toàn bộ family khi gặp lỗi này.
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
