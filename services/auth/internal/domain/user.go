package domain

import (
	"net/mail"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	MinPasswordLength = 8
	MaxPasswordLength = 72
	MaxFullNameLength = 100
	MaxPhoneLength    = 20
)

var phonePattern = regexp.MustCompile(`^[+]?[0-9]{7,15}$`)

// User là end-user của một project (email + password, provider "email" cho
// MVP). Email chỉ unique trong phạm vi ProjectID, không unique toàn hệ thống
// — user của project A và B được phép trùng email (theo BACKIFY.md §7.9).
type User struct {
	ID           string
	ProjectID    string
	Email        string
	PasswordHash string
	FullName     string
	Phone        string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func validateEmail(email string) error {
	trimmed := strings.TrimSpace(strings.ToLower(email))
	if trimmed == "" {
		return ErrInvalidInput.WithDetails(map[string]interface{}{
			"field":  "email",
			"reason": "must not be empty",
		})
	}
	if _, err := mail.ParseAddress(trimmed); err != nil {
		return ErrInvalidInput.WithDetails(map[string]interface{}{
			"field":  "email",
			"reason": "invalid email format",
		})
	}
	return nil
}

// ValidatePassword kiểm tra độ dài thô (8-72 ký tự) trước khi hash bằng
// argon2id — 72 là giới hạn an toàn thực tế cho hầu hết thuật toán hash mật
// khẩu, không phải rule độ mạnh mật khẩu (password strength check là Phase 2,
// nằm ngoài scope MVP).
func ValidatePassword(password string) error {
	length := utf8.RuneCountInString(password)
	if length < MinPasswordLength {
		return ErrInvalidInput.WithDetails(map[string]interface{}{
			"field":  "password",
			"reason": "must be at least 8 characters",
		})
	}
	if length > MaxPasswordLength {
		return ErrInvalidInput.WithDetails(map[string]interface{}{
			"field":  "password",
			"reason": "must be at most 72 characters",
		})
	}
	return nil
}

func validateFullName(name string) error {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return nil
	}
	if utf8.RuneCountInString(trimmed) > MaxFullNameLength {
		return ErrInvalidInput.WithDetails(map[string]interface{}{
			"field":  "fullName",
			"reason": "must be at most 100 characters",
		})
	}
	return nil
}

func validatePhone(phone string) error {
	trimmed := strings.TrimSpace(phone)
	if trimmed == "" {
		return nil
	}
	if utf8.RuneCountInString(trimmed) > MaxPhoneLength {
		return ErrInvalidInput.WithDetails(map[string]interface{}{
			"field":  "phone",
			"reason": "must be at most 20 characters",
		})
	}
	if !phonePattern.MatchString(trimmed) {
		return ErrInvalidInput.WithDetails(map[string]interface{}{
			"field":  "phone",
			"reason": "must be 7-15 digits, optional leading +",
		})
	}
	return nil
}

// NewUser tạo User mới sau khi validate email/fullName/phone; PasswordHash
// để trống — caller (usecase signup) hash password riêng bằng argon2id rồi
// gọi SetPasswordHash, vì domain không được import package hash (chỉ stdlib).
// fullName/phone rỗng hợp lệ (chỉ 3 field hệ thống bắt buộc: email, password;
// fullName/phone là optional theo request format trong BACKIFY.md §7.9).
func NewUser(projectID, email, fullName, phone string) (*User, error) {
	if strings.TrimSpace(projectID) == "" {
		return nil, ErrInvalidInput.WithDetails(map[string]interface{}{
			"field":  "project_id",
			"reason": "must not be empty",
		})
	}

	if err := validateEmail(email); err != nil {
		return nil, err
	}
	if err := validateFullName(fullName); err != nil {
		return nil, err
	}
	if err := validatePhone(phone); err != nil {
		return nil, err
	}

	id, err := generateID()
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	return &User{
		ID:        id,
		ProjectID: projectID,
		Email:     strings.TrimSpace(strings.ToLower(email)),
		FullName:  strings.TrimSpace(fullName),
		Phone:     strings.TrimSpace(phone),
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// SetPasswordHash gán hash argon2id đã tính sẵn (domain không tự hash) và
// cập nhật UpdatedAt; trả lỗi nếu hash rỗng để tránh lưu user không có mật khẩu.
func (u *User) SetPasswordHash(hash string) error {
	if hash == "" {
		return ErrInvalidInput.WithDetails(map[string]interface{}{
			"field":  "password_hash",
			"reason": "must not be empty",
		})
	}
	u.PasswordHash = hash
	u.UpdatedAt = time.Now().UTC()
	return nil
}
