package domain

import (
	"crypto/rand"
	"encoding/hex"
)

// generateID sinh ID ngẫu nhiên 16 hex char (8 byte) — đủ để tránh trùng
// trong phạm vi một project mà không cần thêm dependency ngoài stdlib
// (đồng bộ cách control-plane sinh ID entity/field/module).
func generateID() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// generateTokenBytes sinh n byte ngẫu nhiên dùng làm refresh token /
// password reset token thô trước khi hex-encode; tách riêng vì độ dài
// (32 byte, theo BACKIFY.md) khác với generateID.
func generateTokenBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	return b, nil
}
