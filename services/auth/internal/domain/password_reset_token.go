package domain

import (
	"encoding/hex"
	"time"
)

// PasswordResetTokenTTL là thời hạn token reset password — chốt 1 giờ theo
// BACKIFY.md §7.9 (ngắn hơn hẳn refresh token vì đây là thao tác nhạy cảm,
// rò rỉ token trong 1 giờ ít rủi ro hơn 7-30 ngày).
const PasswordResetTokenTTL = time.Hour

// PasswordResetToken lưu hash của token reset password cùng trạng thái
// dùng/hết hạn — không lưu token thật, chỉ lưu TokenHash (giống RefreshToken,
// theo BACKIFY.md §7.9). Không có FamilyID vì reset password không rotate:
// mỗi lần forgot-password tạo một token độc lập, dùng một lần rồi UsedAt.
type PasswordResetToken struct {
	ID        string
	ProjectID string
	UserID    string
	TokenHash string
	ExpiresAt time.Time
	UsedAt    *time.Time
	CreatedAt time.Time
}

// NewPasswordResetToken tạo token mới và trả về token thô (raw) đúng một
// lần — chỉ TokenHash được lưu vào DB, raw chỉ tồn tại để gửi qua email
// trong usecase forgot_password (Bước 6).
func NewPasswordResetToken(projectID, userID string) (token *PasswordResetToken, raw string, err error) {
	if projectID == "" || userID == "" {
		return nil, "", ErrInvalidInput.WithDetails(map[string]interface{}{
			"field":  "password_reset_token",
			"reason": "project_id and user_id are required",
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

	return &PasswordResetToken{
		ID:        id,
		ProjectID: projectID,
		UserID:    userID,
		TokenHash: HashToken(raw),
		ExpiresAt: now.Add(PasswordResetTokenTTL),
		CreatedAt: now,
	}, raw, nil
}

// MarkUsed đánh dấu token đã được dùng để đổi password tại thời điểm at.
func (t *PasswordResetToken) MarkUsed(at time.Time) {
	t.UsedAt = &at
}

// IsExpired báo token đã quá ExpiresAt tại thời điểm now hay chưa.
func (t *PasswordResetToken) IsExpired(now time.Time) bool {
	return !now.Before(t.ExpiresAt)
}

// IsUsed báo token đã được dùng trước đó hay chưa.
func (t *PasswordResetToken) IsUsed() bool {
	return t.UsedAt != nil
}

// CanRedeem kiểm tra token còn dùng để reset password được không. Không có
// nhánh "reuse detection" như RefreshToken — reset password không rotate,
// token đã dùng chỉ đơn giản là không dùng lại được nữa, không cần thu hồi
// dây chuyền vì không có "family" nào để thu hồi.
func (t *PasswordResetToken) CanRedeem(now time.Time) error {
	if t.IsExpired(now) {
		return ErrTokenExpired
	}
	if t.IsUsed() {
		return ErrTokenInvalid
	}
	return nil
}
