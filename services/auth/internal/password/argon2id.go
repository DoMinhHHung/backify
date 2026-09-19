// Package password băm và so khớp mật khẩu bằng argon2id — tách khỏi domain
// vì domain chỉ import stdlib; golang.org/x/crypto/argon2 là dependency
// ngoài nên phải nằm ở tầng infra, domain chỉ giữ PasswordHash string đã
// băm sẵn (xem domain.User.SetPasswordHash).
package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Tham số argon2id chốt theo BACKIFY.md §7.9: memory 64MB, iter 3, parallel 4.
const (
	memoryKiB   = 64 * 1024 // argon2.IDKey nhận memory theo KiB, 64*1024 = 64MB
	iterations  = 3
	parallelism = 4
	saltLength  = 16
	keyLength   = 32
)

// ErrHashMalformed báo encodedHash không đúng format Hash tạo ra — xảy ra
// nếu dữ liệu trong cột password_hash bị hỏng hoặc không phải do Hash sinh ra.
var ErrHashMalformed = errors.New("password hash is malformed")

// Hash băm plain bằng argon2id với salt ngẫu nhiên, trả về chuỗi tự mô tả
// tham số theo format PHC ($argon2id$v=19$m=...,t=...,p=...$salt$hash) —
// nhúng tham số vào hash để sau này đổi memory/iter/parallel vẫn verify
// được hash cũ đã lưu, không cần migrate dữ liệu hàng loạt.
func Hash(plain string) (string, error) {
	salt := make([]byte, saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(plain), salt, iterations, memoryKiB, parallelism, keyLength)

	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		memoryKiB, iterations, parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	), nil
}

// Verify so plain với encodedHash (tạo bởi Hash), dùng
// subtle.ConstantTimeCompare thay vì bytes.Equal để thời gian so sánh không
// phụ thuộc vào việc hash đúng hay sai tới byte nào — tránh timing attack dò
// dần password đúng qua độ trễ phản hồi.
func Verify(plain, encodedHash string) (bool, error) {
	m, t, p, salt, hash, err := decode(encodedHash)
	if err != nil {
		return false, err
	}

	candidate := argon2.IDKey([]byte(plain), salt, t, m, p, uint32(len(hash)))
	return subtle.ConstantTimeCompare(hash, candidate) == 1, nil
}

func decode(encodedHash string) (memory, iterations uint32, parallelism uint8, salt, hash []byte, err error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 || parts[0] != "" || parts[1] != "argon2id" {
		return 0, 0, 0, nil, nil, ErrHashMalformed
	}

	var version int
	if _, scanErr := fmt.Sscanf(parts[2], "v=%d", &version); scanErr != nil || version != argon2.Version {
		return 0, 0, 0, nil, nil, ErrHashMalformed
	}

	if _, scanErr := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism); scanErr != nil {
		return 0, 0, 0, nil, nil, ErrHashMalformed
	}

	salt, err = base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return 0, 0, 0, nil, nil, ErrHashMalformed
	}
	hash, err = base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return 0, 0, 0, nil, nil, ErrHashMalformed
	}

	return memory, iterations, parallelism, salt, hash, nil
}

// Hasher là adapter rỗng để hàm cấp package Hash/Verify thỏa port.PasswordHasher
// (Bước 6) — hàm cấp package không tự thỏa interface được.
type Hasher struct{}

func (Hasher) Hash(password string) (string, error) {
	return Hash(password)
}

func (Hasher) Verify(password, hash string) (bool, error) {
	return Verify(password, hash)
}
